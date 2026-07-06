package confx

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Format identifies a supported configuration format.
type Format string

const (
	FormatTOML  Format = "toml"
	FormatJSON  Format = "json"
	FormatJSONC Format = "jsonc"
	FormatYAML  Format = "yaml"
	FormatINI   Format = "ini"
)

// SupportedFormats returns the formats this package can load.
func SupportedFormats() []Format {
	return []Format{
		FormatTOML,
		FormatJSON,
		FormatJSONC,
		FormatYAML,
		FormatINI,
	}
}

// FormatForPath returns the format implied by path's file extension.
func FormatForPath(path string) (Format, error) {
	return formatForSource(path)
}

func formatForSource(source string) (Format, error) {
	ext := strings.ToLower(filepath.Ext(source))
	switch ext {
	case ".toml":
		return FormatTOML, nil
	case ".json":
		return FormatJSON, nil
	case ".jsonc":
		return FormatJSONC, nil
	case ".yaml", ".yml":
		return FormatYAML, nil
	case ".ini":
		return FormatINI, nil
	case "":
		return "", fmt.Errorf("unsupported config format for %q: missing file extension: %w", source, ErrUnsupportedFormat)
	default:
		return "", fmt.Errorf("unsupported config format for %q: %s: %w", source, ext, ErrUnsupportedFormat)
	}
}