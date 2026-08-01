package suzume

/*
#include "suzume/suzume_c.h"
*/
import "C"

// Numeric codes returned by the C ABI are decoded into the public string labels
// here, so the analyzer library carries no per-morpheme formatting. The core
// owns the label text and exports it through the ABI, so the tables below are
// read from the loaded library once at startup instead of being restated in Go,
// where a second copy would drift from the analyzer without anything noticing.
//
// The Japanese part-of-speech labels are the exception: the C ABI does not
// export them, so they are spelled out here.

// labelTableSize covers every value a uint8 code can take, which lets a lookup
// index the table directly and treat an empty entry as "unknown code".
const labelTableSize = 256

var (
	// posEnglishLabels maps a suzume_pos_t code to its English label.
	posEnglishLabels = readLabels(func(code uint8) *C.char {
		return C.suzume_pos_label(C.suzume_pos_t(code))
	})

	// extendedPOSLabels maps a suzume_extended_pos_t code to its stable code string.
	extendedPOSLabels = readLabels(func(code uint8) *C.char {
		return C.suzume_extended_pos_label(C.suzume_extended_pos_t(code))
	})

	// conjugationTypeLabels maps a suzume_conjugation_type_t code to its Japanese
	// label. Code 0 means "none" and decodes to the empty string.
	conjugationTypeLabels = readLabels(func(code uint8) *C.char {
		return C.suzume_conjugation_type_label(C.suzume_conjugation_type_t(code))
	})

	// conjugationFormLabels maps a suzume_conjugation_form_t code to its Japanese label.
	conjugationFormLabels = readLabels(func(code uint8) *C.char {
		return C.suzume_conjugation_form_label(C.suzume_conjugation_form_t(code))
	})
)

// posJapaneseLabels maps a suzume_pos_t code to its Japanese part-of-speech
// label. These are not part of the C ABI, so they are kept in step with the
// numeric codes by hand.
var posJapaneseLabels = [...]string{
	"その他",
	"名詞",
	"動詞",
	"形容詞",
	"副詞",
	"助詞",
	"助動詞",
	"接続詞",
	"連体詞",
	"代名詞",
	"接頭辞",
	"接尾辞",
	"感動詞",
	"記号",
	"その他",
}

// readLabels materializes one C label table into Go strings. A code the core
// does not recognize yields a NULL label, which decodes to the empty string and
// is mapped to the caller-facing fallback at lookup time.
func readLabels(label func(code uint8) *C.char) [labelTableSize]string {
	var table [labelTableSize]string
	for code := range table {
		table[code] = C.GoString(label(uint8(code)))
	}
	return table
}

// posEnglish decodes a numeric POS code to its English label.
func posEnglish(code uint8) string {
	if label := posEnglishLabels[code]; label != "" {
		return label
	}
	return "OTHER"
}

// posJapanese decodes a numeric POS code to its Japanese label.
func posJapanese(code uint8) string {
	if int(code) < len(posJapaneseLabels) {
		return posJapaneseLabels[code]
	}
	return "その他"
}

// extendedPOS decodes a numeric ExtendedPOS code to its stable string code.
func extendedPOS(code uint8) string {
	if label := extendedPOSLabels[code]; label != "" {
		return label
	}
	return "UNKNOWN"
}

// conjugationType decodes a numeric conjugation-type code to its Japanese label,
// returning the empty string when out of range or when the code means "none".
func conjugationType(code uint8) string {
	return conjugationTypeLabels[code]
}

// conjugationForm decodes a numeric conjugation-form code to its Japanese label,
// returning the empty string when out of range.
func conjugationForm(code uint8) string {
	return conjugationFormLabels[code]
}
