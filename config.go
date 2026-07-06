// Package confx provides small helpers for loading configuration files into
// caller-defined structs.
package confx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

var errEmptyPath = errors.New("invalid path: empty")

type decoder func(data []byte, value any) error

// LoadFile reads a config file and decodes it into T.
func LoadFile[T any](path string) (T, error) {
	var zero T
	if err := validatePath(path); err != nil {
		return zero, fmt.Errorf("read %q: %w", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return zero, fmt.Errorf("read %q: %w", path, err)
	}

	return decode[T](data, path)
}

// LoadOptionalFile reads a config file and decodes it into T.
//
// If the file does not exist, it returns the zero value of T, false, nil.
func LoadOptionalFile[T any](path string) (T, bool, error) {
	var zero T
	if err := validatePath(path); err != nil {
		return zero, false, fmt.Errorf("read %q: %w", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("read %q: %w", path, err)
	}

	value, err := decode[T](data, path)
	if err != nil {
		return zero, false, err
	}

	return value, true, nil
}

func validatePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errEmptyPath
	}
	return nil
}

func decode[T any](data []byte, source string) (T, error) {
	var value T
	decoder, err := decoderForSource(source)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("decode %q: %w", source, err)
	}

	if err := decoder(data, &value); err != nil {
		var zero T
		return zero, fmt.Errorf("decode %q: %w", source, err)
	}

	return value, nil
}

func decoderForSource(source string) (decoder, error) {
	ext := strings.ToLower(filepath.Ext(source))
	switch ext {
	case ".toml":
		return decodeTOML, nil
	case "":
		return nil, fmt.Errorf("unsupported config format for %q: missing file extension", source)
	default:
		return nil, fmt.Errorf("unsupported config format for %q: %s", source, ext)
	}
}

func decodeTOML(data []byte, value any) error {
	return toml.Unmarshal(data, value)
}
