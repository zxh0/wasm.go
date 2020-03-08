package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/zxh0/wasm.go/text"
)

const appHelpTemplate = `NAME:
   {{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} [options] {{.ArgsUsage}}{{end}}{{if .VisibleCommands}}

OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`

const (
	flagNameCompile = "compile"
	flagNameTest    = "test"
)

// wastgo --lex     file.wast # print tokens
// wastgo --cst     file.wast # print CST
// wastgo --compile file.wast # compile wast
// wastgo --test    file.wast # test wast
// wastgo           file.wast # test wast
func main() {
	app := &cli.App{
		Version:   "0.1.0",
		Usage:     "Wasm.go CLI",
		ArgsUsage: "[file]",
		Flags: []cli.Flag{
			boolFlag(flagNameCompile, "compile wast", false),
			boolFlag(flagNameTest, "test wast", true),
		},
		CustomAppHelpTemplate: appHelpTemplate,
		Action: func(ctx *cli.Context) error {
			filename := ctx.Args().Get(0)
			if ctx.Bool(flagNameCompile) {
				return compileWast(filename)
			} else {
				return testWast(filename)
			}
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func boolFlag(name, usage string, value bool) cli.Flag {
	return &cli.BoolFlag{
		Name:  name,
		Usage: usage,
		Value: value,
	}
}

func compileWast(filename string) error {
	s, err := text.CompileScriptFile(filename)
	if err != nil {
		return err
	}

	//bytes, err := json.Marshal(s)
	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}

func testWast(filename string) error {
	s, err := text.CompileScriptFile(filename)
	if err != nil {
		return err
	}
	return newWastTester(s).test()
}
