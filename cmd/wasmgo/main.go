package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/zxh0/wasm.go/aot"
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
	"github.com/zxh0/wasm.go/interpreter"
	"github.com/zxh0/wasm.go/jit"
	"github.com/zxh0/wasm.go/spectest"
	"github.com/zxh0/wasm.go/text"
	"github.com/zxh0/wasm.go/validator"
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
	flagNameAOT     = "aot"
	flagNameCheck   = "check"
	flagNameCompile = "compile"
	flagNameDump    = "dump"
	flagNameExec    = "exec"
	flagNameLLVM    = "llvm"
	flagNameTest    = "test"
)

// wasmgo             file.wasm # exec
// wasmgo -A|-aot     file.wasm
// wasmgo -C|-check   file.wasm
// wasmgo -D|-dump    file.wasm
// wasmgo -K|-compile file.wat
// wasmgo -T|-test    file.wast
func main() {
	app := &cli.App{
		Version:   "0.1.0",
		Usage:     "Wasm.go CLI",
		ArgsUsage: "[file]",
		Flags: []cli.Flag{
			boolFlag(flagNameAOT, "A", "compile .wasm file to Go plugin", false),
			boolFlag(flagNameCheck, "C", "check .wasm file", false),
			boolFlag(flagNameDump, "D", "dump .wasm file", false),
			boolFlag(flagNameExec, "E", "execute .wasm file", true),
			boolFlag(flagNameCompile, "K", "compile .wat file", false),
			boolFlag(flagNameLLVM, "L", "compile .wasm file to LLVM IR", false),
			boolFlag(flagNameTest, "T", "test .wast file", false),
		},
		CustomAppHelpTemplate: appHelpTemplate,
		Action: func(ctx *cli.Context) error {
			filename := ctx.Args().Get(0)
			if ctx.Bool(flagNameAOT) {
				return compileWasmToGo(filename)
			} else if ctx.Bool(flagNameCheck) {
				return checkWasm(filename)
			} else if ctx.Bool(flagNameDump) {
				return dumpWasm(filename)
			} else if ctx.Bool(flagNameCompile) {
				return compileWatToWasm(filename)
			} else if ctx.Bool(flagNameLLVM) {
				return compileWasmToLLVM(filename)
			} else if ctx.Bool(flagNameTest) {
				return testWast(filename)
			} else {
				return execFile(filename)
			}
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func boolFlag(name, alias, usage string, value bool) cli.Flag {
	return &cli.BoolFlag{
		Name:    name,
		Aliases: []string{alias},
		Usage:   usage,
		Value:   value,
	}
}

func checkWasm(filename string) error {
	fmt.Println("check " + filename)
	module, err := binary.DecodeFile(filename)
	if err != nil {
		return err
	}
	return validator.Validate(module)
}

func dumpWasm(filename string) error {
	fmt.Printf("file: \n  %s\n\n", filename)

	module, err := binary.DecodeFile(filename)
	if err != nil {
		return err
	}

	dump(module)
	return nil
}

func testWast(filename string) error {
	fmt.Println("test " + filename)
	s, err := text.CompileScriptFile(filename)
	if err != nil {
		return err
	}
	return spectest.TestWast(s)
}

func compileWatToWasm(filename string) error {
	fmt.Println("compile " + filename)
	m, err := text.CompileModuleFile(filename)
	if err != nil {
		return err
	}

	// TODO
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}

func compileWasmToGo(filename string) error {
	//fmt.Println("AOT " + filename)
	if strings.HasSuffix(filename, ".wat") {
		if m, err := text.CompileModuleFile(filename); err != nil {
			return err
		} else {
			aot.Compile(*m)
			return nil
		}
	}
	if m, err := binary.DecodeFile(filename); err != nil {
		return err
	} else {
		aot.Compile(m)
	}
	return nil
}

func compileWasmToLLVM(filename string) error {
	module, err := binary.DecodeFile(filename)
	if err != nil {
		return err
	}
	// TODO
	jit.Compile(module)
	return nil
}

func execFile(filename string) error {
	if strings.HasSuffix(filename, ".wat") {
		return execWat(filename)
	}
	if strings.HasSuffix(filename, ".wasm") {
		return execWasm(filename)
	}
	if strings.HasSuffix(filename, ".so") {
		return execSO(filename)
	}
	fmt.Println("unknown file format: " + filename)
	return nil
}

func execWat(filename string) error {
	//fmt.Println("exec " + filename)
	m, err := text.CompileModuleFile(filename)
	if err != nil {
		return err
	}

	iMap := map[string]instance.Instance{"env": newTestEnv()}
	vm, err := interpreter.NewInstance(*m, iMap)
	if err != nil {
		return err
	}

	_, err = vm.CallFunc("main")
	return err
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

	iMap := map[string]instance.Instance{"env": newTestEnv()}
	vm, err := interpreter.NewInstance(module, iMap)
	if err != nil {
		return err
	}

	_, err = vm.CallFunc("main")
	return err
}

func execSO(filename string) error {
	//fmt.Println("exec " + filename)
	iMap := map[string]instance.Instance{"env": newTestEnv()}
	i, err := aot.Load(filename, iMap)
	if err != nil {
		return err
	}

	_, err = i.CallFunc("main")
	return err
}
