package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

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

	var writer io.Writer
	if *outputFile != "" {
		var err error
		writer, err = os.Create(*outputFile)
		if err != nil {
			log.Fatalf("myccg: create file %s error: %v", *outputFile, err)
		}
		defer writer.(*os.File).Close()
		if *packageName == "" {
			main := "main"
			packageName = &main
		}
	} else {
		writer = new(bytes.Buffer)
	}

	ccg.Copy(ccg.Config{
		From:    "github.com/reusee/ccg/" + args[0],
		Params:  params,
		Renames: renames,
		Writer:  writer,
		Package: *packageName,
	})
	if *outputFile == "" {
		pt("%s\n", writer.(*bytes.Buffer).Bytes())
	}
}
