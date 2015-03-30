package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
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
	From    string `short:"f" description:"template package import path"`
	Params  string `short:"t" description:"parameters"`
	Renames string `short:"r" description:"renames"`
	Package string `short:"p" description:"output package name"`
	Output  string `short:"o" description:"output file path"`
	Uses    string `short:"u" description:"names to be used only"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	if len(opts.From) == 0 {
		log.Fatal("no template package specified")
	}

	// params
	params := map[string]string{}
	if len(opts.Params) > 0 {
		for _, pairStr := range strings.Split(opts.Params, ",") {
			pair := strings.SplitN(pairStr, "=", 2)
			if len(pair) != 2 {
				log.Fatalf("invalid parameterize spec: %s", pairStr)
			}
			params[pair[0]] = pair[1]
		}
	}

	// renames
	renames := map[string]string{}
	if len(opts.Renames) > 0 {
		for _, pairStr := range strings.Split(opts.Renames, ",") {
			pair := strings.SplitN(pairStr, "=", 2)
			if len(pair) != 2 {
				log.Fatalf("invalid rename spec: %s", pairStr)
			}
			renames[pair[0]] = pair[1]
		}
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
	}

	var usesNames []string
	if len(opts.Uses) > 0 {
		for _, name := range strings.Split(opts.Uses, ",") {
			usesNames = append(usesNames, name)
		}
	}

	err = ccg.Copy(ccg.Config{
		From:       opts.From,
		Params:     params,
		Renames:    renames,
		Writer:     buf,
		Package:    opts.Package,
		Decls:      decls,
		FileSet:    fileSet,
		Uses:       usesNames,
		OutputFile: opts.Output,
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
