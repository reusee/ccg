package ccg

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"log"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
)

var (
	pt = fmt.Printf
)

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
