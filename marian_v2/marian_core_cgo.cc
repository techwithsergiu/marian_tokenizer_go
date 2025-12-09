// marian_v2/marian_core_cgo.cc

// This file is only needed so that cgo compiles marian_core.cc
// and links the implementation of marian_tok_* into the v2 binary.

#include "../deps/marian_tokenizer_core/src/marian_core.cc"
