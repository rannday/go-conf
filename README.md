# go-config

Tiny config loading helper for Go.

This package loads TOML into caller-provided structs. Callers own defaults,
validation, merge logic, and overrides. TOML is the only supported format for
now, selected by the `.toml` file extension.

## Install

```sh
go get github.com/rannday/go-config
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/rannday/go-config"
)

type AppConfig struct {
	Name    string `toml:"name"`
	Port    *int   `toml:"port"`
	Enabled *bool  `toml:"enabled"`
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

## API

```go
func LoadFile[T any](path string) (T, error)
func LoadOptionalFile[T any](path string) (T, bool, error)
```

Missing optional files are not errors. Empty paths are rejected. Unsupported
file extensions return decode errors.
