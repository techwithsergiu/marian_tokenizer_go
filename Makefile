# =========================
# Top-level Makefile
# =========================

#===
# Detect OS
UNAME_S := $(shell uname -s)

ifeq ($(OS),Windows_NT)
    # Native Windows CMD or PowerShell
    DETECTED_OS := windows
else ifeq ($(findstring MINGW,$(UNAME_S)),MINGW)
    # Git Bash / MSYS2 / MinGW
    DETECTED_OS := windows
else ifeq ($(findstring MSYS,$(UNAME_S)),MSYS)
    DETECTED_OS := windows
else ifeq ($(findstring CYGWIN,$(UNAME_S)),CYGWIN)
    DETECTED_OS := windows
else ifeq ($(UNAME_S),Linux)
    DETECTED_OS := linux
else ifeq ($(UNAME_S),Darwin)
    DETECTED_OS := darwin
else
    $(error Unsupported OS: $(UNAME_S))
endif

# Detect ARCH
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_M),x86_64)
    DETECTED_ARCH := amd64
else ifeq ($(UNAME_M),amd64)
    DETECTED_ARCH := amd64
else ifeq ($(UNAME_M),aarch64)
    DETECTED_ARCH := arm64
else ifeq ($(UNAME_M),arm64)
    DETECTED_ARCH := arm64
else
    $(error Unsupported ARCH: $(UNAME_M))
endif

TARGET_OS   ?= $(DETECTED_OS)
TARGET_ARCH ?= $(DETECTED_ARCH)

TARGET := $(TARGET_OS)_$(TARGET_ARCH)
#===


.PHONY: all deps marian_tokenizer_core upload_ru_en_model \
        build_v1 run_v1 demo_v1 \
        build_v2 run_v2 demo_v2 \
        build_v3 run_v3 demo_v3 \
        clean

# -------------------------
# Main targets
# -------------------------

# Make all deps and demos
all: deps build_v1 build_v2 build_v3

# Make all deps (marian_tokenizer_core + upload_ru_en_model)
deps: marian_tokenizer_core upload_ru_en_model

# -------------------------
# Deps: Marian core (C++)
# -------------------------

marian_tokenizer_core:
	@echo "Makefile - Building for: $(TARGET)"
	chmod +x ./scripts/build_marian_tokenizer_core.sh
	TARGET_OS=$(TARGET_OS) TARGET_ARCH=$(TARGET_ARCH) ./scripts/build_marian_tokenizer_core.sh

# -------------------------
# Deps: Upload a Tokenizer from a Helsinki-NLP/opus-mt-ru-en model
# -------------------------

upload_ru_en_model:
	chmod +x ./scripts/upload_ru_en_model.sh
	./scripts/upload_ru_en_model.sh

# -------------------------
# Demo v1
# Dynamic linking with SentencePiece
# -------------------------

# Build demo_v1
build_v1:
	CGO_ENABLED=1 go build ./marian_v1/cmd/demo_v1

# Run demo_v1
run_v1:
	CGO_ENABLED=1 LD_LIBRARY_PATH=./deps/sentencepiece/$(TARGET)/lib go run ./marian_v1/cmd/demo_v1

# Build and run
demo_v1: build_v1 run_v1

# -------------------------
# Demo v2
# Static SentencePiece, core inside binary file
# -------------------------

build_v2:
	CGO_ENABLED=1 go build ./marian_v2/cmd/demo_v2

run_v2:
	CGO_ENABLED=1 go run ./marian_v2/cmd/demo_v2

demo_v2: build_v2 run_v2

# -------------------------
# Demo v3
# Dynamic linking with libmarian_core.so
# -------------------------

build_v3:
	CGO_ENABLED=1 go build ./marian_v3/cmd/demo_v3

run_v3:
	CGO_ENABLED=1 LD_LIBRARY_PATH=./deps/marian_tokenizer_core/$(TARGET)/lib go run ./marian_v3/cmd/demo_v3

demo_v3: build_v3 run_v3

# -------------------------
# Clean
# -------------------------

clean:
	rm -rf ./deps/marian_tokenizer_core/include ./deps/marian_tokenizer_core/$(TARGET) ./deps/sentencepiece/include ./deps/sentencepiece/$(TARGET)
