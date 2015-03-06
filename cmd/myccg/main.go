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

	getArg := func(index int, usage string) string {
		if len(args) < index+1 {
			pt("usage: %s %s\n", os.Args[0], usage)
			os.Exit(-1)
		}
		return args[index]
	}

	command := args[0]
	switch command {
	case "sorter":
		usage := "sorter [type] [sorter type]"
		t := getArg(1, usage)
		sorterType := getArg(2, usage)
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
			From: "github.com/reusee/ccg/sorter",
			Params: map[string]string{
				"T": t,
			},
			Renames: map[string]string{
				"Sorter": sorterType,
			},
			Writer:  writer,
			Package: *packageName,
		})
		if *outputFile != "" {
		} else {
			pt("%s\n", writer.(*bytes.Buffer).Bytes())
		}
	default:
		pt("unknown command: %s\n", command)
		return
	}
}
