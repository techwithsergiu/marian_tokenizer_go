#!/usr/bin/env bash
rm -rf ./models
mkdir -p ./models/opus-mt-ru-en
cd ./models/opus-mt-ru-en

wget --no-check-certificate https://huggingface.co/Helsinki-NLP/opus-mt-ru-en/resolve/main/config.json
wget --no-check-certificate https://huggingface.co/Helsinki-NLP/opus-mt-ru-en/resolve/main/vocab.json
wget --no-check-certificate https://huggingface.co/Helsinki-NLP/opus-mt-ru-en/resolve/main/target.spm
wget --no-check-certificate https://huggingface.co/Helsinki-NLP/opus-mt-ru-en/resolve/main/source.spm
