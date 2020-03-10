# wasm.go

[![Build Status](https://travis-ci.com/zxh0/wasm.go.svg?branch=master)](https://travis-ci.com/zxh0/wasm.go)

An experimental [Wasm](https://webassembly.org/) implementation written in Go.

![jaws](jaws.png)



## Features

* **binary**
  * **types** Go structs translated from Wasm binary format (as simple and direct as possible)
  * **decoder** Wasm binary format decoder
  * **encoder (TBD)** Wasm binary format encoder
* **validator** Wasm binary format validator
* **interpreter** Wasm interpreter 
* **text (WIP)** WAT & WAST compiler powered by [ANTLR](https://www.antlr.org/)
* **aot (WIP)** AOT (Wasm binary -> Go plugin) compiler

