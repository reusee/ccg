package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"go/ast"
	"go/parser"
	"go/token"

	"github.com/jessevdk/go-flags"
	"github.com/reusee/ccg"
)

var (
	pt = fmt.Printf
)

var opts struct {
	Output  string `short:"o" long:"output" description:"Output file path"`
	Package string `short:"p" long:"package" description:"Output package name"`
	Uses    string `short:"u" long:"uses" description:"comma-separated names to be used only"`
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}
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
	fileSet := new(token.FileSet)
	if opts.Output != "" {
		content, err := ioutil.ReadFile(opts.Output)
		if err == nil {
			astFile, err := parser.ParseFile(fileSet, opts.Output, content, 0)
			if err == nil {
				decls = astFile.Decls
			}
		}
		if opts.Package == "" {
			opts.Package = "main"
		}
	}

	var usesNames []string
	if len(opts.Uses) > 0 {
		for _, name := range strings.Split(opts.Uses, ",") {
			usesNames = append(usesNames, name)
		}
	}

	err = ccg.Copy(ccg.Config{
		From:    "github.com/reusee/codes/" + args[0],
		Params:  params,
		Renames: renames,
		Writer:  buf,
		Package: opts.Package,
		Decls:   decls,
		FileSet: fileSet,
		Uses:    usesNames,
	})
	if err != nil {
		log.Fatalf("ccg: copy error %v", err)
	}
	if opts.Output == "" {
		pt("%s\n", buf.Bytes())
	} else {
		err = ioutil.WriteFile(opts.Output, buf.Bytes(), 0644)
		if err != nil {
			log.Fatalf("ccg: write file error %v", err)
		}
	}
}
