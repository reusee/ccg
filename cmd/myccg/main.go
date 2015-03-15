package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"go/ast"
	"go/parser"
	"go/token"

	"github.com/reusee/ccg"
)

var (
	pt = fmt.Printf

	outputFile  = flag.String("output", "", "output file")
	packageName = flag.String("package", "", "output package")
)

func init() {
	flag.Parse()
}

func main() {
	args := flag.Args()
	if len(args) < 1 {
		pt("usage: %s [command] [args...]\n", os.Args[0])
		return
	}

	type CmdSpec struct {
		Usage   string
		Params  []string
		Renames []string
	}

	specs := map[string]CmdSpec{
		"sorter": {
			"[element type] [sorter name]",
			[]string{"T"},
			[]string{"Sorter"},
		},
		"set": {
			"[element type] [set type] [constructor name]",
			[]string{"T"},
			[]string{"Set", "New"},
		},
		"infchan": {
			"[element type]",
			[]string{"T"},
			[]string{"New"},
		},
		"slice": {
			"[element type] [slice type]",
			[]string{"T"},
			[]string{"Ts"},
		},
	}

	spec, ok := specs[args[0]]
	if !ok {
		log.Fatalf("unknown subcommand %s", args[0])
	}
	if len(args[1:]) != len(spec.Params)+len(spec.Renames) {
		log.Fatalf("usage: %s %s %s", os.Args[0], args[0], spec.Usage)
	}
	params := map[string]string{}
	for i, param := range spec.Params {
		params[param] = args[1+i]
	}
	renames := map[string]string{}
	start := len(spec.Params)
	for i, orig := range spec.Renames {
		renames[orig] = args[1+start+i]
	}

	buf := new(bytes.Buffer)
	var decls []ast.Decl
	if *outputFile != "" {
		content, err := ioutil.ReadFile(*outputFile)
		if err == nil {
			astFile, err := parser.ParseFile(new(token.FileSet), *outputFile, content, 0)
			if err == nil {
				decls = astFile.Decls
			}
		}
		if *packageName == "" {
			main := "main"
			packageName = &main
		}
	}

	err := ccg.Copy(ccg.Config{
		From:    "github.com/reusee/ccg/" + args[0],
		Params:  params,
		Renames: renames,
		Writer:  buf,
		Package: *packageName,
		Decls:   decls,
	})
	if err != nil {
		log.Fatalf("ccg: copy error %v", err)
	}
	if *outputFile == "" {
		pt("%s\n", buf.Bytes())
	} else {
		err = ioutil.WriteFile(*outputFile, buf.Bytes(), 0644)
		if err != nil {
			log.Fatalf("ccg: write file error %v", err)
		}
	}
}
