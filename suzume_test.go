package suzume

import (
	"testing"
)

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

func TestAnalyze(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	morphemes := s.Analyze("東京都に住んでいます")
	if len(morphemes) == 0 {
		t.Fatal("expected morphemes, got none")
	}

	// Verify first morpheme has expected fields
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
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	morphemes := s.Analyze("")
	if len(morphemes) != 0 {
		t.Errorf("expected 0 morphemes for empty string, got %d", len(morphemes))
	}
}

func TestAnalyzeVerb(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	morphemes := s.Analyze("食べる")
	if len(morphemes) == 0 {
		t.Fatal("expected morphemes")
	}

	// Should find a verb with conjugation info
	for _, m := range morphemes {
		if m.POS == "VERB" {
			if m.ConjType == "" {
				t.Error("verb should have ConjType")
			}
			return
		}
	}
	t.Error("expected to find a VERB morpheme")
}

func TestGenerateTags(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

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

func TestGenerateTagsWithOptions(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

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

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("version should not be empty")
	}
	t.Logf("suzume version: %s", v)
}

func TestCloseMultipleTimes(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	s.Close() // should not panic
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

func TestLoadBinaryDictionary(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Empty data should fail
	if err := s.LoadBinaryDictionary(nil); err == nil {
		t.Error("expected error for nil data")
	}
	if err := s.LoadBinaryDictionary([]byte{}); err == nil {
		t.Error("expected error for empty data")
	}
}

// morphNames extracts surface forms for test diagnostics.
func morphNames(ms []Morpheme) []string {
	names := make([]string, len(ms))
	for i, m := range ms {
		names[i] = m.Surface
	}
	return names
}
