//go:build cgo && amd64 && (linux || windows)

package marian_v3

/*
#cgo CXXFLAGS: -std=c++17 -I${SRCDIR}/../deps/marian_tokenizer_core/include
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../deps/marian_tokenizer_core/linux_amd64/lib -lmarian_core -lstdc++
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/../deps/marian_tokenizer_core/windows_amd64/lib -lmarian_core -static-libstdc++ -static-libgcc -Wl,-Bstatic -lwinpthread -Wl,-Bdynamic
#include <stdlib.h>
#include "../deps/marian_tokenizer_core/include/marian_core.h"
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/techwithsergiu/marian_tokenizer_go/marian"
)

// Tokenizer is a Marian-core–backed tokenizer implementation.
// It delegates all logic (SP + vocab + special tokens) to the C++ core.
type Tokenizer struct {
	h           C.marian_tok_t
	padID       int64
	modelMaxLen int
}

// Ensure Tokenizer satisfies the common interface.
var _ marian.Tokenizer = (*Tokenizer)(nil)

// NewTokenizer creates a new Marian-core tokenizer for the given model directory.
// The directory must contain: config.json, vocab.json, source.spm, target.spm
func NewTokenizer(modelDir string) (marian.Tokenizer, error) {
	cDir := C.CString(modelDir)
	defer C.free(unsafe.Pointer(cDir))

	h := C.marian_tok_new(cDir)
	if h == nil {
		return nil, fmt.Errorf("marian_tok_new failed")
	}

	pad := C.marian_tok_get_pad_id(h)
	maxLen := C.marian_tok_get_model_max_length(h)

	return &Tokenizer{
		h:           h,
		padID:       int64(pad),
		modelMaxLen: int(maxLen),
	}, nil
}

// Close releases the underlying native Marian tokenizer handle.
func (t *Tokenizer) Close() {
	if t.h != nil {
		C.marian_tok_free(t.h)
		t.h = nil
	}
}

// encodeInternal is a small helper that calls marian_tok_encode with configurable addEOS.
func (t *Tokenizer) encodeInternal(text string, addEOS bool) ([]int64, error) {
	if t.h == nil {
		return nil, fmt.Errorf("tokenizer closed")
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	maxTokens := t.modelMaxLen
	if maxTokens <= 0 {
		return nil, fmt.Errorf("model_max_length is not positive")
	}

	buf := make([]C.longlong, maxTokens)

	var add C.int
	if addEOS {
		add = 1
	} else {
		add = 0
	}

	n := C.marian_tok_encode(
		t.h,
		cText,
		&buf[0],
		C.int(maxTokens),
		add,
	)
	if n < 0 {
		return nil, fmt.Errorf("marian_tok_encode failed: %d", int(n))
	}

	out := make([]int64, int(n))
	for i := 0; i < int(n); i++ {
		out[i] = int64(buf[i])
	}
	return out, nil
}

// Encode encodes a single source sentence into token IDs.
// If addEOS is true, EOS token is appended.
func (t *Tokenizer) Encode(text string, addEOS bool) ([]int64, error) {
	return t.encodeInternal(text, addEOS)
}

// EncodeBatch encodes a batch of sentences and returns:
//   - inputIDs: shape (batch, maxLen)
//   - attentionMask: shape (batch, maxLen) with 1 for tokens and 0 for padding.
func (t *Tokenizer) EncodeBatch(texts []string) ([][]int64, [][]int64, error) {
	if t.h == nil {
		return nil, nil, fmt.Errorf("tokenizer closed")
	}

	batch := len(texts)
	if batch == 0 {
		return [][]int64{}, [][]int64{}, nil
	}

	// 1) Convert Go strings to C strings.
	cTexts := make([]*C.char, batch)
	for i, s := range texts {
		cTexts[i] = C.CString(s)
	}
	defer func() {
		for _, p := range cTexts {
			if p != nil {
				C.free(unsafe.Pointer(p))
			}
		}
	}()

	maxLen := t.modelMaxLen
	if maxLen <= 0 {
		return nil, nil, fmt.Errorf("model_max_length is not positive")
	}

	flatIDs := make([]C.longlong, batch*maxLen)
	seqLens := make([]C.int, batch)
	flatMask := make([]C.int, batch*maxLen)

	// 2) Batch encode in C++ (always add EOS for batch encode).
	maxUsed := C.marian_tok_encode_batch(
		t.h,
		(**C.char)(unsafe.Pointer(&cTexts[0])),
		C.int(batch),
		C.int(maxLen),
		&flatIDs[0],
		&seqLens[0],
		1, // add_eos = 1
	)
	if maxUsed < 0 {
		return nil, nil, fmt.Errorf("marian_tok_encode_batch failed: %d", int(maxUsed))
	}
	usedLen := int(maxUsed)

	// 3) Build attention mask up to maxLen, then we’ll slice to usedLen.
	rc := C.marian_tok_build_attention_mask(
		&seqLens[0],
		C.int(batch),
		C.int(maxLen),
		&flatMask[0],
	)
	if rc < 0 {
		return nil, nil, fmt.Errorf("marian_tok_build_attention_mask failed: %d", int(rc))
	}

	// 4) Reshape into [batch][usedLen].
	inputIDs := make([][]int64, batch)
	attn := make([][]int64, batch)

	for b := 0; b < batch; b++ {
		inputIDs[b] = make([]int64, usedLen)
		attn[b] = make([]int64, usedLen)

		rowOffset := b * maxLen
		for j := 0; j < usedLen; j++ {
			inputIDs[b][j] = int64(flatIDs[rowOffset+j])
			attn[b][j] = int64(flatMask[rowOffset+j])
		}
	}

	return inputIDs, attn, nil
}

// Decode converts token IDs back to a target sentence.
// If skipSpecial is true, EOS / PAD / UNK are removed before decoding.
func (t *Tokenizer) Decode(ids []int64, skipSpecial bool) (string, error) {
	if t.h == nil {
		return "", fmt.Errorf("tokenizer closed")
	}

	if len(ids) == 0 {
		return "", nil
	}

	cids := make([]C.longlong, len(ids))
	for i, v := range ids {
		cids[i] = C.longlong(v)
	}

	const maxText = 4096
	buf := make([]C.char, maxText)

	var skip C.int
	if skipSpecial {
		skip = 1
	} else {
		skip = 0
	}

	n := C.marian_tok_decode(
		t.h,
		&cids[0],
		C.int(len(cids)),
		skip,
		&buf[0],
		C.int(maxText),
	)
	if n < 0 {
		return "", fmt.Errorf("marian_tok_decode failed: %d", int(n))
	}

	return C.GoString(&buf[0]), nil
}
