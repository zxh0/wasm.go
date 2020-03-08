#!/usr/bin/env bash

G4=./text/grammar/WAST.g4
OUT=./text/parser
rm ${OUT}/*
antlr -Dlanguage=Go -Xexact-output-dir -visitor -no-listener -o ${OUT} ${G4}