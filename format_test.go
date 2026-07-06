package confx

import (
	"errors"
	"testing"
)

func TestFormatForPath(t *testing.T) {
	tests := []struct {
		path string
		want Format
	}{
		{path: "config.toml", want: FormatTOML},
		{path: "config.JSON", want: FormatJSON},
		{path: "config.jsonc", want: FormatJSONC},
		{path: "config.yaml", want: FormatYAML},
		{path: "config.yml", want: FormatYAML},
		{path: "config.ini", want: FormatINI},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := FormatForPath(tt.path)
			if err != nil {
				t.Fatalf("FormatForPath returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("FormatForPath = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatForPathUnsupported(t *testing.T) {
	_, err := FormatForPath("config.env")
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestSupportedFormats(t *testing.T) {
	want := []Format{FormatTOML, FormatJSON, FormatJSONC, FormatYAML, FormatINI}
	got := SupportedFormats()
	if len(got) != len(want) {
		t.Fatalf("len(SupportedFormats()) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SupportedFormats()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}