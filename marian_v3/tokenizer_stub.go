//go:build !(cgo && amd64 && (linux || windows))

package marian_v3

import (
	"fmt"

	"github.com/techwithsergiu/marian_tokenizer_go/marian"
)

type Tokenizer struct{}

// ensure interface implementation
var _ marian.Tokenizer = (*Tokenizer)(nil)

var ErrUnsupported = fmt.Errorf("marian_v3: supported only on linux/amd64 and windows/amd64 with cgo")

func NewTokenizer(modelDir string) (*Tokenizer, error) {
	return &Tokenizer{}, ErrUnsupported
}

func (t *Tokenizer) Close() {}

func (t *Tokenizer) Encode(text string, addEOS bool) ([]int64, error) {
	return nil, ErrUnsupported
}

func (t *Tokenizer) EncodeBatch(texts []string) ([][]int64, [][]int64, error) {
	return nil, nil, ErrUnsupported
}

func (t *Tokenizer) Decode(ids []int64, skipSpecial bool) (string, error) {
	return "", ErrUnsupported
}
