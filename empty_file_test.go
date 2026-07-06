package confx

import (
	"path/filepath"
	"testing"
)

func TestEmptyFileDecode(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		wantDecode bool
	}{
		{name: "TOML", filename: "empty.toml", wantDecode: true},
		{name: "JSON", filename: "empty.json", wantDecode: false},
		{name: "JSONC", filename: "empty.jsonc", wantDecode: false},
		{name: "YAML", filename: "empty.yaml", wantDecode: true},
		{name: "YML", filename: "empty.yml", wantDecode: true},
		{name: "INI", filename: "empty.ini", wantDecode: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.filename)
			writeFile(t, path, "")

			_, err := LoadFile[testConfig](path)
			if tt.wantDecode && err != nil {
				t.Fatalf("LoadFile returned error: %v", err)
			}
			if !tt.wantDecode && err == nil {
				t.Fatal("LoadFile returned nil error")
			}
		})
	}
}

func TestEmptyFileDecodeReturnsZeroStruct(t *testing.T) {
	tests := []string{"empty.toml", "empty.yaml", "empty.ini"}

	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), filename)
			writeFile(t, path, "")

			got, err := LoadFile[testConfig](path)
			if err != nil {
				t.Fatalf("LoadFile returned error: %v", err)
			}
			if got.Name != "" || got.Port != nil || got.Enabled != nil || len(got.Tags) != 0 {
				t.Fatalf("got = %#v, want zero struct", got)
			}
		})
	}
}

func TestEmptyFileValidate(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		wantValidate bool
	}{
		{name: "TOML", filename: "empty.toml", wantValidate: true},
		{name: "JSON", filename: "empty.json", wantValidate: false},
		{name: "JSONC", filename: "empty.jsonc", wantValidate: false},
		{name: "YAML", filename: "empty.yaml", wantValidate: true},
		{name: "INI", filename: "empty.ini", wantValidate: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.filename)
			writeFile(t, path, "")

			err := ValidateFile(path)
			if tt.wantValidate && err != nil {
				t.Fatalf("ValidateFile returned error: %v", err)
			}
			if !tt.wantValidate && err == nil {
				t.Fatal("ValidateFile returned nil error")
			}
		})
	}
}