package suzume

// Numeric codes returned by the C ABI are decoded into the public string labels
// here, so the analyzer library carries no per-morpheme English/Japanese
// formatting. The tables mirror the codes emitted by the Suzume core and must
// stay in lockstep with them.

// posEnglishLabels maps a suzume_pos_t code to its English part-of-speech label.
var posEnglishLabels = [...]string{
	"OTHER",
	"NOUN",
	"VERB",
	"ADJ",
	"ADV",
	"PARTICLE",
	"AUX",
	"CONJ",
	"DET",
	"PRON",
	"PREFIX",
	"SUFFIX",
	"INTJ",
	"SYMBOL",
	"OTHER",
}

// posJapaneseLabels maps a suzume_pos_t code to its Japanese part-of-speech label.
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

// extendedPOSLabels maps a suzume_extended_pos_t code to its stable string code.
var extendedPOSLabels = [...]string{
	"UNKNOWN",
	"VERB_終止",
	"VERB_連用",
	"VERB_未然",
	"VERB_音便",
	"VERB_て形",
	"VERB_仮定",
	"VERB_命令",
	"VERB_連体",
	"VERB_た形",
	"VERB_たら形",
	"ADJ_終止",
	"ADJ_連用",
	"ADJ_語幹",
	"ADJ_かっ",
	"ADJ_け形",
	"ADJ_NA",
	"AUX_過去",
	"AUX_丁寧",
	"AUX_否定",
	"AUX_否定古",
	"AUX_願望",
	"AUX_意志",
	"AUX_受身",
	"AUX_使役",
	"AUX_可能",
	"AUX_継続",
	"AUX_完了",
	"AUX_準備",
	"AUX_試行",
	"AUX_進行",
	"AUX_接近",
	"AUX_開始",
	"AUX_様態",
	"AUX_推定",
	"AUX_みたい",
	"AUX_断定",
	"AUX_丁寧断定",
	"AUX_尊敬",
	"AUX_丁重",
	"AUX_過度",
	"AUX_ガル",
	"PART_格",
	"PART_係",
	"PART_終",
	"PART_接続",
	"PART_引用",
	"PART_副",
	"PART_準体",
	"PART_係結",
	"NOUN",
	"NOUN_形式",
	"NOUN_転成",
	"NOUN_固有",
	"NOUN_姓",
	"NOUN_名",
	"NOUN_数",
	"PRON",
	"PRON_疑問",
	"ADV",
	"ADV_引用",
	"CONJ",
	"DET",
	"PREFIX",
	"SUFFIX",
	"SYMBOL",
	"INTJ",
	"OTHER",
	"ADJ_未然",
	"AUX_打消推量",
	"AUX_文語断定",
	"AUX_文語過去",
	"AUX_文語断定連体",
	"AUX_文語完了",
	"AUX_文語当為",
	"AUX_不可能",
	"AUX_授受",
	"SUFFIX_直後",
	"SUFFIX_傾向",
	"DET_引用",
	"AUX_よう",
	"AUX_KURUWA_POLITE",
}

// conjugationTypeLabels maps a suzume_conjugation_type_t code to its Japanese
// label. Index 0 means "none" and decodes to the empty string.
var conjugationTypeLabels = [...]string{
	"",
	"一段",
	"五段・カ行",
	"五段・ガ行",
	"五段・サ行",
	"五段・タ行",
	"五段・ナ行",
	"五段・バ行",
	"五段・マ行",
	"五段・ラ行",
	"五段・ワ行",
	"サ変",
	"カ変",
	"形容詞",
}

// conjugationFormLabels maps a suzume_conjugation_form_t code to its Japanese label.
var conjugationFormLabels = [...]string{
	"終止形",
	"未然形",
	"連用形",
	"連用形",
	"仮定形",
	"命令形",
	"意志形",
}

// posEnglish decodes a numeric POS code to its English label.
func posEnglish(code uint8) string {
	if int(code) < len(posEnglishLabels) {
		return posEnglishLabels[code]
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
	if int(code) < len(extendedPOSLabels) {
		return extendedPOSLabels[code]
	}
	return "UNKNOWN"
}

// conjugationType decodes a numeric conjugation-type code to its Japanese label,
// returning the empty string when out of range or when the code means "none".
func conjugationType(code uint8) string {
	if int(code) < len(conjugationTypeLabels) {
		return conjugationTypeLabels[code]
	}
	return ""
}

// conjugationForm decodes a numeric conjugation-form code to its Japanese label,
// returning the empty string when out of range.
func conjugationForm(code uint8) string {
	if int(code) < len(conjugationFormLabels) {
		return conjugationFormLabels[code]
	}
	return ""
}
