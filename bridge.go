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

// init refuses to run against a library built from a different revision of the
// C ABI. A layout change is invisible at link time, since the symbols still
// resolve, and the binding would then read every result struct at the wrong
// offsets.
func init() {
	header := uint32(C.SUZUME_ABI_VERSION)
	if library := uint32(C.suzume_abi_version()); library != header {
		panic(fmt.Sprintf(
			"suzume: C ABI mismatch: headers describe revision %d, libsuzume.a provides revision %d; run 'make lib' to rebuild",
			header, library))
	}
}

// ErrorCode is a stable failure code reported by the Suzume C API.
type ErrorCode uint8

// Error codes reported by the C API alongside its diagnostic message.
const (
	ErrorCodeSuccess              ErrorCode = C.SUZUME_ERROR_SUCCESS
	ErrorCodeInvalidUTF8          ErrorCode = C.SUZUME_ERROR_INVALID_UTF8
	ErrorCodeDictionaryLoadFailed ErrorCode = C.SUZUME_ERROR_DICTIONARY_LOAD_FAILED
	ErrorCodeFileNotFound         ErrorCode = C.SUZUME_ERROR_FILE_NOT_FOUND
	ErrorCodeParse                ErrorCode = C.SUZUME_ERROR_PARSE
	ErrorCodeOutOfMemory          ErrorCode = C.SUZUME_ERROR_OUT_OF_MEMORY
	ErrorCodeInvalidInput         ErrorCode = C.SUZUME_ERROR_INVALID_INPUT
	ErrorCodeInternal             ErrorCode = C.SUZUME_ERROR_INTERNAL
)

// Error is a failure reported by the Suzume C API. Use errors.As to reach Code
// when the failure needs to be handled by kind rather than by message.
type Error struct {
	// Code is the stable error code the library set for this failure.
	Code ErrorCode

	// Message is the full message, combining the failed operation with the
	// library diagnostic.
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string { return e.Message }

// lastError builds an Error from the diagnostic the C API left behind, prefixed
// with what the binding was doing. Call it only from a goroutine pinned to the
// OS thread that made the failing call.
func lastError(what string) *Error {
	err := &Error{Code: LastErrorCode(), Message: what}
	if msg := LastError(); msg != "" {
		err.Message = what + ": " + msg
	}
	return err
}

// callWithError runs a C entry point that reports failure through the C API
// diagnostic. The goroutine is pinned for the call because that diagnostic is
// thread-local, and an unpinned goroutine can be rescheduled onto a different
// OS thread between making the call and reading the failure.
func callWithError(what string, call func() bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if call() {
		return nil
	}
	return lastError(what)
}

// newHandle wraps a C constructor, pinning the goroutine so a failure can be
// read back from the thread that produced it.
func newHandle(what string, create func() C.suzume_t) (*Suzume, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	h := create()
	if h == nil {
		return nil, lastError(what)
	}
	s := &Suzume{handle: h}
	runtime.SetFinalizer(s, (*Suzume).Close)
	return s, nil
}

// cBool converts a Go bool to the C uint8 convention (1 = true, 0 = false).
func cBool(b bool) C.uint8_t {
	if b {
		return 1
	}
	return 0
}

// emptyText backs the pointer handed to the C API for empty input, which
// rejects a NULL text even when the length is zero.
var emptyText [1]C.char

// cText exposes a Go string to the C API as a pointer/length pair without
// copying it. The analyzer only reads the bytes for the duration of the call,
// which is what the cgo rules require of Go memory passed to C. Callers must
// keep the string alive across the call.
func cText(text string) (*C.char, C.size_t) {
	if text == "" {
		return &emptyText[0], 0
	}
	return (*C.char)(unsafe.Pointer(unsafe.StringData(text))), C.size_t(len(text))
}

// Suzume is a Japanese morphological analyzer instance.
type Suzume struct {
	handle C.suzume_t
}

// New creates a new Suzume instance with default options.
func New() (*Suzume, error) {
	return newHandle("failed to create suzume instance", func() C.suzume_t {
		return C.suzume_create()
	})
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

	return newHandle("failed to create suzume instance with options", func() C.suzume_t {
		return C.suzume_create_with_extended_options(&copts)
	})
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
	copts.skip_user_dictionary = cBool(opts.SkipUserDictionary)
	copts.skip_core_dictionary = cBool(opts.SkipCoreDictionary)
	copts.report_scorer_config = cBool(opts.ReportScorerConfig)
	copts.skip_env_config = cBool(opts.SkipEnvConfig)

	// The C API borrows both strings for the duration of the create call only,
	// and reads NULL as "unset".
	if opts.ScorerOptionsJSON != "" {
		cjson := C.CString(opts.ScorerOptionsJSON)
		defer C.free(unsafe.Pointer(cjson))
		copts.scorer_options_json = cjson
	}
	if opts.DataDirectory != "" {
		cdir := C.CString(opts.DataDirectory)
		defer C.free(unsafe.Pointer(cdir))
		copts.data_directory = cdir
	}

	return newHandle("failed to create suzume instance with extended options", func() C.suzume_t {
		return C.suzume_create_with_extended_options(&copts)
	})
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

// Mode returns the analysis mode currently in effect, or ModeInvalid for a
// closed instance.
func (s *Suzume) Mode() AnalysisMode {
	if s.handle == nil {
		return ModeInvalid
	}
	return AnalysisMode(C.suzume_mode(s.handle))
}

// SetMode changes the analysis mode of an existing instance without reloading
// its dictionaries.
func (s *Suzume) SetMode(mode AnalysisMode) error {
	if s.handle == nil {
		return errors.New("suzume instance is closed")
	}
	return callWithError("failed to set analysis mode", func() bool {
		return C.suzume_set_mode(s.handle, C.uint8_t(mode)) != 0
	})
}

// Analyze performs morphological analysis on the given Japanese text.
func (s *Suzume) Analyze(text string) []Morpheme {
	return s.AnalyzeWithNormalizedText(text).Morphemes
}

// AnalyzeWithNormalizedText performs morphological analysis and also returns
// the normalized text that Morpheme.Start and Morpheme.End index into.
func (s *Suzume) AnalyzeWithNormalizedText(text string) AnalysisResult {
	var out AnalysisResult
	if s.handle == nil {
		return out
	}

	ctext, csize := cText(text)
	result := C.suzume_analyze_n(s.handle, ctext, csize)
	runtime.KeepAlive(text)
	if result == nil {
		return out
	}
	defer C.suzume_result_free(result)

	if result.normalized_text != nil {
		out.NormalizedText = C.GoStringN(result.normalized_text, C.int(result.normalized_text_size))
	}

	count := int(result.count)
	if count == 0 {
		return out
	}

	out.Morphemes = make([]Morpheme, count)
	cMorphemes := unsafe.Slice(result.morphemes, count)

	for i := range cMorphemes {
		cm := &cMorphemes[i]
		pos := uint8(cm.pos)
		flags := uint8(cm.flags)
		m := Morpheme{
			Surface:          C.GoStringN(cm.surface, C.int(cm.surface_size)),
			POS:              posEnglish(pos),
			BaseForm:         C.GoStringN(cm.base_form, C.int(cm.base_form_size)),
			POSJa:            posJapanese(pos),
			ExtendedPOS:      extendedPOS(uint8(cm.extended_pos)),
			Start:            int(cm.start),
			End:              int(cm.end),
			IsUserDict:       flags&uint8(C.SUZUME_MORPHEME_USER_DICT) != 0,
			IsFormalNoun:     flags&uint8(C.SUZUME_MORPHEME_FORMAL_NOUN) != 0,
			IsLowInfo:        flags&uint8(C.SUZUME_MORPHEME_LOW_INFO) != 0,
			IsUnknown:        flags&uint8(C.SUZUME_MORPHEME_UNKNOWN) != 0,
			IsFromDictionary: flags&uint8(C.SUZUME_MORPHEME_FROM_DICTIONARY) != 0,
			IsConjugatable:   flags&uint8(C.SUZUME_MORPHEME_CONJUGATABLE) != 0,
			Score:            float32(cm.score),
		}
		// The analyzer flags every morpheme whose conjugation fields carry
		// meaning, which covers auxiliaries as well as verbs and adjectives;
		// the remaining parts of speech leave those fields empty.
		if m.IsConjugatable {
			m.ConjType = conjugationType(uint8(cm.conjugation_type))
			m.ConjForm = conjugationForm(uint8(cm.conjugation_form))
		}
		out.Morphemes[i] = m
	}

	return out
}

// GenerateTags extracts keyword tags from the given Japanese text.
func (s *Suzume) GenerateTags(text string) []Tag {
	if s.handle == nil {
		return nil
	}

	ctext, csize := cText(text)
	result := C.suzume_generate_tags_n(s.handle, ctext, csize)
	runtime.KeepAlive(text)
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

	var copts C.suzume_tag_options_t
	// Seed the C defaults so a field this binding does not expose yet keeps its
	// documented value, then override everything opts covers.
	C.suzume_init_tag_options(&copts)
	copts.pos_filter = C.uint8_t(opts.POSFilter)
	copts.exclude_basic = cBool(opts.ExcludeBasic)
	copts.use_lemma = cBool(opts.UseLemma)
	copts.min_length = C.size_t(max(opts.MinLength, 0))
	copts.max_tags = C.size_t(max(opts.MaxTags, 0))
	copts.exclude_particles = cBool(opts.ExcludeParticles)
	copts.exclude_auxiliaries = cBool(opts.ExcludeAuxiliaries)
	copts.exclude_formal_nouns = cBool(opts.ExcludeFormalNouns)
	copts.exclude_low_info = cBool(opts.ExcludeLowInfo)
	copts.remove_duplicates = cBool(opts.RemoveDuplicates)

	ctext, csize := cText(text)
	result := C.suzume_generate_tags_with_options_n(s.handle, ctext, csize, &copts)
	runtime.KeepAlive(text)
	if result == nil {
		return nil
	}
	defer C.suzume_tags_free(result)

	return convertTags(result)
}

// LoadUserDictionary loads user dictionary entries from memory. The data is TSV
// (surface, POS, optional conjugation type, optional lemma); the legacy
// 3-column CSV form is also accepted. Loads accumulate until
// ClearUserDictionaries is called.
func (s *Suzume) LoadUserDictionary(data []byte) error {
	_, err := s.LoadUserDictionaryCount(data)
	return err
}

// LoadUserDictionaryCount loads user dictionary entries from memory like
// LoadUserDictionary and reports how many entries were installed.
func (s *Suzume) LoadUserDictionaryCount(data []byte) (int, error) {
	if s.handle == nil {
		return 0, errors.New("suzume instance is closed")
	}
	if len(data) == 0 {
		return 0, errors.New("dictionary data is empty")
	}

	var installed C.size_t
	err := callWithError("failed to load user dictionary", func() bool {
		cdata := (*C.char)(unsafe.Pointer(&data[0]))
		installed = C.suzume_load_user_dict_count(s.handle, cdata, C.size_t(len(data)))
		runtime.KeepAlive(data)
		// The analyzer reports a source it could parse but found no entries in
		// as a failure, so a zero count always comes with a diagnostic.
		return installed > 0
	})
	if err != nil {
		return 0, err
	}
	return int(installed), nil
}

// LoadBinaryDictionary loads a binary .dic format dictionary from memory.
// Loads accumulate, and a failed load leaves the existing dictionaries intact.
func (s *Suzume) LoadBinaryDictionary(data []byte) error {
	if s.handle == nil {
		return errors.New("suzume instance is closed")
	}
	if len(data) == 0 {
		return errors.New("dictionary data is empty")
	}

	return callWithError("failed to load binary dictionary", func() bool {
		cdata := (*C.uint8_t)(unsafe.Pointer(&data[0]))
		ok := C.suzume_load_binary_dict(s.handle, cdata, C.size_t(len(data)))
		runtime.KeepAlive(data)
		return ok != 0
	})
}

// ClearUserDictionaries removes the dictionaries loaded through this instance.
// The bundled user dictionary the analyzer loads on its own stays installed.
func (s *Suzume) ClearUserDictionaries() error {
	if s.handle == nil {
		return errors.New("suzume instance is closed")
	}
	return callWithError("failed to clear user dictionaries", func() bool {
		return C.suzume_clear_user_dictionaries(s.handle) != 0
	})
}

// HasCoreDictionary reports whether the instance loaded the binary core
// dictionary. It is false when the dictionary could not be found, which
// degrades segmentation quality rather than failing outright.
func (s *Suzume) HasCoreDictionary() bool {
	if s.handle == nil {
		return false
	}
	return C.suzume_has_core_dictionary(s.handle) != 0
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

// LastError returns the last C API error message for the current OS thread, or
// an empty string when none is set.
//
// The C API keeps this diagnostic per OS thread while the Go runtime is free to
// move a goroutine between threads, so a later read is not guaranteed to see
// the failure. Prefer the *Error returned by the call that failed.
func LastError() string {
	return C.GoString(C.suzume_last_error())
}

// LastErrorCode returns the code of the last C API error for the current OS
// thread, or ErrorCodeSuccess when none is set. It carries the same
// thread-affinity caveat as LastError.
func LastErrorCode() ErrorCode {
	return ErrorCode(C.suzume_last_error_code())
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
