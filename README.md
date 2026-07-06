# go-config

Tiny config loading helper for Go.

This package loads TOML, JSON, JSONC, YAML, and INI into caller-provided
structs. Callers own defaults, validation, merge logic, and overrides. Format
is selected by file extension: `.toml`, `.json`, `.jsonc`, `.yaml`, `.yml`,
and `.ini`.

The module path is `github.com/rannday/go-config`, but the package name is
`confx`. Import it with an alias:

```go
import confx "github.com/rannday/go-config"
```

## Install

```sh
go get github.com/rannday/go-config@latest
```

## Usage

```go
package main

import (
	"fmt"

	confx "github.com/rannday/go-config"
)

type AppConfig struct {
	Name    string `toml:"name" json:"name" yaml:"name" ini:"name"`
	Port    *int   `toml:"port" json:"port" yaml:"port" ini:"port"`
	Enabled *bool  `toml:"enabled" json:"enabled" yaml:"enabled" ini:"enabled"`
}

func main() {
	cfg, loaded, err := confx.LoadOptionalFile[AppConfig]("config.toml")
	if err != nil {
		panic(err)
	}

	if !loaded {
		fmt.Println("no config file found")
		return
	}

	if cfg.Port != nil {
		fmt.Println("port:", *cfg.Port)
	}
}
```

### Embedded configs

```go
//go:embed config.toml
var configFS embed.FS

cfg, err := confx.LoadFileFS[AppConfig](configFS, "config.toml")
```

### Bytes without a real file

```go
cfg, err := confx.Decode[AppConfig](data, "config.json")
err = confx.Validate(data, "config.json")
```

The `source` argument must include a supported extension so the decoder can be
selected. It is also included in decode and validate errors.

## API

```go
func LoadFile[T any](path string) (T, error)
func LoadFileFS[T any](fsys fs.FS, path string) (T, error)
func LoadOptionalFile[T any](path string) (T, bool, error)
func LoadOptionalFileFS[T any](fsys fs.FS, path string) (T, bool, error)
func Decode[T any](data []byte, source string) (T, error)
func ValidateFile(path string) error
func ValidateFileFS(fsys fs.FS, path string) error
func Validate(data []byte, source string) error
func FormatForPath(path string) (Format, error)
func SupportedFormats() []Format
```

```go
var ErrEmptyPath = errors.New("invalid path: empty")
var ErrUnsupportedFormat = errors.New("unsupported config format")
```

Missing optional files are not errors. Empty paths are rejected with
`ErrEmptyPath`. Unsupported file extensions return errors wrapping
`ErrUnsupportedFormat`. `ValidateFile` and `Validate` check syntax only and do
not decode into caller-defined structs.

## Struct tags

Each decoder reads its own struct tag:

| Format | Tag |
|--------|-----|
| TOML | `toml:"field"` |
| JSON / JSONC | `json:"field"` |
| YAML | `yaml:"field"` |
| INI | `ini:"field"` |

If you load more than one format into the same struct, include the tags for
every format you plan to use.

Use pointer fields such as `*int` and `*bool` when you need to tell the
difference between a field that was omitted and one set to a zero value.

## Format notes

- `.json` expects standard JSON.
- `.jsonc` allows comments and trailing commas via JWCC parsing.
- `.yaml` and `.yml` use YAML 1.2 parsing. Valid JSON is often also valid YAML.
- `.ini` maps keys from the default section unless your struct is set up for
  named sections:

```go
type Config struct {
    Server struct {
        Name string `ini:"name"`
    } `ini:"server"`
}
```

## Edge cases

| Case | Behavior |
|------|----------|
| Missing file with `LoadOptionalFile` | zero value, `loaded == false`, `nil` error |
| Missing file with `LoadFile` | read error |
| Empty path | `ErrEmptyPath` |
| Unsupported extension | error wrapping `ErrUnsupportedFormat` |
| Wrong syntax for extension | decode/validate error from the selected parser |

### Empty files

| Extension | `LoadFile` / `Decode` | `ValidateFile` / `Validate` |
|-----------|----------------------|----------------------------|
| `.toml` | zero struct, no error | no error |
| `.json` | error | error |
| `.jsonc` | error | error |
| `.yaml`, `.yml` | zero struct, no error | no error |
| `.ini` | zero struct, no error | no error |

An empty TOML, YAML, or INI file therefore loads as the zero value of your
struct. An empty JSON or JSONC file is rejected.