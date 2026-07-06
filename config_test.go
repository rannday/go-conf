package confx

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type testConfig struct {
	Name    string   `toml:"name" json:"name" yaml:"name" ini:"name"`
	Port    *int     `toml:"port" json:"port" yaml:"port" ini:"port"`
	Enabled *bool    `toml:"enabled" json:"enabled" yaml:"enabled" ini:"enabled"`
	Tags    []string `toml:"tags" json:"tags" yaml:"tags" ini:"tags"`
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
			_, err := Decode[testConfig]([]byte(tt.data), tt.source)
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
			if !errors.Is(err, ErrEmptyPath) {
				t.Fatalf("error = %q, want ErrEmptyPath", err)
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
			got, err := Decode[testConfig]([]byte(tt.toml), tt.name+".toml")
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
			got, err := Decode[testConfig]([]byte(tt.toml), tt.name+".toml")
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
			got, err := Decode[testConfig](tt.data, tt.name+".toml")
			if err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}

			assertConfig(t, got, tt.want)
		})
	}
}

func TestLoadFileJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeFile(t, path, `{
  "name": "example",
  "port": 8080,
  "enabled": true,
  "tags": ["api", "worker"]
}`)

	got, err := LoadFile[testConfig](path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	assertConfig(t, got, testConfig{
		Name:    "example",
		Port:    ptr(8080),
		Enabled: ptr(true),
		Tags:    []string{"api", "worker"},
	})
}

func TestLoadFileJSONC(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.jsonc")
	writeFile(t, path, `{
  // app settings
  "name": "example",
  "port": 8080,
  "enabled": true,
  "tags": ["api", "worker",], /* trailing comma ok */
}`)

	got, err := LoadFile[testConfig](path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	assertConfig(t, got, testConfig{
		Name:    "example",
		Port:    ptr(8080),
		Enabled: ptr(true),
		Tags:    []string{"api", "worker"},
	})
}

func TestLoadFileYAML(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: "yaml extension", filename: "config.yaml"},
		{name: "yml extension", filename: "config.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.filename)
			writeFile(t, path, `
name: example
port: 8080
enabled: true
tags:
  - api
  - worker
`)

			got, err := LoadFile[testConfig](path)
			if err != nil {
				t.Fatalf("LoadFile returned error: %v", err)
			}

			assertConfig(t, got, testConfig{
				Name:    "example",
				Port:    ptr(8080),
				Enabled: ptr(true),
				Tags:    []string{"api", "worker"},
			})
		})
	}
}

func TestLoadFileINI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.ini")
	writeFile(t, path, `
name = example
port = 8080
enabled = true
tags = api, worker
`)

	got, err := LoadFile[testConfig](path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	assertConfig(t, got, testConfig{
		Name:    "example",
		Port:    ptr(8080),
		Enabled: ptr(true),
		Tags:    []string{"api", "worker"},
	})
}

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		data     string
	}{
		{name: "valid TOML", filename: "config.toml", data: `name = "ok"`},
		{name: "valid JSON", filename: "config.json", data: `{"name":"ok"}`},
		{name: "valid JSONC", filename: "config.jsonc", data: `{"name":"ok", /* comment */}`},
		{name: "valid YAML", filename: "config.yaml", data: "name: ok\n"},
		{name: "valid INI", filename: "config.ini", data: "name = ok\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.filename)
			writeFile(t, path, tt.data)

			if err := ValidateFile(path); err != nil {
				t.Fatalf("ValidateFile returned error: %v", err)
			}
		})
	}
}

func TestValidateFileRejectsInvalidSyntax(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		data     string
	}{
		{name: "invalid TOML", filename: "config.toml", data: "name ="},
		{name: "invalid JSON", filename: "config.json", data: `{"name":`},
		{name: "invalid JSONC", filename: "config.jsonc", data: `{"name": // broken`},
		{name: "invalid YAML", filename: "config.yaml", data: "name: [\n"},
		{name: "invalid INI", filename: "config.ini", data: "[unclosed\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.filename)
			writeFile(t, path, tt.data)

			err := ValidateFile(path)
			if err == nil {
				t.Fatal("ValidateFile returned nil error")
			}
			if !strings.Contains(err.Error(), `validate "`+path+`"`) {
				t.Fatalf("error = %q, want validate source", err)
			}
		})
	}
}

func TestWrongFormatContentFailsDecode(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		data     string
	}{
		{
			name:     "JSON content in TOML file",
			filename: "config.toml",
			data:     `{"name":"example"}`,
		},
		{
			name:     "TOML content in JSON file",
			filename: "config.json",
			data:     `name = "example"`,
		},
		{
			name:     "TOML content in JSONC file",
			filename: "config.jsonc",
			data:     `name = "example"`,
		},
		{
			name:     "TOML content in YAML file",
			filename: "config.yaml",
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
			if !errors.Is(err, ErrUnsupportedFormat) {
				t.Fatalf("error = %q, want ErrUnsupportedFormat", err)
			}
		})
	}
}

func TestLoadFileFS(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "config.toml"), `
name = "embedded"
port = 9090
`)

	got, err := LoadFileFS[testConfig](os.DirFS(dir), "config.toml")
	if err != nil {
		t.Fatalf("LoadFileFS returned error: %v", err)
	}

	assertConfig(t, got, testConfig{
		Name: "embedded",
		Port: ptr(9090),
	})
}

func TestLoadOptionalFileFS(t *testing.T) {
	dir := t.TempDir()

	got, loaded, err := LoadOptionalFileFS[testConfig](os.DirFS(dir), "missing.toml")
	if err != nil {
		t.Fatalf("LoadOptionalFileFS returned error: %v", err)
	}
	if loaded {
		t.Fatal("loaded = true, want false")
	}
	if got.Name != "" || got.Port != nil || got.Enabled != nil || len(got.Tags) != 0 {
		t.Fatalf("got = %#v, want zero value", got)
	}

	writeFile(t, filepath.Join(dir, "config.json"), `{"name":"fs","port":7070}`)
	got, loaded, err = LoadOptionalFileFS[testConfig](os.DirFS(dir), "config.json")
	if err != nil {
		t.Fatalf("LoadOptionalFileFS returned error: %v", err)
	}
	if !loaded {
		t.Fatal("loaded = false, want true")
	}
	assertConfig(t, got, testConfig{
		Name: "fs",
		Port: ptr(7070),
	})
}

func TestValidateFileFS(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "config.yaml"), "name: ok\n")

	if err := ValidateFileFS(os.DirFS(dir), "config.yaml"); err != nil {
		t.Fatalf("ValidateFileFS returned error: %v", err)
	}
}

func TestValidateBytes(t *testing.T) {
	if err := Validate([]byte(`{"name":"ok"}`), "config.json"); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestExportedErrors(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		_, err := LoadFile[testConfig]("")
		if !errors.Is(err, ErrEmptyPath) {
			t.Fatalf("error = %v, want ErrEmptyPath", err)
		}
	})

	t.Run("unsupported format", func(t *testing.T) {
		_, err := Decode[testConfig]([]byte(`name = "x"`), "config.env")
		if !errors.Is(err, ErrUnsupportedFormat) {
			t.Fatalf("error = %v, want ErrUnsupportedFormat", err)
		}
	})

	t.Run("validate unsupported format", func(t *testing.T) {
		err := Validate([]byte("x"), "config")
		if !errors.Is(err, ErrUnsupportedFormat) {
			t.Fatalf("error = %v, want ErrUnsupportedFormat", err)
		}
	})
}

func TestLoadFileFSRejectsInvalidPath(t *testing.T) {
	_, err := LoadFileFS[testConfig](os.DirFS(t.TempDir()), " ")
	if !errors.Is(err, ErrEmptyPath) {
		t.Fatalf("error = %v, want ErrEmptyPath", err)
	}
}

func TestLoadFileFSPropagatesNotExist(t *testing.T) {
	_, err := LoadFileFS[testConfig](os.DirFS(t.TempDir()), "missing.toml")
	if err == nil {
		t.Fatal("LoadFileFS returned nil error")
	}
	if !strings.Contains(err.Error(), `read "missing.toml"`) {
		t.Fatalf("error = %q, want read path", err)
	}
	if !isNotExist(err) {
		t.Fatalf("error = %v, want not-exist in error chain", err)
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
