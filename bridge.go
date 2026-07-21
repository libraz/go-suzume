package suzume

/*
#cgo CFLAGS: -I${SRCDIR}/csuzume/include -DSUZUME_STATIC
#cgo LDFLAGS: -L${SRCDIR}/csuzume/build/lib -lsuzume -lstdc++ -lm
#cgo darwin LDFLAGS: -framework CoreFoundation
#include "suzume/suzume_c.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// cBool converts a Go bool to the C uint8 convention (1 = true, 0 = false).
func cBool(b bool) C.uint8_t {
	if b {
		return 1
	}
	return 0
}

// Suzume is a Japanese morphological analyzer instance.
type Suzume struct {
	handle C.suzume_t
}

// New creates a new Suzume instance with default options.
func New() (*Suzume, error) {
	h := C.suzume_create()
	if h == nil {
		return nil, errors.New("failed to create suzume instance")
	}
	s := &Suzume{handle: h}
	runtime.SetFinalizer(s, (*Suzume).Close)
	return s, nil
}

// NewWithOptions creates a new Suzume instance with the given normalization
// options. It builds on the extended option set, keeping the library defaults
// for the analysis mode and lemmatization while applying the normalization
// toggles from opts. Use NewWithExtendedOptions for full control.
func NewWithOptions(opts Options) (*Suzume, error) {
	var copts C.suzume_extended_options_t
	C.suzume_init_extended_options(&copts)
	copts.preserve_vu = cBool(opts.PreserveVu)
	copts.preserve_case = cBool(opts.PreserveCase)
	copts.preserve_symbols = cBool(opts.PreserveSymbols)

	h := C.suzume_create_with_extended_options(&copts)
	if h == nil {
		if msg := LastError(); msg != "" {
			return nil, fmt.Errorf("failed to create suzume instance with options: %s", msg)
		}
		return nil, errors.New("failed to create suzume instance with options")
	}
	s := &Suzume{handle: h}
	runtime.SetFinalizer(s, (*Suzume).Close)
	return s, nil
}

// NewWithExtendedOptions creates a new Suzume instance with the extended option
// set, including the analysis mode, lemmatization, and compound merging.
//
// Prefer starting from DefaultExtendedOptions, since the zero value of
// ExtendedOptions does not match the library defaults.
func NewWithExtendedOptions(opts ExtendedOptions) (*Suzume, error) {
	var copts C.suzume_extended_options_t
	// Initialize defaults, then override every field from opts so the Go struct
	// is the single source of truth for the caller.
	C.suzume_init_extended_options(&copts)
	copts.preserve_vu = cBool(opts.PreserveVu)
	copts.preserve_case = cBool(opts.PreserveCase)
	copts.preserve_symbols = cBool(opts.PreserveSymbols)
	copts.mode = C.uint8_t(opts.Mode)
	copts.lemmatize = cBool(opts.Lemmatize)
	copts.merge_compounds = cBool(opts.MergeCompounds)

	h := C.suzume_create_with_extended_options(&copts)
	if h == nil {
		if msg := LastError(); msg != "" {
			return nil, fmt.Errorf("failed to create suzume instance with extended options: %s", msg)
		}
		return nil, errors.New("failed to create suzume instance with extended options")
	}
	s := &Suzume{handle: h}
	runtime.SetFinalizer(s, (*Suzume).Close)
	return s, nil
}

// Close destroys the Suzume instance and frees resources.
// Safe to call multiple times.
func (s *Suzume) Close() {
	if s.handle != nil {
		C.suzume_destroy(s.handle)
		s.handle = nil
		runtime.SetFinalizer(s, nil)
	}
}

// Analyze performs morphological analysis on the given Japanese text.
func (s *Suzume) Analyze(text string) []Morpheme {
	if s.handle == nil {
		return nil
	}

	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))

	result := C.suzume_analyze(s.handle, ctext)
	if result == nil {
		return nil
	}
	defer C.suzume_result_free(result)

	count := int(result.count)
	if count == 0 {
		return nil
	}

	morphemes := make([]Morpheme, count)
	cMorphemes := unsafe.Slice(result.morphemes, count)

	for i := 0; i < count; i++ {
		cm := cMorphemes[i]
		pos := uint8(cm.pos)
		flags := uint8(cm.flags)
		morphemes[i] = Morpheme{
			Surface:          C.GoString(cm.surface),
			POS:              posEnglish(pos),
			BaseForm:         C.GoString(cm.base_form),
			POSJa:            posJapanese(pos),
			ExtendedPOS:      extendedPOS(uint8(cm.extended_pos)),
			Start:            int(cm.start),
			End:              int(cm.end),
			IsUserDict:       flags&uint8(C.SUZUME_MORPHEME_USER_DICT) != 0,
			IsFormalNoun:     flags&uint8(C.SUZUME_MORPHEME_FORMAL_NOUN) != 0,
			IsLowInfo:        flags&uint8(C.SUZUME_MORPHEME_LOW_INFO) != 0,
			IsUnknown:        flags&uint8(C.SUZUME_MORPHEME_UNKNOWN) != 0,
			IsFromDictionary: flags&uint8(C.SUZUME_MORPHEME_FROM_DICTIONARY) != 0,
			Score:            float32(cm.score),
		}
		// Conjugation applies only to verbs (POS 2) and adjectives (POS 3);
		// other parts of speech leave the conjugation fields empty.
		if pos == uint8(C.SUZUME_POS_VERB) || pos == uint8(C.SUZUME_POS_ADJECTIVE) {
			morphemes[i].ConjType = conjugationType(uint8(cm.conjugation_type))
			morphemes[i].ConjForm = conjugationForm(uint8(cm.conjugation_form))
		}
	}

	return morphemes
}

// GenerateTags extracts keyword tags from the given Japanese text.
func (s *Suzume) GenerateTags(text string) []Tag {
	if s.handle == nil {
		return nil
	}

	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))

	result := C.suzume_generate_tags(s.handle, ctext)
	if result == nil {
		return nil
	}
	defer C.suzume_tags_free(result)

	return convertTags(result)
}

// GenerateTagsWithOptions extracts keyword tags with the given options.
func (s *Suzume) GenerateTagsWithOptions(text string, opts TagOptions) []Tag {
	if s.handle == nil {
		return nil
	}

	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))

	var copts C.suzume_tag_options_t
	copts.pos_filter = C.uint8_t(opts.POSFilter)
	if opts.ExcludeBasic {
		copts.exclude_basic = 1
	}
	if opts.UseLemma {
		copts.use_lemma = 1
	}
	copts.min_length = C.size_t(opts.MinLength)
	copts.max_tags = C.size_t(opts.MaxTags)
	if opts.ExcludeParticles {
		copts.exclude_particles = 1
	}
	if opts.ExcludeAuxiliaries {
		copts.exclude_auxiliaries = 1
	}
	if opts.ExcludeFormalNouns {
		copts.exclude_formal_nouns = 1
	}
	if opts.ExcludeLowInfo {
		copts.exclude_low_info = 1
	}
	if opts.RemoveDuplicates {
		copts.remove_duplicates = 1
	}

	result := C.suzume_generate_tags_with_options(s.handle, ctext, &copts)
	if result == nil {
		return nil
	}
	defer C.suzume_tags_free(result)

	return convertTags(result)
}

// LoadUserDictionary loads a CSV-format user dictionary from memory.
func (s *Suzume) LoadUserDictionary(data []byte) error {
	if s.handle == nil {
		return errors.New("suzume instance is closed")
	}
	if len(data) == 0 {
		return errors.New("dictionary data is empty")
	}

	cdata := (*C.char)(unsafe.Pointer(&data[0]))
	ok := C.suzume_load_user_dict(s.handle, cdata, C.size_t(len(data)))
	if ok == 0 {
		if msg := LastError(); msg != "" {
			return fmt.Errorf("failed to load user dictionary: %s", msg)
		}
		return errors.New("failed to load user dictionary")
	}
	return nil
}

// LoadBinaryDictionary loads a binary .dic format dictionary from memory.
func (s *Suzume) LoadBinaryDictionary(data []byte) error {
	if s.handle == nil {
		return errors.New("suzume instance is closed")
	}
	if len(data) == 0 {
		return errors.New("dictionary data is empty")
	}

	cdata := (*C.uint8_t)(unsafe.Pointer(&data[0]))
	ok := C.suzume_load_binary_dict(s.handle, cdata, C.size_t(len(data)))
	if ok == 0 {
		if msg := LastError(); msg != "" {
			return fmt.Errorf("failed to load binary dictionary: %s", msg)
		}
		return errors.New("failed to load binary dictionary")
	}
	return nil
}

// DictionaryWarnings returns the warnings accumulated while auto-loading
// dictionaries for this instance. It returns nil when there are none.
func (s *Suzume) DictionaryWarnings() []string {
	if s.handle == nil {
		return nil
	}
	count := int(C.suzume_dictionary_warning_count(s.handle))
	if count == 0 {
		return nil
	}
	warnings := make([]string, 0, count)
	for i := 0; i < count; i++ {
		w := C.suzume_dictionary_warning(s.handle, C.size_t(i))
		if w == nil {
			continue
		}
		warnings = append(warnings, C.GoString(w))
	}
	return warnings
}

// Version returns the Suzume library version string.
func Version() string {
	return C.GoString(C.suzume_version())
}

// LastError returns the last C API error message for the current thread, or an
// empty string when none is set.
func LastError() string {
	return C.GoString(C.suzume_last_error())
}

// convertTags converts a C suzume_tags_t to a Go []Tag slice.
func convertTags(result *C.suzume_tags_t) []Tag {
	count := int(result.count)
	if count == 0 {
		return nil
	}

	tags := make([]Tag, count)
	cTags := unsafe.Slice(result.tags, count)
	cPOS := unsafe.Slice(result.pos, count)

	for i := 0; i < count; i++ {
		tags[i] = Tag{
			Tag: C.GoString(cTags[i]),
			POS: posEnglish(uint8(cPOS[i])),
		}
	}
	return tags
}
