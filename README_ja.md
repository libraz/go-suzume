# go-suzume

[![CI](https://img.shields.io/github/actions/workflow/status/libraz/go-suzume/ci.yml?branch=main&label=CI)](https://github.com/libraz/go-suzume/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/libraz/go-suzume.svg)](https://pkg.go.dev/github.com/libraz/go-suzume)
[![codecov](https://codecov.io/gh/libraz/go-suzume/branch/main/graph/badge.svg)](https://codecov.io/gh/libraz/go-suzume)
[![License](https://img.shields.io/badge/license-MIT-blue)](https://github.com/libraz/go-suzume/blob/main/LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![CGO](https://img.shields.io/badge/requires-CGO-orange)](https://pkg.go.dev/cmd/cgo)

[Suzume](https://github.com/libraz/suzume) の Go バインディングです。Suzume は 400KB 未満の辞書で動く軽量な日本語トークナイザーです。

個人が余暇に開発しているプロジェクトで、まだ 1.0 前です。リリース間で API が変わる可能性があります。

## 概要

Suzume は大きな辞書ファイルではなく、文字パターンによる素性ベースの解析を使います。MeCab のような完全な形態素解析器ではなく、検索やアプリケーションコードで扱いやすいトークン境界を狙った設計ですが、品詞と原形も返します。

| | MeCab | Suzume |
|---|---|---|
| **辞書サイズ** | 20-50MB 以上 | 400KB 未満 |
| **未知語** | 苦手 | 素性ベース |
| **セットアップ** | 複雑 | 設定不要 |
| **バインディング** | C | C / WASM / Python / Go (CGO) |

具体例とトレードオフは [MeCab との違い](https://suzume.libraz.net/docs/mecab-comparison) を参照してください。

### 機能

- **解析** — 品詞・原形・活用情報と、正規化テキスト上の文字オフセット付きのトークン分割
- **タグ生成** — 品詞フィルタと原形化によるキーワード抽出
- **ユーザー辞書** — TSV とバイナリ辞書を実行時に読み込み、インスタンス単位で破棄可能
- **辞書の同梱** — コア辞書をパッケージに埋め込んで自動読み込み。外部ファイルは不要
- **並行性** — インスタンスごとに並行実行可能。単一インスタンスへの同時呼び出しは非対応

## 前提条件

- Go 1.26 以降
- C++17 コンパイラ (GCC 8+, Clang 10+, Apple Clang 12+)
- CMake 3.15 以降

## インストール

このパッケージは Suzume の C++ ソースから作る静的ライブラリにリンクし、同じソースから生成した辞書を埋め込みます。どちらも Go モジュールには含まれないため、`go get` だけではビルドできません。リポジトリをクローンして一度ビルドし、利用側のモジュールからそのクローンを参照してください。

```bash
git clone https://github.com/libraz/go-suzume.git
cd go-suzume
make lib    # suzume の C++ ソースを取得して libsuzume.a をビルド
make test   # 任意: テストを実行
```

利用側のモジュールでは次のようにします。

```bash
go mod edit -replace github.com/libraz/go-suzume=/path/to/go-suzume
go get github.com/libraz/go-suzume
```

ビルド成果物はクローン先に書き込まれるので、書き込み可能なディレクトリに置いてください。Go のモジュールキャッシュは読み取り専用なので、そこには置けません。

## クイックスタート

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

	// 形態素解析
	morphemes := s.Analyze("東京都に住んでいます")
	for _, m := range morphemes {
		fmt.Printf("%s\t%s\t%s\n", m.Surface, m.POS, m.BaseForm)
	}

	// タグ生成 (キーワード抽出)
	tags := s.GenerateTags("東京都の天気予報を確認する")
	for _, t := range tags {
		fmt.Printf("%s (%s)\n", t.Tag, t.POS)
	}
}
```

## タグ生成のオプション

ライブラリ既定のフィルタを保つには `DefaultTagOptions` から始めて、必要なフィールドだけ上書きします。`TagOptions` のゼロ値はすべての除外フィルタを無効にするので注意してください。

```go
opts := suzume.DefaultTagOptions()
opts.POSFilter = suzume.POSNoun // 名詞のみ
opts.MaxTags = 10               // 最大 10 件

tags := s.GenerateTagsWithOptions("東京都の天気予報を確認する", opts)
```

## 解析モード

分割モードの選択、原形化や複合語結合の切り替えには `NewWithExtendedOptions` を使います。`ExtendedOptions` のゼロ値はライブラリ既定と一致しないため、`DefaultExtendedOptions` から始めてください。

```go
opts := suzume.DefaultExtendedOptions()
opts.Mode = suzume.ModeSearch // 細かく分割し、名詞の複合語を結合する

s, err := suzume.NewWithExtendedOptions(opts)
if err != nil {
	log.Fatal(err)
}
defer s.Close()
```

利用できるモードは `ModeNormal` (既定)、`ModeSearch`、`ModeSplit` です。

作成済みのインスタンスは、辞書を読み直さずにモードを切り替えられます。

```go
if err := s.SetMode(suzume.ModeSplit); err != nil {
	log.Fatal(err)
}
```

## 文字オフセット

`Morpheme.Start` と `Morpheme.End` は*正規化後*のテキスト上の文字オフセットで、入力そのものとは限りません。オフセットからテキストを取り出すときは `AnalyzeWithNormalizedText` を使ってください。

```go
result := s.AnalyzeWithNormalizedText("東京都に住んでいます")
runes := []rune(result.NormalizedText)
for _, m := range result.Morphemes {
	fmt.Println(string(runes[m.Start:m.End]))
}
```

## 辞書

コア辞書とユーザー辞書はパッケージに埋め込まれ、自動で読み込まれます。ディスク上に配置する準備は不要です。追加の語彙は TSV (`表層形<TAB>品詞[<TAB>活用型][<TAB>原形]`) かコンパイル済みの `.dic` から実行時に読み込めます。読み込みは破棄するまで積み重なります。

```go
if err := s.LoadUserDictionary([]byte("ゲリラ豪雨\tNOUN\n")); err != nil {
	log.Fatal(err)
}
defer s.ClearUserDictionaries() // 同梱の辞書は読み込まれたまま残る
```

同梱辞書の扱いは `ExtendedOptions` で制御します。`SkipCoreDictionary` と `SkipUserDictionary` は読み込みを省き、`DataDirectory` は指定したディレクトリだけから辞書を読み込みます。コア辞書が見つからない場合はエラーにはならず分割精度が落ちるだけなので、`HasCoreDictionary` で読み込めたかどうかを確認できます。

## エラー

解析器が返す失敗は `*suzume.Error` で、メッセージに加えてライブラリの安定したエラーコードを持ちます。

```go
var serr *suzume.Error
if err := s.LoadBinaryDictionary(data); errors.As(err, &serr) {
	fmt.Println(serr.Code, serr.Message)
}
```

## 開発

```bash
make sync       # GitHub から suzume の C++ ソースを取得
make sync-local # ../suzume からコピー (ローカル開発用)
make lib        # libsuzume.a をビルド
make test       # テストを実行
make test-race  # レース検出付きでテストを実行
make lint       # golangci-lint を実行
make coverage   # カバレッジレポートを生成
```

## ドキュメント

- [Go API リファレンス](https://pkg.go.dev/github.com/libraz/go-suzume)
- [Go バインディングガイド](https://suzume.libraz.net/docs/go)
- [ユーザー辞書](https://suzume.libraz.net/docs/user-dictionary)

## ライセンス

[MIT](LICENSE)。本パッケージがビルド対象とする Suzume の C++ ソースは、別途 [Apache-2.0](https://github.com/libraz/suzume/blob/main/LICENSE) でライセンスされています。
