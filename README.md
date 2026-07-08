# go-suzume

[![CI](https://img.shields.io/github/actions/workflow/status/libraz/go-suzume/ci.yml?branch=main&label=CI)](https://github.com/libraz/go-suzume/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/libraz/go-suzume.svg)](https://pkg.go.dev/github.com/libraz/go-suzume)
[![Go](https://img.shields.io/badge/go-%E2%89%A51.23-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![CGO](https://img.shields.io/badge/requires-CGO-orange)](https://pkg.go.dev/cmd/cgo)

Go bindings for [Suzume](https://github.com/libraz/suzume) — a lightweight Japanese morphological analyzer with a small dictionary (<400KB).

## Overview

Suzume uses feature-based analysis with character patterns instead of large dictionary files, providing Japanese tokenization, POS tagging, and lemmatization.

| | MeCab | Suzume |
|---|---|---|
| **Dictionary** | 20-50MB+ | <400KB |
| **Unknown words** | Poor | Feature-based |
| **Setup** | Complex | Zero-config |
| **Binding** | C | C / WASM / Go (CGO) |

### Features

- **Morphological Analysis** — Tokenization with POS, base form, conjugation info, and character offsets
- **Tag Generation** — Keyword extraction with POS filtering and lemmatization
- **User Dictionary** — CSV and binary dictionary loading at runtime
- **Bundled Dictionary** — Core dictionary is embedded and auto-loaded; no external files required
- **Thread Safe** — Each instance is independent

## Prerequisites

- Go 1.26+
- C++17 compiler (GCC 8+, Clang 10+, Apple Clang 12+)
- CMake 3.15+

## Installation

```bash
go get github.com/libraz/go-suzume
```

Build the C++ library before first use:

```bash
cd $(go env GOPATH)/pkg/mod/github.com/libraz/go-suzume@latest
make lib
```

Or clone and build:

```bash
git clone https://github.com/libraz/go-suzume.git
cd go-suzume
make lib    # Fetches suzume source and builds libsuzume.a
make test   # Run tests
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/libraz/go-suzume"
)

func main() {
	s, err := suzume.New()
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// Morphological analysis
	morphemes := s.Analyze("東京都に住んでいます")
	for _, m := range morphemes {
		fmt.Printf("%s\t%s\t%s\n", m.Surface, m.POS, m.BaseForm)
	}

	// Tag generation (keyword extraction)
	tags := s.GenerateTags("東京都の天気予報を確認する")
	for _, t := range tags {
		fmt.Printf("%s (%s)\n", t.Tag, t.POS)
	}
}
```

## Tag Generation Options

Start from `DefaultTagOptions` to keep the library's default filtering, then
override only the fields you need. The zero value of `TagOptions` disables every
exclusion filter.

```go
opts := suzume.DefaultTagOptions()
opts.POSFilter = suzume.POSNoun // Nouns only
opts.MaxTags = 10               // Up to 10 tags

tags := s.GenerateTagsWithOptions("東京都の天気予報を確認する", opts)
```

## Analysis Modes

Use `NewWithExtendedOptions` to select the segmentation mode and toggle
lemmatization or compound merging. Start from `DefaultExtendedOptions`, since
the zero value of `ExtendedOptions` does not match the library defaults.

```go
opts := suzume.DefaultExtendedOptions()
opts.Mode = suzume.ModeSearch // Finer segmentation, merges noun compounds

s, err := suzume.NewWithExtendedOptions(opts)
if err != nil {
	log.Fatal(err)
}
defer s.Close()
```

Available modes: `ModeNormal` (default), `ModeSearch`, and `ModeSplit`.

## Development

```bash
make sync       # Fetch suzume C++ source from GitHub
make sync-local # Copy from ../suzume (local dev)
make lib        # Build libsuzume.a
make test       # Run tests
make test-race  # Run tests with race detector
make lint       # Run golangci-lint
make coverage   # Generate coverage report
```

## License

[MIT](LICENSE)
