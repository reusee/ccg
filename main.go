package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"log"
	"strconv"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
)

var (
	pt = fmt.Printf

	checkErr = func(desc string, err error) {
		if err != nil {
			log.Fatalf("%s error: %v", desc, err)
		}
	}

	fromPkg       = flag.String("from", "", "package to read from")
	typeParamsStr = flag.String("params", "", "types for argumentation")
	renamesStr    = flag.String("renames", "", "comma-separated old=new pairs of rename spec")
)

func main() {
	flag.Parse()

	// check source package
	if len(*fromPkg) == 0 {
		log.Fatalf("no package specified")
	}

	// check type parameters
	if len(*typeParamsStr) == 0 {
		log.Fatalf("no type parameter specified")
	}
	params := strings.Split(*typeParamsStr, ",")
	typeParams := map[string]string{
		"T": params[0],
	}
	n := 1
	for _, t := range params[1:] {
		typeParams["T"+strconv.Itoa(n)] = t
		n++
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

	// load package
	var config loader.Config
	config.Import(*fromPkg)
	program, err := config.Load()
	checkErr("load package", err)
	info := program.Imported[*fromPkg]

	// remove type param declarations
	for _, f := range info.Files {
		newDecls := []ast.Decl{}
	decls:
		for _, decl := range f.Decls {
			if decl, ok := decl.(*ast.GenDecl); ok {
				if decl.Tok == token.TYPE {
					for _, spec := range decl.Specs {
						if spec, ok := spec.(*ast.TypeSpec); ok {
							if _, ok := typeParams[spec.Name.Name]; ok {
								continue decls
							}
						}
					}
				}
			}
			newDecls = append(newDecls, decl)
		}
		f.Decls = newDecls
	}

	// collect objects to rename
	objects := make(map[types.Object]string)
	collectObjects := func(mapping map[string]string) {
		for from, to := range mapping {
			obj := info.Pkg.Scope().Lookup(from)
			if obj == nil {
				log.Fatalf("name not found %s", from)
			}
			objects[obj] = to
		}
	}
	collectObjects(typeParams)
	collectObjects(renames)

	// rename
	rename := func(defs map[*ast.Ident]types.Object) {
		for id, obj := range defs {
			if to, ok := objects[obj]; ok {
				id.Name = to
			}
		}
	}
	rename(info.Defs)
	rename(info.Uses)

	// collect output declarations
	decls := []ast.Decl{}
	for _, f := range info.Files {
		//TODO filter decls
		decls = append(decls, f.Decls...)
	}

	// output
	buf := new(bytes.Buffer)
	err = format.Node(buf, program.Fset, decls)
	checkErr("format", err)
	pt("%s\n", buf.Bytes())
}
