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

	// POSJa is the part of speech in Japanese (e.g. "名詞", "動詞").
	POSJa string

	// ConjType is the conjugation type in Japanese, empty unless
	// IsConjugatable is set. Some conjugatable morphemes carry no type.
	ConjType string

	// ConjForm is the conjugation form in Japanese, empty unless
	// IsConjugatable is set.
	ConjForm string

	// ExtendedPOS is the stable extended POS code (e.g. "VERB_連用").
	ExtendedPOS string

	// Start is the start character offset in the normalized text, which
	// AnalyzeWithNormalizedText returns alongside the morphemes.
	Start int

	// End is the end character offset in the normalized text.
	End int

	// IsUserDict reports whether the morpheme came from a user dictionary.
	IsUserDict bool

	// IsFormalNoun reports whether the morpheme is a formal noun (形式名詞).
	IsFormalNoun bool

	// IsLowInfo reports whether the morpheme is a low information word.
	IsLowInfo bool

	// IsUnknown reports whether the morpheme is an unknown word.
	IsUnknown bool

	// IsFromDictionary reports whether the morpheme came from a dictionary.
	IsFromDictionary bool

	// IsConjugatable reports whether the morpheme conjugates, which is when
	// ConjType and ConjForm carry meaning.
	IsConjugatable bool

	// Score is the candidate score/cost assigned by the analyzer.
	Score float32
}

// AnalysisResult pairs the analyzed morphemes with the normalized text their
// Start and End offsets index into.
type AnalysisResult struct {
	// NormalizedText is the text the analyzer produced after normalization.
	NormalizedText string

	// Morphemes are the analyzed morphemes, in surface order.
	Morphemes []Morpheme
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

	// PreserveSymbols keeps punctuation-like symbol tokens in the output;
	// currency, unit and emoji code points are always emitted with POS "OTHER".
	PreserveSymbols bool
}

// AnalysisMode selects how the analyzer segments text.
type AnalysisMode int

const (
	// ModeNormal is the default segmentation mode.
	ModeNormal AnalysisMode = 0

	// ModeSearch favors finer segmentation and merges noun compounds,
	// suited for building search indices.
	ModeSearch AnalysisMode = 1

	// ModeSplit produces the finest-grained segmentation.
	ModeSplit AnalysisMode = 2

	// ModeInvalid is reported by Mode for an instance that has been closed.
	ModeInvalid AnalysisMode = 255
)

// ExtendedOptions configures the Suzume analyzer with the full option set,
// including the analysis mode, lemmatization, and compound merging.
//
// The zero value does not match the library defaults: it disables
// lemmatization and case/vu preservation. Start from DefaultExtendedOptions
// and override only the fields you need.
type ExtendedOptions struct {
	// PreserveVu preserves ヴ characters (don't normalize to ビ etc.)
	// (default: true).
	PreserveVu bool

	// PreserveCase preserves letter case (don't lowercase ASCII)
	// (default: true).
	PreserveCase bool

	// PreserveSymbols keeps punctuation-like symbol tokens in the output;
	// currency, unit and emoji code points are always emitted with POS "OTHER".
	PreserveSymbols bool

	// Mode selects the segmentation mode (default: ModeNormal).
	Mode AnalysisMode

	// Lemmatize applies lemmatization to derive base forms (default: true).
	Lemmatize bool

	// MergeCompounds merges consecutive noun compounds into a single token.
	MergeCompounds bool

	// SkipUserDictionary skips the automatic loading of the bundled user
	// dictionary.
	SkipUserDictionary bool

	// SkipCoreDictionary skips the automatic loading of the bundled core
	// dictionary.
	SkipCoreDictionary bool

	// ReportScorerConfig adds the scorer configuration to the diagnostics
	// reported by DictionaryWarnings.
	ReportScorerConfig bool

	// SkipEnvConfig ignores the analyzer's scorer configuration environment
	// variables.
	SkipEnvConfig bool

	// ScorerOptionsJSON overrides individual scorer parameters with a UTF-8
	// JSON document. It takes precedence over the environment configuration
	// unless SkipEnvConfig disables that.
	ScorerOptionsJSON string

	// DataDirectory loads dictionaries exclusively from this directory instead
	// of searching the default locations, which includes the dictionaries this
	// package embeds.
	DataDirectory string
}

// DefaultExtendedOptions returns ExtendedOptions populated with the library
// default values. Use it as a starting point and override only the fields you
// need.
func DefaultExtendedOptions() ExtendedOptions {
	return ExtendedOptions{
		PreserveVu:   true,
		PreserveCase: true,
		Mode:         ModeNormal,
		Lemmatize:    true,
	}
}

// POS filter bitmask constants for TagOptions.POSFilter.
const (
	POSNoun      uint8 = 1  // Keep nouns
	POSVerb      uint8 = 2  // Keep verbs
	POSAdjective uint8 = 4  // Keep adjectives
	POSAdverb    uint8 = 8  // Keep adverbs
	POSParticle  uint8 = 16 // Keep particles
	POSAuxiliary uint8 = 32 // Keep auxiliaries
)

// TagOptions configures tag generation behavior.
//
// The zero value disables every exclusion filter, which differs from the
// library defaults. Start from DefaultTagOptions and override individual
// fields to keep the recommended filtering behavior.
type TagOptions struct {
	// POSFilter is a bitmask for POS filtering (0 = all filterable POS).
	// Use the POSNoun, POSVerb, POSAdjective, POSAdverb, POSParticle, and
	// POSAuxiliary constants.
	POSFilter uint8

	// ExcludeBasic excludes basic words (hiragana-only lemma).
	ExcludeBasic bool

	// UseLemma uses lemma instead of surface form for tags (default: true).
	UseLemma bool

	// MinLength is the minimum tag length in characters (default: 2).
	MinLength int

	// MaxTags is the maximum number of tags to return (0 = unlimited).
	MaxTags int

	// ExcludeParticles excludes particles (助詞) (default: true).
	ExcludeParticles bool

	// ExcludeAuxiliaries excludes auxiliaries (助動詞) (default: true).
	ExcludeAuxiliaries bool

	// ExcludeFormalNouns excludes formal nouns (形式名詞) (default: true).
	ExcludeFormalNouns bool

	// ExcludeLowInfo excludes low information words (default: true).
	ExcludeLowInfo bool

	// RemoveDuplicates removes duplicate tags (default: true).
	RemoveDuplicates bool
}

// DefaultTagOptions returns TagOptions populated with the library default
// values. Use it as a starting point and override only the fields you need.
func DefaultTagOptions() TagOptions {
	return TagOptions{
		UseLemma:           true,
		MinLength:          2,
		ExcludeParticles:   true,
		ExcludeAuxiliaries: true,
		ExcludeFormalNouns: true,
		ExcludeLowInfo:     true,
		RemoveDuplicates:   true,
	}
}
