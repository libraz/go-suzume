# go-suzume

[![CI](https://img.shields.io/github/actions/workflow/status/libraz/go-suzume/ci.yml?branch=main&label=CI)](https://github.com/libraz/go-suzume/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/libraz/go-suzume.svg)](https://pkg.go.dev/github.com/libraz/go-suzume)
[![codecov](https://codecov.io/gh/libraz/go-suzume/branch/main/graph/badge.svg)](https://codecov.io/gh/libraz/go-suzume)
[![License](https://img.shields.io/badge/license-MIT-blue)](https://github.com/libraz/go-suzume/blob/main/LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![CGO](https://img.shields.io/badge/requires-CGO-orange)](https://pkg.go.dev/cmd/cgo)

Go bindings for [Suzume](https://github.com/libraz/suzume) — a lightweight Japanese tokenizer with a small dictionary (<400KB).

This is a spare-time project and is pre-1.0: the API can still change between
releases.

## Overview

Suzume uses feature-based analysis with character patterns instead of large
dictionary files. It is not a full morphological analyzer like MeCab — it aims
at practical token boundaries for search and application code — but it still
returns POS tags and lemmas.

| | MeCab | Suzume |
|---|---|---|
| **Dictionary** | 20-50MB+ | <400KB |
| **Unknown words** | Poor | Feature-based |
| **Setup** | Complex | Zero-config |
| **Binding** | C | C / WASM / Python / Go (CGO) |

See [Differences from MeCab](https://suzume.libraz.net/docs/mecab-comparison)
for examples and trade-offs.

### Features

- **Analysis** — Tokenization with POS, base form, conjugation info, and character offsets into the normalized text
- **Tag Generation** — Keyword extraction with POS filtering and lemmatization
- **User Dictionary** — TSV and binary dictionary loading at runtime, clearable per instance
- **Bundled Dictionary** — Core dictionary is embedded and auto-loaded; no external files required
- **Concurrency** — Separate instances run concurrently; a single instance is not safe for concurrent calls

## Prerequisites

- Go 1.26+
- C++17 compiler (GCC 8+, Clang 10+, Apple Clang 12+)
- CMake 3.15+

## Installation

The package links against a static library and embeds dictionaries that are
both built from the Suzume C++ sources, and neither is published inside the Go
module. `go get` on its own therefore fails to build: clone the repository,
build once, and point your module at the clone.

```bash
git clone https://github.com/libraz/go-suzume.git
cd go-suzume
make lib    # Fetches the suzume C++ source and builds libsuzume.a
make test   # Optional: run the test suite
```

Then, from the module that uses it:

```bash
go mod edit -replace github.com/libraz/go-suzume=/path/to/go-suzume
go get github.com/libraz/go-suzume
```

The build writes into the checkout, so it has to live in a directory you can
write to; the Go module cache is read-only and cannot host it.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	suzume "github.com/libraz/go-suzume"
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

An existing instance can switch modes without reloading its dictionaries:

```go
if err := s.SetMode(suzume.ModeSplit); err != nil {
	log.Fatal(err)
}
```

## Character Offsets

`Morpheme.Start` and `Morpheme.End` are character offsets into the *normalized*
text, which is not always the input. Use `AnalyzeWithNormalizedText` when the
offsets need to be resolved back to text.

```go
result := s.AnalyzeWithNormalizedText("東京都に住んでいます")
runes := []rune(result.NormalizedText)
for _, m := range result.Morphemes {
	fmt.Println(string(runes[m.Start:m.End]))
}
```

## Dictionaries

The core and user dictionaries are embedded in the package and loaded
automatically, so nothing has to be staged on disk. Additional entries are
loaded at runtime from TSV (`surface<TAB>POS[<TAB>conjugation type][<TAB>lemma]`)
or from a compiled `.dic` file; loads accumulate until they are cleared.

```go
if err := s.LoadUserDictionary([]byte("ゲリラ豪雨\tNOUN\n")); err != nil {
	log.Fatal(err)
}
defer s.ClearUserDictionaries() // Keeps the bundled dictionaries loaded
```

`ExtendedOptions` controls the bundled dictionaries: `SkipCoreDictionary` and
`SkipUserDictionary` leave them out, and `DataDirectory` loads dictionaries
exclusively from a directory of your own. `HasCoreDictionary` reports whether
the core dictionary was found, since a missing one degrades segmentation
instead of failing.

## Errors

Failures reported by the analyzer are `*suzume.Error`, which carries the
library's stable error code alongside the message.

```go
var serr *suzume.Error
if err := s.LoadBinaryDictionary(data); errors.As(err, &serr) {
	fmt.Println(serr.Code, serr.Message)
}
```

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

## Documentation

- [Go API reference](https://pkg.go.dev/github.com/libraz/go-suzume)
- [Go bindings guide](https://suzume.libraz.net/docs/go)
- [User dictionary](https://suzume.libraz.net/docs/user-dictionary)

## License

[MIT](LICENSE). The Suzume C++ sources this package builds against are licensed
separately under [Apache-2.0](https://github.com/libraz/suzume/blob/main/LICENSE).
