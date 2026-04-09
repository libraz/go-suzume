// Package suzume provides Go bindings for the Suzume Japanese morphological analyzer.
//
// Suzume is a lightweight Japanese tokenizer that uses feature-based analysis
// with a small dictionary (<400KB), providing POS tagging and lemmatization
// without requiring large dictionary files.
//
// This package requires CGO and the Suzume C++ library. Run `make lib` to build
// the required static library before using this package.
//
// Basic usage:
//
//	s, err := suzume.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer s.Close()
//
//	morphemes := s.Analyze("東京都に住んでいます")
//	for _, m := range morphemes {
//	    fmt.Printf("%s\t%s\t%s\n", m.Surface, m.POS, m.BaseForm)
//	}
package suzume

// Morpheme represents a single morphological analysis result.
type Morpheme struct {
	// Surface is the surface form as it appears in the text.
	Surface string

	// POS is the part of speech in English (e.g. "NOUN", "VERB").
	POS string

	// BaseForm is the base/dictionary form (lemma).
	BaseForm string

	// Reading is the reading in katakana.
	Reading string

	// POSJa is the part of speech in Japanese (e.g. "名詞", "動詞").
	POSJa string

	// ConjType is the conjugation type in Japanese (only for verbs/adjectives).
	ConjType string

	// ConjForm is the conjugation form in Japanese (only for verbs/adjectives).
	ConjForm string

	// ExtendedPOS is the extended POS tag in English (e.g. "VerbRenyokei").
	ExtendedPOS string
}

// Tag represents a generated tag from text analysis.
type Tag struct {
	// Tag is the extracted keyword/tag string.
	Tag string

	// POS is the part of speech in English (e.g. "NOUN", "VERB").
	POS string
}

// Options configures the Suzume analyzer behavior.
type Options struct {
	// PreserveVu preserves ヴ characters (don't normalize to ビ etc.).
	PreserveVu bool

	// PreserveCase preserves letter case (don't lowercase ASCII).
	PreserveCase bool

	// PreserveSymbols preserves symbols/emoji in the output.
	PreserveSymbols bool
}

// POS filter bitmask constants for TagOptions.POSFilter.
const (
	POSNoun      uint8 = 1 // Filter to nouns only
	POSVerb      uint8 = 2 // Filter to verbs only
	POSAdjective uint8 = 4 // Filter to adjectives only
	POSAdverb    uint8 = 8 // Filter to adverbs only
)

// TagOptions configures tag generation behavior.
type TagOptions struct {
	// POSFilter is a bitmask for POS filtering (0 = all).
	// Use POSNoun, POSVerb, POSAdjective, POSAdverb constants.
	POSFilter uint8

	// ExcludeBasic excludes basic words (hiragana-only lemma).
	ExcludeBasic bool

	// UseLemma uses lemma instead of surface form for tags.
	UseLemma bool

	// MinLength is the minimum tag length in characters (default: 2).
	MinLength int

	// MaxTags is the maximum number of tags to return (0 = unlimited).
	MaxTags int
}
