//go:build linux && amd64 && cgo

package marian_v1

/*
#cgo CXXFLAGS: -std=c++17 -I${SRCDIR} -I${SRCDIR}/../deps/sentencepiece/include
#cgo LDFLAGS: -L${SRCDIR}/../deps/sentencepiece/linux_amd64/lib/static -lsentencepiece -lstdc++
#include <stdlib.h>
#include "sp_wrapper.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/techwithsergiu/marian_tokenizer_go/marian"
)

type Tokenizer struct {
	spSource C.sp_handle_t
	spTarget C.sp_handle_t

	config   marian.Config
	token2id map[string]int64
	id2token []string

	unkID int64
}

// ensure interface implementation
var _ marian.Tokenizer = (*Tokenizer)(nil)

// loadConfig loads config.json from disk, unmarshals it into Config,
// and normalizes the result to apply default values.
func loadConfig(path string) (marian.Config, error) {
	var cfg marian.Config
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}

	cfg.NormalizeConfig()

	return cfg, nil
}

func loadVocab(path string) (map[string]int64, []string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	raw := map[string]int64{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, err
	}

	var maxID int64
	for _, id := range raw {
		if id > maxID {
			maxID = id
		}
	}
	id2token := make([]string, maxID+1)
	for tok, id := range raw {
		id2token[id] = tok
	}
	return raw, id2token, nil
}

// NewTokenizer creates a SentencePiece-based Marian tokenizer from a model directory
// containing: config.json, source.spm, target.spm, vocab.json.
func NewTokenizer(modelDir string) (marian.Tokenizer, error) {
	modelDir = filepath.Clean(modelDir)

	cfg, err := loadConfig(filepath.Join(modelDir, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	token2id, id2token, err := loadVocab(filepath.Join(modelDir, "vocab.json"))
	if err != nil {
		return nil, fmt.Errorf("load vocab: %w", err)
	}

	// source.spm
	cSrc := C.CString(filepath.Join(modelDir, "source.spm"))
	defer C.free(unsafe.Pointer(cSrc))
	spSrc := C.sp_new(cSrc)
	if spSrc == nil {
		return nil, fmt.Errorf("sp_new(source.spm) failed")
	}

	// target.spm
	cTgt := C.CString(filepath.Join(modelDir, "target.spm"))
	defer C.free(unsafe.Pointer(cTgt))
	spTgt := C.sp_new(cTgt)
	if spTgt == nil {
		C.sp_free(spSrc)
		return nil, fmt.Errorf("sp_new(target.spm) failed")
	}

	unkID, ok := token2id["<unk>"]
	if !ok {
		unkID = 1
	}

	return &Tokenizer{
		spSource: spSrc,
		spTarget: spTgt,
		config:   cfg,
		token2id: token2id,
		id2token: id2token,
		unkID:    unkID,
	}, nil
}

// Close releases any underlying native resources (SentencePiece, Marian, etc.).
func (t *Tokenizer) Close() {
	if t.spSource != nil {
		C.sp_free(t.spSource)
		t.spSource = nil
	}
	if t.spTarget != nil {
		C.sp_free(t.spTarget)
		t.spTarget = nil
	}
}

// Config returns the tokenizer configuration.
//
// The configuration is loaded and cached during tokenizer initialization and
// remains immutable for the lifetime of the tokenizer. The returned pointer
// refers to the tokenizer's internal cached copy and must not be modified by
// the caller.
func (t *Tokenizer) Config() (*marian.Config, error) {
	return &t.config, nil
}

// encodeInternal is the internal implementation: SP encode -> pieces -> vocab ids.
func (t *Tokenizer) encodeInternal(text string, addEOS bool) ([]int64, error) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	maxTokens := t.config.ModelMaxLength
	if addEOS && maxTokens > 0 {
		maxTokens -= 1
	}

	if maxTokens <= 0 {
		return nil, fmt.Errorf("model_max_length too small")
	}

	buf := make([]C.int, maxTokens)

	n := C.sp_encode_as_ids(t.spSource, cText, &buf[0], C.int(maxTokens))
	if n < 0 {
		return nil, fmt.Errorf("sp_encode_as_ids failed: %d", int(n))
	}

	// SP id -> piece
	pieces := make([]string, int(n))
	tmp := make([]C.char, maxTokens) // can be increased

	for i := 0; i < int(n); i++ {
		id := int(buf[i])

		res := C.sp_id_to_piece(t.spSource, C.int(id), &tmp[0], C.int(len(tmp)))
		if res < 0 {
			return nil, fmt.Errorf("sp_id_to_piece failed for id=%d, code=%d", id, int(res))
		}
		pieces[i] = C.GoString(&tmp[0])
	}

	// piece -> Marian id via vocab.json
	ids := make([]int64, 0, len(pieces)+1)
	for _, p := range pieces {
		if id, ok := t.token2id[p]; ok {
			ids = append(ids, id)
		} else {
			ids = append(ids, t.unkID)
		}
	}

	if addEOS {
		ids = append(ids, t.config.EosTokenID)
	}

	return ids, nil
}

// Encode encodes a single source sentence into token IDs.
// If addEOS is true, EOS token is appended.
func (t *Tokenizer) Encode(text string, addEOS bool) ([]int64, error) {
	return t.encodeInternal(text, addEOS)
}

// EncodeBatch encodes a batch of sentences and returns:
//  - inputIDs: shape (batch, maxLen)
//  - attentionMask: shape (batch, maxLen) with 1 for tokens and 0 for padding.
func (t *Tokenizer) EncodeBatch(texts []string) ([][]int64, [][]int64, error) {
	all := make([][]int64, len(texts))
	maxUsed := 0

	for i, s := range texts {
		ids, err := t.encodeInternal(s, true)
		if err != nil {
			return nil, nil, err
		}
		all[i] = ids
		if len(ids) > maxUsed {
			maxUsed = len(ids)
		}
	}

	batchSize := len(texts)
	inputIDs := make([][]int64, batchSize)
	attn := make([][]int64, batchSize)

	for i := 0; i < batchSize; i++ {
		inputIDs[i] = make([]int64, maxUsed)
		attn[i] = make([]int64, maxUsed)

		seq := all[i]
		for j := 0; j < maxUsed; j++ {
			if j < len(seq) {
				inputIDs[i][j] = seq[j]
				attn[i][j] = 1
			} else {
				inputIDs[i][j] = t.config.PadTokenID
				attn[i][j] = 0
			}
		}
	}

	return inputIDs, attn, nil
}

// Decode converts token IDs back to a target sentence.
// If skipSpecial is true, EOS / PAD / UNK are removed before decoding.
func (t *Tokenizer) Decode(ids []int64, skipSpecial bool) (string, error) {
	if skipSpecial {
		filtered := make([]int64, 0, len(ids))
		for _, id := range ids {
			if id == t.config.EosTokenID || id == t.config.PadTokenID || id == t.unkID {
				continue
			}
			filtered = append(filtered, id)
		}
		ids = filtered
	}

	if len(ids) == 0 {
		return "", nil
	}

	// Marian id -> token (piece string)
	pieces := make([]*C.char, len(ids))
	for i, id := range ids {
		if id < 0 || int(id) >= len(t.id2token) || t.id2token[id] == "" {
			pieces[i] = C.CString("<unk>")
		} else {
			pieces[i] = C.CString(t.id2token[id])
		}
	}
	// free all C strings
	defer func() {
		for _, p := range pieces {
			C.free(unsafe.Pointer(p))
		}
	}()

	const maxTextLen = 4096
	buf := make([]C.char, maxTextLen)

	n := C.sp_decode_pieces(
		t.spTarget,
		(**C.char)(unsafe.Pointer(&pieces[0])),
		C.int(len(pieces)),
		&buf[0],
		C.int(maxTextLen),
	)
	if n < 0 {
		return "", fmt.Errorf("sp_decode_pieces failed: %d", int(n))
	}

	return C.GoString(&buf[0]), nil
}
