package ccg

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"log"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/imports"
)

var (
	pt = fmt.Printf
)

type Config struct {
	From    string
	Params  map[string]string
	Renames map[string]string
	Writer  io.Writer
	Package string
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

	// remove param declarations
	for _, f := range info.Files {
		f.Decls = AstDecls(f.Decls).Filter(func(decl ast.Decl) bool {
			if decl, ok := decl.(*ast.GenDecl); !ok {
				return true
			} else {
				if decl.Tok == token.TYPE {
					decl.Specs = AstSpecs(decl.Specs).Filter(func(spec ast.Spec) bool {
						name := spec.(*ast.TypeSpec).Name.Name
						_, exists := config.Params[name]
						return !exists
					})
					return len(decl.Specs) > 0
				} else if decl.Tok == token.VAR {
					decl.Specs = AstSpecs(decl.Specs).Filter(func(sp ast.Spec) bool {
						spec := sp.(*ast.ValueSpec)
						names := []*ast.Ident{}
						values := []ast.Expr{}
						for i, name := range spec.Names {
							if _, exists := config.Params[name.Name]; !exists {
								names = append(names, name)
								values = append(values, spec.Values[i])
							}
						}
						spec.Names = names
						spec.Values = values
						return len(spec.Names) > 0
					})
					return len(decl.Specs) > 0
				}
				//TODO token.CONST
			}
			return true
		})
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
		//TODO merge existing decls
		decls = append(decls, f.Decls...)
	}

	// output
	if config.Writer != nil {
		if config.Package != "" { // output complete file
			file := &ast.File{
				Name:  ast.NewIdent(config.Package),
				Decls: decls,
			}
			buf := new(bytes.Buffer)
			err = format.Node(buf, program.Fset, file)
			if err != nil {
				return fmt.Errorf("ccg: format output %v", err)
			}
			bs, err := imports.Process("", buf.Bytes(), nil)
			if err != nil {
				return fmt.Errorf("ccg: format output %v", err)
			}
			config.Writer.Write(bs)
		} else { // output decls only
			err = format.Node(config.Writer, program.Fset, decls)
			if err != nil {
				return fmt.Errorf("ccg: format output %v", err)
			}
		}
	}

	return nil
}
