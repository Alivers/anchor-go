package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/alivers/anchor-go/internal/generator"
	"github.com/alivers/anchor-go/internal/idl"
)

var src = flag.String("src", "", "Path to source; can use multiple times.")
var dst = flag.String("dst", "generated", "Destination folder, the program name will be the folder for the genrated files")
var generateTests = flag.Bool("tests", true, "Generate tests")

func main() {
	flag.Parse()

	idlFile, err := os.Open(*src)
	if err != nil {
		panic(err)
	}

	var program *idl.Idl
	dec := json.NewDecoder(idlFile)
	err = dec.Decode(&program)
	if err != nil {
		panic(err)
	}

	generator.Generate(*dst, *generateTests, program)
}
