package confx

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type testConfig struct {
	Name    string   `toml:"name"`
	Port    *int     `toml:"port"`
	Enabled *bool    `toml:"enabled"`
	Tags    []string `toml:"tags"`
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func TestLoadFile(t *testing.T) {
	tests := []struct {
		name string
		toml string
		want testConfig
	}{
		{
			name: "valid TOML",
			toml: `
name = "example"
port = 8080
enabled = true
tags = ["api", "worker"]
`,
			want: testConfig{
				Name:    "example",
				Port:    ptr(8080),
				Enabled: ptr(true),
				Tags:    []string{"api", "worker"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.toml")
			writeFile(t, path, tt.toml)

			got, err := LoadFile[testConfig](path)
			if err != nil {
				t.Fatalf("LoadFile returned error: %v", err)
			}

			assertConfig(t, got, tt.want)
		})
	}
}

func TestLoadOptionalFile(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantLoaded bool
		want       testConfig
	}{
		{
			name: "missing file",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "missing.toml")
			},
		},
		{
			name: "existing file",
			setup: func(t *testing.T) string {
				t.Helper()
				path := filepath.Join(t.TempDir(), "config.toml")
				writeFile(t, path, `
name = "optional"
port = 3000
enabled = false
tags = ["one", "two"]
`)
				return path
			},
			wantLoaded: true,
			want: testConfig{
				Name:    "optional",
				Port:    ptr(3000),
				Enabled: ptr(false),
				Tags:    []string{"one", "two"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, loaded, err := LoadOptionalFile[testConfig](tt.setup(t))
			if err != nil {
				t.Fatalf("LoadOptionalFile returned error: %v", err)
			}
			if loaded != tt.wantLoaded {
				t.Fatalf("loaded = %v, want %v", loaded, tt.wantLoaded)
			}

			assertConfig(t, got, tt.want)
		})
	}
}

func TestDecodeErrorsIncludeSource(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		source string
	}{
		{
			name:   "invalid TOML",
			data:   "name =",
			source: "config.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decode[testConfig]([]byte(tt.data), tt.source)
			if err == nil {
				t.Fatal("Decode returned nil error")
			}
			if !strings.Contains(err.Error(), `decode "`+tt.source+`"`) {
				t.Fatalf("error = %q, want decode source", err)
			}
		})
	}
}

func TestReadErrorsIncludePath(t *testing.T) {
	tests := []struct {
		name string
		path func(t *testing.T) string
	}{
		{
			name: "directory path",
			path: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path(t)

			_, err := LoadFile[testConfig](path)
			if err == nil {
				t.Fatal("LoadFile returned nil error")
			}
			if !strings.Contains(err.Error(), `read "`+path+`"`) {
				t.Fatalf("error = %q, want read path", err)
			}
		})
	}
}

func TestReadErrorsIncludeEmptyPath(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "empty path",
		},
		{
			name: "whitespace path",
			path: " \t ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadFile[testConfig](tt.path)
			if err == nil {
				t.Fatal("LoadFile returned nil error")
			}
			if !strings.Contains(err.Error(), "read "+strconv.Quote(tt.path)) {
				t.Fatalf("error = %q, want read path", err)
			}
			if !strings.Contains(err.Error(), "invalid path: empty") {
				t.Fatalf("error = %q, want empty path reason", err)
			}
		})
	}
}

func TestPointerFields(t *testing.T) {
	tests := []struct {
		name string
		toml string
		want testConfig
	}{
		{
			name: "pointers set when present",
			toml: `
port = 9000
enabled = true
`,
			want: testConfig{
				Port:    ptr(9000),
				Enabled: ptr(true),
			},
		},
		{
			name: "pointers nil when omitted",
			toml: `name = "minimal"`,
			want: testConfig{
				Name: "minimal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decode[testConfig]([]byte(tt.toml), tt.name+".toml")
			if err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}

			assertConfig(t, got, tt.want)
		})
	}
}

func TestSliceFields(t *testing.T) {
	tests := []struct {
		name string
		toml string
		want []string
	}{
		{
			name: "string slice",
			toml: `tags = ["alpha", "beta", "stable"]`,
			want: []string{"alpha", "beta", "stable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decode[testConfig]([]byte(tt.toml), tt.name+".toml")
			if err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}

			assertStringSlice(t, got.Tags, tt.want)
		})
	}
}

func TestDecodeTOML(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want testConfig
	}{
		{
			name: "bytes only",
			data: []byte(`
name = "decoded"
tags = ["memory"]
`),
			want: testConfig{
				Name: "decoded",
				Tags: []string{"memory"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decode[testConfig](tt.data, tt.name+".toml")
			if err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}

			assertConfig(t, got, tt.want)
		})
	}
}

func TestUnsupportedFormatsFailDecode(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		data     string
	}{
		{
			name:     "JSON",
			filename: "config.json",
			data:     `{"name":"example"}`,
		},
		{
			name:     "YAML",
			filename: "config.yaml",
			data:     `name: example`,
		},
		{
			name:     "env",
			filename: ".env",
			data:     `NAME=example`,
		},
		{
			name:     "missing extension",
			filename: "config",
			data:     `name = "example"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.filename)
			writeFile(t, path, tt.data)

			_, err := LoadFile[testConfig](path)
			if err == nil {
				t.Fatal("LoadFile returned nil error")
			}
			if !strings.Contains(err.Error(), `decode "`+path+`"`) {
				t.Fatalf("error = %q, want decode source", err)
			}
			if !strings.Contains(err.Error(), "unsupported config format") {
				t.Fatalf("error = %q, want unsupported format", err)
			}
		})
	}
}

func ptr[T any](value T) *T {
	return &value
}

func assertConfig(t *testing.T, got, want testConfig) {
	t.Helper()

	if got.Name != want.Name {
		t.Fatalf("Name = %q, want %q", got.Name, want.Name)
	}
	assertPointer(t, "Port", got.Port, want.Port)
	assertPointer(t, "Enabled", got.Enabled, want.Enabled)
	assertStringSlice(t, got.Tags, want.Tags)
}

func assertPointer[T comparable](t *testing.T, name string, got, want *T) {
	t.Helper()

	switch {
	case got == nil && want == nil:
		return
	case got == nil || want == nil:
		t.Fatalf("%s = %v, want %v", name, got, want)
	case *got != *want:
		t.Fatalf("%s = %v, want %v", name, *got, *want)
	}
}

func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(slice) = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("slice[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
