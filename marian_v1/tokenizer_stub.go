//go:build !(linux && amd64 && cgo)

package marian_v1

import (
	"fmt"

	"github.com/techwithsergiu/marian_tokenizer_go/marian"
)

type Tokenizer struct{}

// ensure interface implementation
var _ marian.Tokenizer = (*Tokenizer)(nil)

var ErrUnsupported = fmt.Errorf("marian_v1: tokenizer v1 is only supported on linux/amd64 with cgo")

func NewTokenizer(modelDir string) (*Tokenizer, error) {
	return &Tokenizer{}, ErrUnsupported
}

func (t *Tokenizer) Close() {}

func (t *Tokenizer) Config() (*marian.Config, error) {
	return nil, ErrUnsupported
}

func (t *Tokenizer) Encode(text string, addEOS bool) ([]int64, error) {
	return nil, ErrUnsupported
}

func (t *Tokenizer) EncodeBatch(texts []string) ([][]int64, [][]int64, error) {
	return nil, nil, ErrUnsupported
}

func (t *Tokenizer) Decode(ids []int64, skipSpecial bool) (string, error) {
	return "", ErrUnsupported
}
