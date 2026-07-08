package suzume

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"os"
	"path/filepath"
)

// dictFS carries the core and user dictionaries so a caller does not have to
// stage them on disk. The files are vendored by sync-upstream.sh alongside the
// C++ sources, matching what the upstream CLI auto-loads by default.
//
//go:embed csuzume/data/core.dic csuzume/data/user.dic
var dictFS embed.FS

// embeddedDicts lists the embedded dictionary files the analyzer auto-loads.
var embeddedDicts = []string{"core.dic", "user.dic"}

// init makes the embedded dictionaries discoverable by the C++ analyzer, which
// searches SUZUME_DATA_DIR before other locations at instance creation time.
// An explicit SUZUME_DATA_DIR is always respected, and any failure is
// non-fatal: the analyzer simply falls back to running without the core
// dictionary.
func init() {
	if os.Getenv("SUZUME_DATA_DIR") != "" {
		return
	}
	dir, err := materializeDicts()
	if err != nil {
		return
	}
	_ = os.Setenv("SUZUME_DATA_DIR", dir)
}

// materializeDicts writes the embedded dictionaries into a content-addressed
// cache directory and returns its path. The directory name derives from the
// dictionary contents, so a dictionary update lands in a fresh directory and
// stale files are never reused.
func materializeDicts() (string, error) {
	data := make(map[string][]byte, len(embeddedDicts))
	h := sha256.New()
	for _, name := range embeddedDicts {
		b, err := dictFS.ReadFile("csuzume/data/" + name)
		if err != nil {
			return "", err
		}
		data[name] = b
		h.Write([]byte(name))
		h.Write(b)
	}

	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "go-suzume", hex.EncodeToString(h.Sum(nil))[:16])
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	for name, b := range data {
		if err := writeIfChanged(filepath.Join(dir, name), b); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// writeIfChanged writes data to path only when the file is missing or differs,
// using a temp-file rename so concurrent processes never observe a partial
// file.
func writeIfChanged(path string, data []byte) error {
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, data) {
		return nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}
