package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/zxh0/wasm.go/aot"
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
	"github.com/zxh0/wasm.go/interpreter"
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
	flagNameAOT   = "aot"
	flagNameCheck = "check"
	flagNameDump  = "dump"
	flagNameExec  = "exec"
)

// wasmgo    file.wasm # exec
// wasmgo -d file.wasm # dump
// wasmgo -c file.wasm # check
func main() {
	app := &cli.App{
		Version:   "0.1.0",
		Usage:     "Wasm.go CLI",
		ArgsUsage: "[file]",
		Flags: []cli.Flag{
			boolFlag(flagNameAOT, "aot compile wasm file", false),
			boolFlag(flagNameCheck, "check wasm file", false),
			boolFlag(flagNameDump, "dump wasm file", false),
			boolFlag(flagNameExec, "execute wasm file", true),
		},
		CustomAppHelpTemplate: appHelpTemplate,
		Action: func(ctx *cli.Context) error {
			filename := ctx.Args().Get(0)
			if ctx.Bool(flagNameAOT) {
				return aotWasm(filename)
			} else if ctx.Bool(flagNameCheck) {
				return checkWasm(filename)
			} else if ctx.Bool(flagNameDump) {
				return dumpWasm(filename)
			} else {
				return execWasm(filename)
			}
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func boolFlag(name, usage string, value bool) cli.Flag {
	return &cli.BoolFlag{
		Name:    name,
		Aliases: []string{usage[0:1]},
		Usage:   usage,
		Value:   value,
	}
}

func aotWasm(filename string) error {
	module, err := binary.DecodeFile(filename)
	if err != nil {
		return err
	}
	// TODO
	aot.Compile(module)
	return nil
}

func checkWasm(filename string) error {
	fmt.Println("check " + filename)
	return nil
}

func dumpWasm(filename string) error {
	fmt.Printf("file: \n  %s\n\n", filename)

	module, err := binary.DecodeFile(filename)
	if err != nil {
		return err
	}

	newDumper(module).dump()
	return nil
}

func execWasm(filename string) error {
	//fmt.Println("exec " + filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	module, err := binary.Decode(data)
	if err != nil {
		return err
	}

	ni := &NativeInstance{}
	mm := map[string]instance.Instance{"env": ni}
	vm, err := interpreter.NewInstance(module, mm)
	if err != nil {
		return err
	}

	//ni.mem, _ = vm.GetMemory("")
	_, err = vm.CallFunc("main")
	return err
}
