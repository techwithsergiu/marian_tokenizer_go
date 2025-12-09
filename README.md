
# Marian Tokenizer, Go bindings

## Go Bindings and C++ Core (Static & Dynamic Linking)

This repository provides a complete, modular tokenizer stack for Marian NMT models using:

- **Google SentencePiece**
- **Custom C++ Marian tokenizer core** (`marian-tokenizer-core`)
- **Go bindings for three tokenizer versions**

The project demonstrates clean interoperability between C++, static/dynamic linking, and Go cgo bindings.

---

## Features

- Full SentencePiece encode/decode
- Marian vocab remapping (`vocab.json`)
- Config parsing (`config.json`)
- Batch encoding (`input_ids`, `attention_mask`)
- Static & dynamic linking options
- Modular C++ core reusable across languages
- Zero Python dependencies

---

## Repository Structure

```bash
marian_tokenizer_go/
│
├── deps/                           # Final built dependencies
│   ├── sentencepiece/              # libsentencepiece.a / .so / headers
│   │   ├── include/
│   │   └── '$OS_ARCH'/lib/
│   │       └── static/
│   └── marian-tokenizer-core/      # libmarian_core.so / headers
│       ├── include/
│       ├── src/
│       └── '$OS_ARCH'/lib/
│           └── static/
│
├── third_party/
│   └── marian-tokenizer-core/      # git submodule (TechWithSergiu marian-tokenizer-core + Google Sentencepiece)
│
├── marian_v1/                      # Version 1 - dynamic SP
│   ├── sp_wrapper.cc
│   ├── sp_wrapper.h
│   ├── tokenizer_cgo.go
│   ├── tokenizer_stub.go
│   └── cmd/demo_v1/
│           └── main.go
│
├── marian_v2/                      # Version 2 - fully static build
│   ├-─ marian_core_cgo.cc
│   ├── tokenizer_cgo.go
│   ├── tokenizer_stub.go
│   └── cmd/demo_v2/
│           └── main.go
│
├── marian_v3/                      # Version 3 - dynamic Marian core
│   ├── tokenizer_cgo.go
│   ├── tokenizer_stub.go
│   └── cmd/demo_v3/
│           └── main.go
│
├── models/opus-mt-ru-en/           # Tokenizer from a Helsinki-NLP/opus-mt-ru-en model
│          ├── config.json          # These files are not included
│          ├── source.spm
│          ├── target.spm
│          └── vocab.json
│
├── scripts/
│   ├── build_marian_tokenizer_core.sh
│   └── upload_ru_en_model.sh
│
├── Makefile
└── README.md
```

---

## Build & Setup

### Clone WITH submodules

```bash
git clone --recurse-submodules https://github.com/techwithsergiu/marian-tokenizer-go.git
cd marian-tokenizer-go

git submodule update --init --recursive
```

### Build all dependencies (SentencePiece + Marian Tokenizer Core + Tokenizer from a model)

```bash
make deps
```

This runs:

- `scripts/build_marian_tokenizer_core.sh`
- `scripts/upload_ru_en_model.sh`

and places all build artifacts into:

```bash
deps/sentencepiece/
deps/marian_tokenizer_core/

# Uploads a Tokenizer from a Helsinki-NLP/opus-mt-ru-en model
models/opus-mt-ru-en/
```

### Windows build all dependencies

To rebuild native C++ libraries on Windows you’ll need:
> Note: you don’t need this step if you only want to use the library.
Prebuilt binaries for common platforms are already included in the repo.

1. Install [MSYS2](https://www.msys2.org/).
2. Open **"MSYS2 MinGW 64-bit"** shell (not plain MSYS).
3. Install toolchain:

   ```bash
   pacman -Syu
   pacman -S base-devel git mingw-w64-x86_64-cmake mingw-w64-x86_64-make mingw-w64-x86_64-gcc mingw-w64-x86_64-go
   ```

4. Then cd to the project and run the make commands:

   ```bash
   make deps

   make demo_v1
   make demo_v2
   make demo_v3
   ```

   ### Optional step (for advanced users only)

   1. Download Arch Linux.
   2. Flash it to a USB drive.
   3. Remove Windows.
   4. Install Arch.
   5. Go back to the Linux section of this README and enjoy shorter instructions.

---

## Demo Versions

Three tokenizer versions demonstrate different linking modes.

---

### Version 1 - Pure SentencePiece (Dynamic Linking)

- Only linux with amd64 is supported
- Uses SentencePiece dynamically (`libsentencepiece.so`)
- No Marian C++ core
- Equivalent to the original Python prototype

Run

```bash
make demo_v1
```

or manually:

```bash
TARGET=linux_amd64
CGO_ENABLED=1 LD_LIBRARY_PATH=./deps/sentencepiece/$TARGET/lib go run ./marian_v1/cmd/demo_v1
```

Build:

```bash
CGO_ENABLED=1 go build ./marian_v1/cmd/demo_v1
```

---

### Version 2 - Fully Static (SP + Marian Tokenizer Core in one binary)

- Statically links:
  - `libsentencepiece.a`
  - `marian_core.cc` compiled inside Go package
- Produces a **standalone binary** with **no external shared libraries**
- Ideal for deployments, containers, CLI tools, AWS Lambda

Run

```bash
make demo_v2
```

or manually:

```bash
CGO_ENABLED=1 go run ./marian_v2/cmd/demo_v2
CGO_ENABLED=1 go build ./marian_v2/cmd/demo_v2
./demo_v2
```

---

### Version 3 - Marian Tokenizer Core (C++) (`libmarian_core.so`)

- Uses:
  - `libmarian_core.so` (dynamic)
  - `libsentencepiece.so` (dependency)
- Ideal for multi-language bindings:
  - Java (JNI), Node.js (N-API), Android NDK, C#, Python (ctypes)

Run

```bash
make demo_v3
```

or manually:

```bash
TARGET=linux_amd64
CGO_ENABLED=1 LD_LIBRARY_PATH=./deps/marian_tokenizer_core/$TARGET/lib go run ./marian_v3/cmd/demo_v3
```

---

## Architecture Overview

### Encoder/Decoder Flow

```bash
input text
   ↓
SentencePiece (source.spm)
   ↓
SPM pieces
   ↓
vocab.json remapping → token IDs
   ↓
max_length, pad_id
   ↓
input_ids + attention_mask
```

### Why three versions?

| Version | SP Linking | Marian Tokenizer Core | Type | Purpose |
|--------|------------|-------------|------|---------|
| **v1** | dynamic `.so` | none | dynamic | simplest, Python-like |
| **v2** | static `.a` | static compiled-in | fully static | ideal for production |
| **v3** | static `.a` | dynamic `.so` | shared library | ideal for multi-language reuse |

---

## Makefile Commands

| Command | Description |
|--------|-------------|
| `make deps` | Build SentencePiece + Marian Tokenizer Core + Tokenizer from a model |
| `make demo_v1` | Run version 1 |
| `make demo_v2` | Run version 2 |
| `make demo_v3` | Run version 3 |
| `make build_v1` / `build_v2` / `build_v3` | Build binaries |
| `make run_v1` / `run_v2` / `run_v3` | Run binaries |
| `make clean` | Remove generated binaries |

---

## License

This project is licensed under the **Apache License 2.0**.

You are free to use, modify, and distribute this software in both open-source
and commercial applications, as long as you comply with the terms of the
Apache 2.0 License.

Full license text:  
[LICENSE](LICENSE)

---

## Third-party Licenses

This project relies on several third-party libraries, all using permissive
licenses fully compatible with Apache 2.0:

- **Marian Tokenizer Core** — Apache License 2.0 (© TechWithSergiu)  
  [github.com/techwithsergiu/marian-tokenizer-core](https://github.com/techwithsergiu/marian-tokenizer-core)
- **SentencePiece (C++ core)** — Apache License 2.0 (© Google)  
  [github.com/google/sentencepiece](https://github.com/google/sentencepiece)

This makes the entire project fully Apache-compatible and safe for commercial use.

---
