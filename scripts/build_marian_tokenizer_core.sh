#!/usr/bin/env bash
set -euo pipefail

TARGET_OS="${TARGET_OS:-}"
TARGET_ARCH="${TARGET_ARCH:-}"

# ========================
# Cleanup
rm -rf ./deps
mkdir -p ./deps/sentencepiece
mkdir -p ./deps/marian_tokenizer_core

# 
make -C third_party/marian-tokenizer-core  \
    TARGET_OS="${TARGET_OS}" \
    TARGET_ARCH="${TARGET_ARCH}" \
    all

# Copy deps
cd ./third_party/marian-tokenizer-core
cp -r ./deps/sentencepiece/** ../../deps/sentencepiece
cp -r ./build/** ../../deps/marian_tokenizer_core
