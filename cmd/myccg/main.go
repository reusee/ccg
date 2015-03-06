package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/reusee/ccg"
)

var (
	pt = fmt.Printf
)

func main() {
	if len(os.Args) < 2 {
		pt("usage: %s [command] [args...]\n", os.Args[0])
		return
	}

	getArg := func(index int, usage string) string {
		if len(os.Args) < index+1 {
			pt("usage: %s %s\n", os.Args[0], usage)
			os.Exit(-1)
		}
		return os.Args[index]
	}

	command := os.Args[1]
	switch command {
	case "sorter":
		usage := "sorter [type] [sorter type]"
		t := getArg(2, usage)
		sorterType := getArg(3, usage)
		buf := new(bytes.Buffer)
		ccg.Copy(ccg.Config{
			From: "github.com/reusee/ccg/sorter",
			Params: map[string]string{
				"T": t,
			},
			Renames: map[string]string{
				"Sorter": sorterType,
			},
			Writer: buf,
		})
		pt("%s\n", buf.Bytes())
	default:
		pt("unknown command: %s\n", command)
		return
	}
}
