#!/usr/bin/env bash
set -ex

go build github.com/zxh0/wasm.go/cmd/wasmgo

WAST_DIR=./spec/test/core
for f in $WAST_DIR/*.wast ; do
  # echo $f
  if [[ "$f" =~ "inline-module.wast" ]]; then
    echo "skip $f"
  else
    ./wasmgo -T $f > /dev/null
  fi
done

rm wasmgo
echo "OK!"
