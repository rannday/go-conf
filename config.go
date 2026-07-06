// Package confx provides small helpers for loading configuration files into
// caller-defined structs. Supported formats are selected by file extension:
// .toml, .json, .jsonc, .yaml, .yml, and .ini.
package confx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/tailscale/hujson"
	"go.yaml.in/yaml/v4"
	"gopkg.in/ini.v1"
)

var (
	// ErrEmptyPath is returned when a path is empty or only whitespace.
	ErrEmptyPath = errors.New("invalid path: empty")

	// ErrUnsupportedFormat is returned when the file extension is not supported.
	ErrUnsupportedFormat = errors.New("unsupported config format")
)

type decoder func(data []byte, value any) error
type validator func(data []byte) error

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

	return Decode[T](data, path)
}

// LoadFileFS reads a config file from fsys and decodes it into T.
func LoadFileFS[T any](fsys fs.FS, path string) (T, error) {
	var zero T
	if err := validatePath(path); err != nil {
		return zero, fmt.Errorf("read %q: %w", path, err)
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return zero, fmt.Errorf("read %q: %w", path, err)
	}

	return Decode[T](data, path)
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
		if isNotExist(err) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("read %q: %w", path, err)
	}

	value, err := Decode[T](data, path)
	if err != nil {
		return zero, false, err
	}

	return value, true, nil
}

// LoadOptionalFileFS reads a config file from fsys and decodes it into T.
//
// If the file does not exist, it returns the zero value of T, false, nil.
func LoadOptionalFileFS[T any](fsys fs.FS, path string) (T, bool, error) {
	var zero T
	if err := validatePath(path); err != nil {
		return zero, false, fmt.Errorf("read %q: %w", path, err)
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		if isNotExist(err) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("read %q: %w", path, err)
	}

	value, err := Decode[T](data, path)
	if err != nil {
		return zero, false, err
	}

	return value, true, nil
}

// ValidateFile checks that path contains syntactically valid data for its file
// extension. It does not decode into caller-defined structs.
func ValidateFile(path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("validate %q: %w", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("validate %q: %w", path, err)
	}

	return Validate(data, path)
}

// ValidateFileFS checks that path in fsys contains syntactically valid data for
// its file extension. It does not decode into caller-defined structs.
func ValidateFileFS(fsys fs.FS, path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("validate %q: %w", path, err)
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return fmt.Errorf("validate %q: %w", path, err)
	}

	return Validate(data, path)
}

// Decode decodes data into T. The source name must include a supported file
// extension, which selects the decoder.
func Decode[T any](data []byte, source string) (T, error) {
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

// Validate checks that data is syntactically valid for the format implied by
// source's file extension.
func Validate(data []byte, source string) error {
	validator, err := validatorForSource(source)
	if err != nil {
		return fmt.Errorf("validate %q: %w", source, err)
	}

	if err := validator(data); err != nil {
		return fmt.Errorf("validate %q: %w", source, err)
	}

	return nil
}

func validatePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return ErrEmptyPath
	}
	return nil
}

func isNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, os.ErrNotExist)
}

func decoderForSource(source string) (decoder, error) {
	format, err := formatForSource(source)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatTOML:
		return decodeTOML, nil
	case FormatJSON:
		return decodeJSON, nil
	case FormatJSONC:
		return decodeJSONC, nil
	case FormatYAML:
		return decodeYAML, nil
	case FormatINI:
		return decodeINI, nil
	default:
		return nil, fmt.Errorf("unsupported config format for %q: %w", source, ErrUnsupportedFormat)
	}
}

func validatorForSource(source string) (validator, error) {
	format, err := formatForSource(source)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatTOML:
		return validateTOML, nil
	case FormatJSON:
		return validateJSON, nil
	case FormatJSONC:
		return validateJSONC, nil
	case FormatYAML:
		return validateYAML, nil
	case FormatINI:
		return validateINI, nil
	default:
		return nil, fmt.Errorf("unsupported config format for %q: %w", source, ErrUnsupportedFormat)
	}
}

func decodeTOML(data []byte, value any) error {
	return toml.Unmarshal(data, value)
}

func decodeJSON(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

func decodeJSONC(data []byte, value any) error {
	standardized, err := hujson.Standardize(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(standardized, value)
}

func decodeYAML(data []byte, value any) error {
	return yaml.Unmarshal(data, value)
}

func decodeINI(data []byte, value any) error {
	return ini.MapTo(value, data)
}

func validateTOML(data []byte) error {
	var value any
	return toml.Unmarshal(data, &value)
}

func validateJSON(data []byte) error {
	if !json.Valid(data) {
		return errors.New("invalid JSON syntax")
	}
	return nil
}

func validateJSONC(data []byte) error {
	_, err := hujson.Parse(data)
	return err
}

func validateYAML(data []byte) error {
	var value any
	return yaml.Unmarshal(data, &value)
}

func validateINI(data []byte) error {
	_, err := ini.Load(data)
	return err
}