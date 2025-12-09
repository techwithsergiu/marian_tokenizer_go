// marian_v1/sp_wrapper.cc
#include "sp_wrapper.h"

#include <string>
#include <vector>
#include <cstring>

#include "../deps/sentencepiece/include/sentencepiece_processor.h"

using sentencepiece::SentencePieceProcessor;

// Load a SentencePiece model from the given path.
// Returns a non-null handle on success, or NULL on failure.
sp_handle_t sp_new(const char* model_path) {
    auto* sp = new SentencePieceProcessor();
    auto status = sp->Load(std::string(model_path));
    if (!status.ok()) {
        delete sp;
        return nullptr;
    }
    return reinterpret_cast<sp_handle_t>(sp);
}

// Destroy a previously created SentencePiece handle.
void sp_free(sp_handle_t handle) {
    if (!handle) return;
    auto* sp = reinterpret_cast<SentencePieceProcessor*>(handle);
    delete sp;
}

// Encode UTF-8 text into SentencePiece internal ids.
// Returns:
//   >= 0: number of ids written to out_ids
//   < 0: error code
int sp_encode_as_ids(sp_handle_t handle,
        const char* text,
        int* out_ids,
        int max_ids) {
    if (!handle || !text || !out_ids || max_ids <= 0) return -1;

    auto* sp = reinterpret_cast<SentencePieceProcessor*>(handle);

    std::vector<int> ids;
    auto status = sp->Encode(std::string(text), &ids);
    if (!status.ok()) return -2;

    if ((int)ids.size() > max_ids) return -3;

    for (int i = 0; i < (int)ids.size(); ++i) {
        out_ids[i] = ids[i];
    }
    return (int)ids.size();
}

// Convert a SentencePiece id to its piece string.
// Copies a null-terminated string into out_buf.
// Returns:
//   >= 0: length of the piece in bytes (excluding '\0')
//   < 0: error code
int sp_id_to_piece(
        sp_handle_t handle,
        int id,
        char* out_buf,
        int max_len) {
    if (!handle || !out_buf || max_len <= 0) return -1;
    auto* sp = reinterpret_cast<SentencePieceProcessor*>(handle);

    const std::string piece = sp->IdToPiece(id);
    int n = (int)piece.size();
    if (n + 1 > max_len) return -2;

    std::memcpy(out_buf, piece.c_str(), n + 1); // include \0
    return n;
}

// Decode an array of piece strings into UTF-8 text.
// Returns:
//   >= 0: length of the decoded string in bytes (excluding '\0')
//   < 0: error code
int sp_decode_pieces(
        sp_handle_t handle,
        const char** pieces,
        int len,
        char* out_buf,
        int max_len) {
    if (!handle || !pieces || len <= 0 || !out_buf || max_len <= 0) return -1;

    auto* sp = reinterpret_cast<SentencePieceProcessor*>(handle);

    std::vector<std::string> vec;
    vec.reserve(len);
    for (int i = 0; i < len; ++i) {
        vec.emplace_back(pieces[i]);
    }

    std::string result;
    auto status = sp->Decode(vec, &result);
    if (!status.ok()) return -2;

    if ((int)result.size() + 1 > max_len) return -3;

    std::memcpy(out_buf, result.c_str(), result.size() + 1); // include \0
    return (int)result.size();
}
