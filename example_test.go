package confx_test

import (
	"fmt"
	"os"

	confx "github.com/rannday/go-config"
)

type appConfig struct {
	Name string `toml:"name" json:"name" yaml:"name" ini:"name"`
	Port *int   `toml:"port" json:"port" yaml:"port" ini:"port"`
}

func ExampleLoadFile() {
	cfg, err := confx.LoadFile[appConfig]("testdata/app.toml")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output: demo
}

func ExampleLoadFileFS() {
	cfg, err := confx.LoadFileFS[appConfig](os.DirFS("testdata"), "app.json")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output: demo
}

func ExampleLoadOptionalFile() {
	cfg, loaded, err := confx.LoadOptionalFile[appConfig]("testdata/missing.toml")
	if err != nil {
		panic(err)
	}
	if !loaded {
		fmt.Println("not loaded")
		return
	}

	fmt.Println(cfg.Name)
	// Output: not loaded
}

func ExampleDecode_toml() {
	cfg, err := confx.Decode[appConfig]([]byte(`name = "demo"`), "config.toml")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output: demo
}

func ExampleDecode_json() {
	cfg, err := confx.Decode[appConfig]([]byte(`{"name":"demo"}`), "config.json")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output: demo
}

func ExampleDecode_jsonc() {
	cfg, err := confx.Decode[appConfig]([]byte(`{
  // jsonc comment
  "name": "demo",
}`), "config.jsonc")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output: demo
}

func ExampleDecode_yaml() {
	cfg, err := confx.Decode[appConfig]([]byte("name: demo\n"), "config.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output: demo
}

func ExampleDecode_ini() {
	cfg, err := confx.Decode[appConfig]([]byte("name = demo\n"), "config.ini")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output: demo
}

func ExampleDecode_iniSection() {
	type serverConfig struct {
		Server struct {
			Name string `ini:"name"`
			Port *int   `ini:"port"`
		} `ini:"server"`
	}

	cfg, err := confx.Decode[serverConfig]([]byte(`
[server]
name = demo
port = 8080
`), "config.ini")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Server.Name)
	// Output: demo
}

func ExampleValidate() {
	err := confx.Validate([]byte(`{"name":"demo"}`), "config.json")
	if err != nil {
		panic(err)
	}

	fmt.Println("valid")
	// Output: valid
}

func ExampleFormatForPath() {
	format, err := confx.FormatForPath("config.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println(format)
	// Output: yaml
}

func ExampleSupportedFormats() {
	for _, format := range confx.SupportedFormats() {
		fmt.Println(format)
	}
	// Output:
	// toml
	// json
	// jsonc
	// yaml
	// ini
}