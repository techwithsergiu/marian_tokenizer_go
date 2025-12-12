// src/marian_core.cc
#include "marian_core.h"

#include "sentencepiece_processor.h"
#include "json.hpp"

#include <string>
#include <vector>
#include <unordered_map>
#include <unordered_set>
#include <fstream>
#include <sstream>
#include <cstdlib>
#include <cstring>

using json = nlohmann::json;

using sentencepiece::SentencePieceProcessor;

struct MarianCoreConfig {
    int vocab_size = 0;
    int decoder_vocab_size = 0;
    long long eos_id = 0;
    long long bos_id = 0;
    long long pad_id = 0;
    long long decoder_start_id = 0;
    int max_length = 512;
    int model_max_length = 512;
    std::vector<std::vector<long long>> bad_words_ids;
};

struct MarianCore {
    SentencePieceProcessor sp_source;
    SentencePieceProcessor sp_target;

    MarianCoreConfig cfg;
    std::string cfg_json;

    std::unordered_map<std::string, long long> token2id;
    std::vector<std::string> id2token;

    long long unk_id = 1;
    std::unordered_set<long long> special_ids;
};

static bool load_file(const std::string& path, std::string& out) {
    std::ifstream in(path);
    if (!in.is_open()) return false;
    std::ostringstream ss;
    ss << in.rdbuf();
    out = ss.str();
    return true;
}

static bool parse_config(const std::string& json_str, MarianCoreConfig& cfg) {
    try {
        json j = json::parse(json_str);

        cfg.vocab_size        = j.at("vocab_size").get<int>();
        cfg.decoder_vocab_size= j.value("decoder_vocab_size", cfg.vocab_size);

        cfg.eos_id            = j.at("eos_token_id").get<long long>();
        cfg.bos_id            = j.value("bos_token_id", cfg.eos_id);

        cfg.pad_id            = j.at("pad_token_id").get<long long>();
        cfg.decoder_start_id  = j.at("decoder_start_token_id").get<long long>();

        cfg.max_length        = j.value("max_length", 512);
        cfg.model_max_length  = j.value("model_max_length", cfg.max_length);

        cfg.bad_words_ids.clear();
        if (j.contains("bad_words_ids")) {
            for (auto& seq : j["bad_words_ids"]) {
                std::vector<long long> v;
                for (auto& x : seq) {
                    v.push_back(x.get<long long>());
                }
                cfg.bad_words_ids.push_back(std::move(v));
            }
        }

        return true;
    } catch (const std::exception& e) {
        return false;
    } catch (...) {
        return false;
    }
}

static bool parse_vocab(
        const std::string& json_str,
        std::unordered_map<std::string, long long>& token2id,
        std::vector<std::string>& id2token) {
    try {
        json j = json::parse(json_str);

        token2id.clear();
        long long max_id = -1;

        for (auto it = j.begin(); it != j.end(); ++it) {
            const std::string tok = it.key();
            long long id = it.value().get<long long>();
            token2id[tok] = id;
            if (id > max_id) max_id = id;
        }

        id2token.assign(max_id + 1, "");
        for (const auto& kv : token2id) {
            const std::string& tok = kv.first;
            long long id = kv.second;
            if (id >= 0 && id < (long long)id2token.size()) {
                id2token[id] = tok;
            }
        }

        return true;
    } catch (const std::exception& e) {
        return false;
    } catch (...) {
        return false;
    }
}

extern "C" {

// Create a Marian tokenizer instance from a model directory.
//
// The directory must contain:
//   - config.json
//   - vocab.json
//   - source.spm
//   - target.spm
marian_tok_t marian_tok_new(const char* model_dir_cstr) {
    if (!model_dir_cstr) return nullptr;

    auto* core = new MarianCore();

    std::string model_dir(model_dir_cstr);

    // 1) config.json
    std::string cfg_str;
    if (!load_file(model_dir + "/config.json", cfg_str)) {
        delete core;
        return nullptr;
    }
    if (!parse_config(cfg_str, core->cfg)) {
        delete core;
        return nullptr;
    }
    core->cfg_json = cfg_str;

    // 2) vocab.json
    std::string vocab_str;
    if (!load_file(model_dir + "/vocab.json", vocab_str)) {
        delete core;
        return nullptr;
    }
    if (!parse_vocab(vocab_str, core->token2id, core->id2token)) {
        delete core;
        return nullptr;
    }

    // 3) sentencepiece models
    auto status_src = core->sp_source.Load(model_dir + "/source.spm");
    if (!status_src.ok()) {
        delete core;
        return nullptr;
    }
    auto status_tgt = core->sp_target.Load(model_dir + "/target.spm");
    if (!status_tgt.ok()) {
        delete core;
        return nullptr;
    }

    // 4) special tokens
    auto it_unk = core->token2id.find("<unk>");
    core->unk_id = (it_unk != core->token2id.end()) ? it_unk->second : 1;

    core->special_ids.clear();
    core->special_ids.insert(core->cfg.eos_id);
    core->special_ids.insert(core->cfg.pad_id);
    core->special_ids.insert(core->unk_id);

    return reinterpret_cast<marian_tok_t>(core);
}

// Destroy a previously created Marian tokenizer instance.
void marian_tok_free(marian_tok_t handle) {
    if (!handle) return;
    auto* core = reinterpret_cast<MarianCore*>(handle);
    delete core;
}

// Free buffers returned by API (json, strings, etc.)
void marian_tok_free_buffer(void* p) {
    std::free(p);
}

// Get config.json as raw JSON bytes
const char* marian_tok_get_config_json(marian_tok_t handle, size_t* out_len) {
    if (out_len) *out_len = 0;
    if (!handle) return nullptr;

    auto* core = reinterpret_cast<MarianCore*>(handle);
    const std::string& s = core->cfg_json;
    if (s.empty()) return nullptr;

    char* buf = (char*)std::malloc(s.size());
    if (!buf) return nullptr;

    std::memcpy(buf, s.data(), s.size());
    if (out_len) *out_len = s.size();
    return buf;
}

// Encode UTF-8 text into Marian token ids.
//
// add_eos: 0 or 1
// Returns:
//   >= 0: number of ids written to out_ids
//   < 0: error code
int marian_tok_encode(
        marian_tok_t handle,
        const char* text,
        long long* out_ids,
        int max_ids,
        int add_eos) {
    if (!handle || !text || !out_ids || max_ids <= 0) return -1;
    auto* core = reinterpret_cast<MarianCore*>(handle);

    std::vector<std::string> pieces;
    auto status = core->sp_source.Encode(std::string(text), &pieces);
    if (!status.ok()) return -2;

    std::vector<long long> ids;
    ids.reserve(pieces.size() + 1);

    for (const auto& p : pieces) {
        auto it = core->token2id.find(p);
        if (it != core->token2id.end()) {
            ids.push_back(it->second);
        } else {
            ids.push_back(core->unk_id);
        }
    }

    if (add_eos) {
        ids.push_back(core->cfg.eos_id);
    }

    if ((int)ids.size() > core->cfg.model_max_length) {
        ids.resize(core->cfg.model_max_length);
    }

    if ((int)ids.size() > max_ids) {
        return -3; // output buffer is too small
    }

    for (int i = 0; i < (int)ids.size(); ++i) {
        out_ids[i] = ids[i];
    }
    return (int)ids.size();
}

// Batch-encode UTF-8 texts into Marian token ids.
//
// texts:       array of C-string pointers of length batch_size
// max_len:     capacity (stride) for each sequence in out_ids
// out_ids:     size [batch_size * max_len], row-major
// out_seq_lens:size [batch_size], actual sequence length per row
// add_eos:     0 or 1
// Returns:
//   >= 0: maximum sequence length across the batch
//   < 0: error code
int marian_tok_encode_batch(
        marian_tok_t handle,
        const char** texts,
        int batch_size,
        int max_len,
        long long* out_ids,
        int* out_seq_lens,
        int add_eos) {
    if (!handle || !texts || batch_size <= 0 || max_len <= 0 || !out_ids || !out_seq_lens) {
        return -1;
    }
    auto* core = reinterpret_cast<MarianCore*>(handle);

    int global_max_len = 0;

    for (int b = 0; b < batch_size; ++b) {
        const char* t = texts[b];
        if (!t) {
            out_seq_lens[b] = 0;
            continue;
        }

        std::vector<std::string> pieces;
        auto status = core->sp_source.Encode(std::string(t), &pieces);
        if (!status.ok()) return -2;

        std::vector<long long> ids;
        ids.reserve(pieces.size() + 1);

        for (const auto& p : pieces) {
            auto it = core->token2id.find(p);
            if (it != core->token2id.end()) {
                ids.push_back(it->second);
            } else {
                ids.push_back(core->unk_id);
            }
        }

        if (add_eos) {
            ids.push_back(core->cfg.eos_id);
        }

        // truncate by model_max_length
        if ((int)ids.size() > core->cfg.model_max_length) {
            ids.resize(core->cfg.model_max_length);
        }

        int seq_len = (int)ids.size();
        if (seq_len > max_len) {
            // buffer is too small
            return -3;
        }

        out_seq_lens[b] = seq_len;
        if (seq_len > global_max_len) {
            global_max_len = seq_len;
        }

        // fills a strings in out_ids with padding
        int row_offset = b * max_len;
        int j = 0;
        for (; j < seq_len; ++j) {
            out_ids[row_offset + j] = ids[j];
        }
        for (; j < max_len; ++j) {
            out_ids[row_offset + j] = core->cfg.pad_id;
        }
    }

    return global_max_len; // actual maximum sequence length in the batch
}

// Build attention masks from sequence lengths.
//
// seq_lens: size [batch_size]
// out_mask: size [batch_size * max_len], row-major, values 0/1
// Returns:
//   0  on success
//  <0 on error
int marian_tok_build_attention_mask(
        const int* seq_lens,
        int batch_size,
        int max_len,
        int* out_mask) {
    if (!seq_lens || !out_mask || batch_size <= 0 || max_len <= 0) {
        return -1;
    }

    for (int b = 0; b < batch_size; ++b) {
        int len = seq_lens[b];
        if (len < 0) len = 0;
        if (len > max_len) len = max_len;

        int row_offset = b * max_len;
        int j = 0;
        for (; j < len; ++j) {
            out_mask[row_offset + j] = 1;
        }
        for (; j < max_len; ++j) {
            out_mask[row_offset + j] = 0;
        }
    }
    return 0;
}

// Decode Marian token ids back to UTF-8 text.
//
// skip_special: 0 or 1; if 1, special tokens are removed before decoding.
// Returns:
//   >= 0: length of the decoded string (without '\0')
//   < 0: error code
int marian_tok_decode(
        marian_tok_t handle,
        const long long* ids,
        int len,
        int skip_special,
        char* out_text,
        int max_text_len) {
    if (!handle || !ids || len <= 0 || !out_text || max_text_len <= 0) return -1;
    auto* core = reinterpret_cast<MarianCore*>(handle);

    std::vector<std::string> pieces;
    pieces.reserve(len);

    for (int i = 0; i < len; ++i) {
        long long id = ids[i];

        if (skip_special && core->special_ids.count(id) > 0) {
            continue;
        }

        if (id < 0 || (size_t)id >= core->id2token.size() || core->id2token[id].empty()) {
            pieces.emplace_back("<unk>");
        } else {
            pieces.emplace_back(core->id2token[id]);
        }
    }

    if (pieces.empty()) {
        if (max_text_len > 0) out_text[0] = '\0';
        return 0;
    }

    std::string result;
    auto status = core->sp_target.Decode(pieces, &result);
    if (!status.ok()) return -2;

    if ((int)result.size() + 1 > max_text_len) {
        return -3; // output buffer is too small
    }

    std::memcpy(out_text, result.c_str(), result.size() + 1);
    return (int)result.size();
}

} // extern "C"
