package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"log"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
)

var (
	pt = fmt.Printf
)

func main() {
	fromPkg := flag.String("from", "", "package to read from")
	typeParamsStr := flag.String("params", "", "comma-separated param=arg pairs of parameterize spec")
	renamesStr := flag.String("renames", "", "comma-separated old=new pairs of rename spec")
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
	Copy(Config{
		From:    *fromPkg,
		Params:  typeParams,
		Renames: renames,
		Writer:  buf,
	})
	pt("%s\n", buf.Bytes())
}

type Config struct {
	From    string
	Params  map[string]string
	Renames map[string]string
	Writer  io.Writer
}

func Copy(config Config) error {
	// load package
	var loadConf loader.Config
	loadConf.Import(config.From)
	program, err := loadConf.Load()
	if err != nil {
		return fmt.Errorf("ccg: load package %v", err)
	}
	info := program.Imported[config.From]

	// remove type param declarations
	for _, f := range info.Files {
		newDecls := []ast.Decl{}
	decls:
		for _, decl := range f.Decls {
			if decl, ok := decl.(*ast.GenDecl); ok {
				if decl.Tok == token.TYPE {
					for _, spec := range decl.Specs {
						if spec, ok := spec.(*ast.TypeSpec); ok {
							if _, ok := config.Params[spec.Name.Name]; ok {
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
	collectObjects(config.Params)
	collectObjects(config.Renames)

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
	if config.Writer != nil {
		err = format.Node(config.Writer, program.Fset, decls)
		if err != nil {
			return fmt.Errorf("ccg: format output %v", err)
		}
	}

	return nil
}
