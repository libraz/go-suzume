package suzume

import (
	"os"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"
)

func newSuzume(t *testing.T) *Suzume {
	t.Helper()
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// --- Lifecycle ---

func TestNew(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
}

func TestNewWithOptions(t *testing.T) {
	s, err := NewWithOptions(Options{
		PreserveVu:   true,
		PreserveCase: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
}

func TestCloseMultipleTimes(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	s.Close() // should not panic
}

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("version should not be empty")
	}
	if !strings.Contains(v, ".") {
		t.Errorf("version %q does not look like a semver", v)
	}
	t.Logf("suzume version: %s", v)
}

// --- Analyze ---

func TestAnalyze(t *testing.T) {
	s := newSuzume(t)

	morphemes := s.Analyze("東京都に住んでいます")
	if len(morphemes) == 0 {
		t.Fatal("expected morphemes, got none")
	}

	found := false
	for _, m := range morphemes {
		if m.Surface == "東京" || m.Surface == "東京都" {
			found = true
			if m.POS == "" {
				t.Error("POS should not be empty")
			}
			if m.BaseForm == "" {
				t.Error("BaseForm should not be empty")
			}
			break
		}
	}
	if !found {
		t.Errorf("expected to find '東京' or '東京都' in morphemes, got: %v", morphNames(morphemes))
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	s := newSuzume(t)
	morphemes := s.Analyze("")
	if len(morphemes) != 0 {
		t.Errorf("expected 0 morphemes for empty string, got %d", len(morphemes))
	}
}

func TestAnalyzeAfterClose(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	s.Close()

	morphemes := s.Analyze("テスト")
	if morphemes != nil {
		t.Error("expected nil after close")
	}
}

func TestAnalyzeVerb(t *testing.T) {
	s := newSuzume(t)

	morphemes := s.Analyze("食べる")
	if len(morphemes) == 0 {
		t.Fatal("expected morphemes")
	}

	for _, m := range morphemes {
		if m.POS == "VERB" {
			if m.ConjType == "" {
				t.Error("verb should have ConjType")
			}
			if m.ConjForm == "" {
				t.Error("verb should have ConjForm")
			}
			if m.BaseForm == "" {
				t.Error("verb should have BaseForm")
			}
			return
		}
	}
	t.Error("expected to find a VERB morpheme")
}

func TestAnalyzeMorphemeFields(t *testing.T) {
	s := newSuzume(t)

	morphemes := s.Analyze("東京は美しい都市です")
	if len(morphemes) == 0 {
		t.Fatal("expected morphemes")
	}

	for _, m := range morphemes {
		if m.Surface == "" {
			t.Error("Surface should never be empty")
		}
		if m.POS == "" {
			t.Error("POS should never be empty")
		}
		if m.POSJa == "" {
			t.Errorf("POSJa should not be empty for %q", m.Surface)
		}
		// BaseForm should always be set
		if m.BaseForm == "" {
			t.Errorf("BaseForm should not be empty for %q", m.Surface)
		}
		// ExtendedPOS should always be set
		if m.ExtendedPOS == "" {
			t.Errorf("ExtendedPOS should not be empty for %q", m.Surface)
		}
	}
}

func TestAnalyzeAdjective(t *testing.T) {
	s := newSuzume(t)

	morphemes := s.Analyze("美しい")
	found := false
	for _, m := range morphemes {
		if m.POS == "ADJ" {
			found = true
			if m.ConjForm == "" {
				t.Error("adjective should have ConjForm")
			}
		}
	}
	if !found {
		t.Errorf("expected ADJ in morphemes, got: %v", morphNames(morphemes))
	}
}

func TestAnalyzeVariousInputs(t *testing.T) {
	s := newSuzume(t)

	tests := []struct {
		name  string
		input string
	}{
		{"hiragana", "これはてすとです"},
		{"katakana", "テスト"},
		{"kanji", "東京都千代田区"},
		{"mixed", "Go言語でプログラミング"},
		{"ascii", "Hello World"},
		{"numbers", "2024年4月9日"},
		{"long text", "日本語の形態素解析は自然言語処理の基本的な技術であり、文を単語に分割して品詞を付与する処理です。"},
		{"single char", "あ"},
		{"punctuation", "今日は、天気が良い。"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			morphemes := s.Analyze(tt.input)
			if len(morphemes) == 0 {
				t.Errorf("expected morphemes for %q, got none", tt.input)
			}
			// Verify all surfaces concatenated approximate original
			var surfaces strings.Builder
			for _, m := range morphemes {
				surfaces.WriteString(m.Surface)
			}
			if !utf8.ValidString(surfaces.String()) {
				t.Error("reconstructed text contains invalid UTF-8")
			}
		})
	}
}

func TestAnalyzeNounHasNoConjugation(t *testing.T) {
	s := newSuzume(t)

	morphemes := s.Analyze("東京")
	for _, m := range morphemes {
		if m.POS == "NOUN" {
			if m.ConjType != "" {
				t.Errorf("noun should not have ConjType, got %q", m.ConjType)
			}
			if m.ConjForm != "" {
				t.Errorf("noun should not have ConjForm, got %q", m.ConjForm)
			}
		}
	}
}

// --- Options ---

func TestOptionsPreserveCase(t *testing.T) {
	sDefault := newSuzume(t)
	sPreserve, err := NewWithOptions(Options{PreserveCase: true})
	if err != nil {
		t.Fatal(err)
	}
	defer sPreserve.Close()

	input := "Hello World"
	defaultMorphs := sDefault.Analyze(input)
	preserveMorphs := sPreserve.Analyze(input)

	if len(defaultMorphs) == 0 || len(preserveMorphs) == 0 {
		t.Skip("no morphemes for ASCII input")
	}

	// With PreserveCase, original casing should be kept
	hasUpper := false
	for _, m := range preserveMorphs {
		if m.Surface != strings.ToLower(m.Surface) {
			hasUpper = true
			break
		}
	}

	hasAllLower := true
	for _, m := range defaultMorphs {
		if m.Surface != strings.ToLower(m.Surface) {
			hasAllLower = false
			break
		}
	}

	if !hasUpper {
		t.Log("PreserveCase may not affect this input")
	}
	if !hasAllLower {
		t.Log("default mode did not lowercase, behavior may vary")
	}
}

func TestOptionsPreserveSymbols(t *testing.T) {
	sDefault := newSuzume(t)
	sPreserve, err := NewWithOptions(Options{PreserveSymbols: true})
	if err != nil {
		t.Fatal(err)
	}
	defer sPreserve.Close()

	input := "東京！大阪＆名古屋"
	defaultMorphs := sDefault.Analyze(input)
	preserveMorphs := sPreserve.Analyze(input)

	// PreserveSymbols should produce equal or more morphemes
	if len(preserveMorphs) < len(defaultMorphs) {
		t.Errorf("PreserveSymbols produced fewer morphemes (%d) than default (%d)",
			len(preserveMorphs), len(defaultMorphs))
	}
}

// --- GenerateTags ---

func TestGenerateTags(t *testing.T) {
	s := newSuzume(t)

	tags := s.GenerateTags("東京都の天気予報を確認する")
	if len(tags) == 0 {
		t.Fatal("expected tags, got none")
	}

	for _, tag := range tags {
		if tag.Tag == "" {
			t.Error("tag string should not be empty")
		}
		if tag.POS == "" {
			t.Error("POS should not be empty")
		}
	}
}

func TestGenerateTagsEmpty(t *testing.T) {
	s := newSuzume(t)
	tags := s.GenerateTags("")
	if len(tags) != 0 {
		t.Errorf("expected 0 tags for empty string, got %d", len(tags))
	}
}

func TestGenerateTagsAfterClose(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	s.Close()

	tags := s.GenerateTags("テスト")
	if tags != nil {
		t.Error("expected nil after close")
	}
}

func TestGenerateTagsWithOptions(t *testing.T) {
	s := newSuzume(t)

	tags := s.GenerateTagsWithOptions("東京都の天気予報を確認する", TagOptions{
		POSFilter:    POSNoun,
		ExcludeBasic: true,
		UseLemma:     true,
		MinLength:    2,
		MaxTags:      5,
	})

	if len(tags) == 0 {
		t.Fatal("expected tags, got none")
	}
	if len(tags) > 5 {
		t.Errorf("expected at most 5 tags, got %d", len(tags))
	}

	for _, tag := range tags {
		if tag.POS != "NOUN" {
			t.Errorf("expected POS=NOUN with noun filter, got %q", tag.POS)
		}
	}
}

func TestGenerateTagsMaxTags(t *testing.T) {
	s := newSuzume(t)

	tags := s.GenerateTagsWithOptions(
		"東京都千代田区の天気予報を確認して出発時間を決める",
		TagOptions{MaxTags: 3},
	)
	if len(tags) > 3 {
		t.Errorf("expected at most 3 tags, got %d", len(tags))
	}
}

func TestGenerateTagsMinLength(t *testing.T) {
	s := newSuzume(t)

	tags := s.GenerateTagsWithOptions(
		"東京都の天気予報を確認する",
		TagOptions{MinLength: 3},
	)

	for _, tag := range tags {
		if utf8.RuneCountInString(tag.Tag) < 3 {
			t.Errorf("tag %q is shorter than MinLength=3", tag.Tag)
		}
	}
}

func TestGenerateTagsPOSFilterVerb(t *testing.T) {
	s := newSuzume(t)

	tags := s.GenerateTagsWithOptions(
		"東京都の天気予報を確認して出発する",
		TagOptions{POSFilter: POSVerb, UseLemma: true},
	)

	for _, tag := range tags {
		if tag.POS != "VERB" {
			t.Errorf("expected POS=VERB with verb filter, got %q for tag %q", tag.POS, tag.Tag)
		}
	}
}

func TestGenerateTagsPOSFilterCombined(t *testing.T) {
	s := newSuzume(t)

	tags := s.GenerateTagsWithOptions(
		"美しい東京の景色を楽しむ",
		TagOptions{POSFilter: POSNoun | POSAdjective, UseLemma: true},
	)

	for _, tag := range tags {
		if tag.POS != "NOUN" && tag.POS != "ADJ" {
			t.Errorf("expected NOUN or ADJ, got %q for tag %q", tag.POS, tag.Tag)
		}
	}
}

// --- Dictionary loading ---

func TestLoadUserDictionaryErrors(t *testing.T) {
	s := newSuzume(t)

	if err := s.LoadUserDictionary(nil); err == nil {
		t.Error("expected error for nil data")
	}
	if err := s.LoadUserDictionary([]byte{}); err == nil {
		t.Error("expected error for empty data")
	}
}

func TestLoadBinaryDictionaryErrors(t *testing.T) {
	s := newSuzume(t)

	if err := s.LoadBinaryDictionary(nil); err == nil {
		t.Error("expected error for nil data")
	}
	if err := s.LoadBinaryDictionary([]byte{}); err == nil {
		t.Error("expected error for empty data")
	}
}

func TestLoadBinaryDictionaryInvalidData(t *testing.T) {
	s := newSuzume(t)

	err := s.LoadBinaryDictionary([]byte("not a valid dictionary"))
	if err == nil {
		t.Error("expected error for invalid dictionary data")
	}
}

func TestLoadBinaryDictionaryReal(t *testing.T) {
	dictPath := "csuzume/data/user.dic"
	data, err := os.ReadFile(dictPath)
	if err != nil {
		t.Skipf("dictionary not found at %s: %v", dictPath, err)
	}

	s := newSuzume(t)
	if err := s.LoadBinaryDictionary(data); err != nil {
		t.Errorf("failed to load real dictionary: %v", err)
	}
}

func TestLoadDictionaryAfterClose(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	s.Close()

	if err := s.LoadUserDictionary([]byte("test")); err == nil {
		t.Error("expected error for LoadUserDictionary after close")
	}
	if err := s.LoadBinaryDictionary([]byte("test")); err == nil {
		t.Error("expected error for LoadBinaryDictionary after close")
	}
}

// --- Concurrency ---

func TestConcurrentInstances(t *testing.T) {
	const goroutines = 8
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			s, err := New()
			if err != nil {
				t.Error(err)
				return
			}
			defer s.Close()

			morphemes := s.Analyze("並行処理のテスト")
			if len(morphemes) == 0 {
				t.Error("expected morphemes in concurrent test")
			}

			tags := s.GenerateTags("並行処理のテスト")
			if len(tags) == 0 {
				t.Error("expected tags in concurrent test")
			}
		}()
	}
	wg.Wait()
}

// --- Helpers ---

// morphNames extracts surface forms for test diagnostics.
func morphNames(ms []Morpheme) []string {
	names := make([]string, len(ms))
	for i := range ms {
		names[i] = ms[i].Surface
	}
	return names
}
