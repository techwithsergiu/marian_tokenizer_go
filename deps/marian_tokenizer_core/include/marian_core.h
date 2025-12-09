// src/marian_core.h
#pragma once

#include <stdint.h>

#if defined(_WIN32) || defined(_WIN64)
  #if defined(MARIAN_CORE_BUILD)
    #define MARIAN_API __declspec(dllexport)
  #else
    #define MARIAN_API __declspec(dllimport)
  #endif
#else
  #if defined(MARIAN_CORE_BUILD)
    #define MARIAN_API __attribute__((visibility("default")))
  #else
    #define MARIAN_API
  #endif
#endif

#ifdef __cplusplus
extern "C" {
#endif

typedef void* marian_tok_t;

// Create a Marian tokenizer instance from a model directory.
//
// The directory must contain:
//   - config.json
//   - vocab.json
//   - source.spm
//   - target.spm
MARIAN_API marian_tok_t marian_tok_new(const char* model_dir);

// Destroy a previously created Marian tokenizer instance.
MARIAN_API void marian_tok_free(marian_tok_t handle);

// Get the PAD token id from the loaded configuration.
MARIAN_API long long marian_tok_get_pad_id(marian_tok_t handle);

// Get the model_max_length from the loaded configuration.
MARIAN_API long long marian_tok_get_model_max_length(marian_tok_t handle);

// Encode UTF-8 text into Marian token ids.
//
// add_eos: 0 or 1
// Returns:
//   >= 0: number of ids written to out_ids
//   < 0: error code
MARIAN_API int marian_tok_encode(
        marian_tok_t handle,
        const char* text,
        long long* out_ids,  // int64_t-compatible type
        int max_ids,
        int add_eos);

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
MARIAN_API int marian_tok_encode_batch(
        marian_tok_t handle,
        const char** texts,
        int batch_size,
        int max_len,
        long long* out_ids,
        int* out_seq_lens,
        int add_eos);

// Build attention masks from sequence lengths.
//
// seq_lens: size [batch_size]
// out_mask: size [batch_size * max_len], row-major, values 0/1
// Returns:
//   0  on success
//  <0 on error
MARIAN_API int marian_tok_build_attention_mask(
        const int* seq_lens,
        int batch_size,
        int max_len,
        int* out_mask);

// Decode Marian token ids back to UTF-8 text.
//
// skip_special: 0 or 1; if 1, special tokens are removed before decoding.
// Returns:
//   >= 0: length of the decoded string (without '\0')
//   < 0: error code
MARIAN_API int marian_tok_decode(
        marian_tok_t handle,
        const long long* ids,
        int len,
        int skip_special,
        char* out_text,
        int max_text_len);

#ifdef __cplusplus
}
#endif
