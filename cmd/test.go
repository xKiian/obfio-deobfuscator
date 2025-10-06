package main

import (
	"os"

	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
	deobf "github.com/xkiian/obfio-deobfuscator"
)

func main() {
	file, err := os.ReadFile("cmd/example.js")
	if err != nil {
		panic(err)
	}
	src := string(file)

	ast, err := parser.ParseFile(src)
	if err != nil {
		panic(err)
	}

	deobf.Deobfuscate(ast)

	os.WriteFile("out.js", []byte(generator.Generate(ast)), 0644)
}
