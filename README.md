# wasm.go

[![Build Status](https://travis-ci.com/zxh0/wasm.go.svg?branch=master)](https://travis-ci.com/zxh0/wasm.go)

An experimental [Wasm](https://webassembly.org/) implementation written in Go.

![jaws](jaws.png)



## Features

* **binary**
  * **types** Go structs translated from Wasm binary format (as simple and direct as possible)
  * **decoder** Wasm binary format decoder
  * **encoder (WIP)** Wasm binary format encoder
* **validator** Wasm binary format validator
* **interpreter** Wasm interpreter 
* **text (WIP)** WAT & WAST compiler powered by [ANTLR](https://www.antlr.org/)
* **aot (WIP)** AOT (Wasm binary -> Go plugin) compiler



## Running "Hello, World!"

Interpreter mode:

```bash
$ git clone https://github.com/zxh0/wasm.go
$ cd wasm.go
$ go run github.com/zxh0/wasm.go/cmd/wasmgo hw.wat
```

AOT mode:

```bash
$ git clone https://github.com/zxh0/wasm.go
$ cd wasm.go
$ go run github.com/zxh0/wasm.go/cmd/wasmgo -aot hw.wat > hw.wasm.go
$ go build -buildmode=plugin -o hw.so hw.wasm.go
$ go run github.com/zxh0/wasm.go/cmd/wasmgo hw.so
```

