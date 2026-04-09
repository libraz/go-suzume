package suzume

/*
#cgo CXXFLAGS: -std=c++17 -I${SRCDIR}/csuzume/src
#cgo LDFLAGS: -L${SRCDIR}/csuzume/build/lib -lsuzume -lsuzume_analysis -lsuzume_postprocess -lsuzume_grammar -lsuzume_dictionary -lsuzume_normalize -lsuzume_pretokenizer -lsuzume_core -lstdc++ -lm
#cgo darwin LDFLAGS: -framework CoreFoundation
#include "csuzume/src/suzume_c.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

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

// NewWithOptions creates a new Suzume instance with the given options.
func NewWithOptions(opts Options) (*Suzume, error) {
	var copts C.suzume_options_t
	if opts.PreserveVu {
		copts.preserve_vu = 1
	}
	if opts.PreserveCase {
		copts.preserve_case = 1
	}
	if opts.PreserveSymbols {
		copts.preserve_symbols = 1
	}

	h := C.suzume_create_with_options(&copts)
	if h == nil {
		return nil, errors.New("failed to create suzume instance with options")
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
		morphemes[i] = Morpheme{
			Surface:     C.GoString(cm.surface),
			POS:         C.GoString(cm.pos),
			BaseForm:    C.GoString(cm.base_form),
			Reading:     C.GoString(cm.reading),
			POSJa:       C.GoString(cm.pos_ja),
			ExtendedPOS: C.GoString(cm.extended_pos),
		}
		if cm.conj_type != nil {
			morphemes[i].ConjType = C.GoString(cm.conj_type)
		}
		if cm.conj_form != nil {
			morphemes[i].ConjForm = C.GoString(cm.conj_form)
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
		return errors.New("failed to load binary dictionary")
	}
	return nil
}

// Version returns the Suzume library version string.
func Version() string {
	return C.GoString(C.suzume_version())
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
			POS: C.GoString(cPOS[i]),
		}
	}
	return tags
}
