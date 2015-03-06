package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/reusee/ccg"
)

var (
	pt = fmt.Printf
)

func main() {
	fromPkg := flag.String("from", "", "package to read from")
	typeParamsStr := flag.String("params", "", "comma-separated param=arg pairs of parameterize spec")
	renamesStr := flag.String("renames", "", "comma-separated old=new pairs of rename spec")
	packageName := flag.String("package", "", "package name of output file")
	flag.Parse()

	// check source package
	if len(*fromPkg) == 0 {
		log.Fatalf("no package specified")
	}

	// check type parameters
	if len(*typeParamsStr) == 0 {
		log.Fatalf("no type parameter specified")
	}
	typeParams := map[string]string{}
	for _, pairStr := range strings.Split(*typeParamsStr, ",") {
		pair := strings.SplitN(pairStr, "=", 2)
		if len(pair) != 2 {
			log.Fatalf("invalid parameterize spec: %s", pairStr)
		}
		typeParams[pair[0]] = pair[1]
	}

	// check renames
	renames := map[string]string{}
	if len(*renamesStr) > 0 {
		for _, pairStr := range strings.Split(*renamesStr, ",") {
			pair := strings.SplitN(pairStr, "=", 2)
			if len(pair) != 2 {
				log.Fatalf("invalid rename spec: %s", pairStr)
			}
			renames[pair[0]] = pair[1]
		}
	}

	buf := new(bytes.Buffer)
	ccg.Copy(ccg.Config{
		From:    *fromPkg,
		Params:  typeParams,
		Renames: renames,
		Writer:  buf,
		Package: *packageName,
	})
	pt("%s\n", buf.Bytes())
}
