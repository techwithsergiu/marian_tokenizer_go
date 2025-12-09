// marian_v1/sp_wrapper.h
#pragma once

#ifdef __cplusplus
extern "C" {
#endif

typedef void* sp_handle_t;

// Load a SentencePiece model from the given path.
// Returns a non-null handle on success, or NULL on failure.
sp_handle_t sp_new(const char* model_path);

// Destroy a previously created SentencePiece handle.
void sp_free(sp_handle_t handle);

// Encode UTF-8 text into SentencePiece internal ids.
// Returns:
//   >= 0: number of ids written to out_ids
//   < 0: error code
int sp_encode_as_ids(
        sp_handle_t handle,
        const char* text,
        int* out_ids,
        int max_ids);

// Convert a SentencePiece id to its piece string.
// Copies a null-terminated string into out_buf.
// Returns:
//   >= 0: length of the piece in bytes (excluding '\0')
//   < 0: error code
int sp_id_to_piece(
        sp_handle_t handle,
        int id,
        char* out_buf,
        int max_len);

// Decode an array of piece strings into UTF-8 text.
// Returns:
//   >= 0: length of the decoded string in bytes (excluding '\0')
//   < 0: error code
int sp_decode_pieces(
        sp_handle_t handle,
        const char** pieces,
        int len,
        char* out_buf,
        int max_len);

#ifdef __cplusplus
}
#endif
