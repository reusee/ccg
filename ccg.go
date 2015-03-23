package ccg

//go:generate myccg -uses AstDecls,AstDecls.Filter -package ccg -output utils.go slice ast.Decl AstDecls
//go:generate myccg -uses AstSpecs,AstSpecs.Filter -package ccg -output utils.go slice ast.Spec AstSpecs
//go:generate myccg -package ccg -output utils.go set types.Object ObjectSet NewObjectSet

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"strings"

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
	Decls   []ast.Decl
	FileSet *token.FileSet
	Uses    []string
}

func Copy(config Config) error {
	// load package
	loadConf := loader.Config{
		Fset: config.FileSet,
	}
	loadConf.Import(config.From)
	program, err := loadConf.Load()
	if err != nil {
		return fmt.Errorf("ccg: load package %v", err)
	}
	info := program.Imported[config.From]

	// remove param declarations
	for _, f := range info.Files {
		f.Decls = AstDecls(f.Decls).Filter(func(decl ast.Decl) (ret bool) {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				case token.TYPE:
					decl.Specs = AstSpecs(decl.Specs).Filter(func(spec ast.Spec) bool {
						name := spec.(*ast.TypeSpec).Name.Name
						_, exists := config.Params[name]
						return !exists
					})
					ret = len(decl.Specs) > 0
				case token.VAR:
					decl.Specs = AstSpecs(decl.Specs).Filter(func(sp ast.Spec) bool {
						spec := sp.(*ast.ValueSpec)
						names := []*ast.Ident{}
						values := []ast.Expr{}
						for i, name := range spec.Names {
							if _, exists := config.Params[name.Name]; !exists {
								names = append(names, name)
								if i < len(spec.Values) {
									values = append(values, spec.Values[i])
								}
							}
						}
						spec.Names = names
						if len(values) == 0 {
							spec.Values = nil
						} else {
							spec.Values = values
						}
						return len(spec.Names) > 0
					})
					ret = len(decl.Specs) > 0
				default:
					return true
				}
			default:
				return true
			}
			return
		})
	}

	// collect objects to rename
	objects := make(map[types.Object]string)
	collectObjects := func(mapping map[string]string) error {
		for from, to := range mapping {
			obj := info.Pkg.Scope().Lookup(from)
			if obj == nil {
				return fmt.Errorf("ccg: name not found %s", from)
			}
			objects[obj] = to
		}
		return nil
	}
	if err := collectObjects(config.Params); err != nil {
		return err
	}
	if err := collectObjects(config.Renames); err != nil {
		return err
	}

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

	// collect existing decls
	existingVars := make(map[string]func(expr ast.Expr))
	existingTypes := make(map[string]func(expr ast.Expr))
	existingFuncs := make(map[string]func(fn *ast.FuncDecl))
	var decls []ast.Decl
	for i, decl := range config.Decls {
		decls = append(decls, decl)
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.VAR:
				for _, spec := range decl.Specs {
					spec := spec.(*ast.ValueSpec)
					for i, name := range spec.Names {
						i := i
						spec := spec
						existingVars[name.Name] = func(expr ast.Expr) {
							spec.Values[i] = expr
						}
					}
				}
			case token.TYPE:
				for i, spec := range decl.Specs {
					spec := spec.(*ast.TypeSpec)
					i := i
					decl := decl
					existingTypes[spec.Name.Name] = func(expr ast.Expr) {
						decl.Specs[i].(*ast.TypeSpec).Type = expr
					}
				}
			}
		case *ast.FuncDecl:
			name := decl.Name.Name
			if decl.Recv != nil {
				name = decl.Recv.List[0].Type.(*ast.Ident).Name + "." + name
			}
			i := i
			existingFuncs[name] = func(fndecl *ast.FuncDecl) {
				decls[i] = fndecl
			}
		}
	}

	// collect output declarations
	var newDecls []ast.Decl
	for _, f := range info.Files {
		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				// var
				case token.VAR:
					newDecl := &ast.GenDecl{
						Tok: token.VAR,
					}
					for _, spec := range decl.Specs {
						spec := spec.(*ast.ValueSpec)
						for i, name := range spec.Names {
							if mutator, ok := existingVars[name.Name]; ok {
								mutator(spec.Values[i])
							} else {
								newDecl.Specs = append(newDecl.Specs, spec)
							}
						}
					}
					if len(newDecl.Specs) > 0 {
						newDecls = append(newDecls, newDecl)
					}
				// type
				case token.TYPE:
					newDecl := &ast.GenDecl{
						Tok: token.TYPE,
					}
					for _, spec := range decl.Specs {
						spec := spec.(*ast.TypeSpec)
						name := spec.Name.Name
						if mutator, ok := existingTypes[name]; ok {
							mutator(spec.Type)
						} else {
							newDecl.Specs = append(newDecl.Specs, spec)
						}
					}
					if len(newDecl.Specs) > 0 {
						newDecls = append(newDecls, newDecl)
					}
				// import or const
				default:
					newDecls = append(newDecls, decl)
				}
			// func
			case *ast.FuncDecl:
				name := decl.Name.Name
				if decl.Recv != nil {
					name = decl.Recv.List[0].Type.(*ast.Ident).Name + "." + name
				}
				if mutator, ok := existingFuncs[name]; ok {
					mutator(decl)
				} else {
					newDecls = append(newDecls, decl)
				}
			}
		}
	}

	// filter by uses
	// get function dependencies
	deps := make(map[types.Object]ObjectSet)
	for _, decl := range newDecls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			obj := info.ObjectOf(decl.Name)
			set := NewObjectSet()
			var visitor astVisitor
			visitor = func(node ast.Node) astVisitor {
				switch node := node.(type) {
				case *ast.Ident:
					dep := info.ObjectOf(node)
					set.Add(dep)
				}
				return visitor
			}
			ast.Walk(visitor, decl)
			deps[obj] = set
		}
	}
	// get uses objects
	uses := NewObjectSet()
	for _, use := range config.Uses {
		parts := strings.Split(use, ".")
		switch len(parts) {
		case 2: // method
			ty := info.Pkg.Scope().Lookup(parts[0])
			typeName, ok := ty.(*types.TypeName)
			if !ok {
				return fmt.Errorf("%s is not a type", parts[0])
			}
			obj, _, _ := types.LookupFieldOrMethod(typeName.Type(), true, info.Pkg, parts[1])
			uses.Add(obj)
		case 1: // non-method
			obj := info.Pkg.Scope().Lookup(parts[0])
			uses.Add(obj)
		default:
			return fmt.Errorf("invalid use spec: %s", use)
		}
	}
	// filter
	if len(uses) > 0 {
		// calculate uses closure
		for {
			l := len(uses)
			for use := range uses {
				if deps, ok := deps[use]; ok {
					for dep := range deps {
						uses.Add(dep)
					}
				}
			}
			if len(uses) == l {
				break
			}
		}
		newDecls = AstDecls(newDecls).Filter(func(decl ast.Decl) bool {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				obj := info.ObjectOf(decl.Name)
				return uses.In(obj)
			}
			return true
		})
	}

	// merge new and existing decls
	decls = append(decls, newDecls...)

	// decls tidy ups
	newDecls = newDecls[0:0]
	var importDecls []ast.Decl
	for _, decl := range decls {
		// ensure linebreak between decls
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Doc == nil {
				decl.Doc = new(ast.CommentGroup)
			}
		case *ast.GenDecl:
			if decl.Doc == nil {
				decl.Doc = new(ast.CommentGroup)
			}
		}
		// move import decls to beginning
		if decl, ok := decl.(*ast.GenDecl); ok && decl.Tok == token.IMPORT {
			importDecls = append(importDecls, decl)
			continue
		}
		newDecls = append(newDecls, decl)
	}
	decls = append(importDecls, newDecls...)

	// output
	if config.Writer != nil {
		if config.Package != "" { // output complete file
			file := &ast.File{
				Name:  ast.NewIdent(config.Package),
				Decls: decls,
			}
			buf := new(bytes.Buffer)
			err = format.Node(buf, program.Fset, file)
			if err != nil { //NOCOVER
				return fmt.Errorf("ccg: format output %v", err)
			}
			bs, err := imports.Process("", buf.Bytes(), nil)
			if err != nil { //NOCOVER
				return fmt.Errorf("ccg: format output %v", err)
			}
			config.Writer.Write(bs)
		} else { // output decls only
			err = format.Node(config.Writer, program.Fset, decls)
			if err != nil { //NOCOVER
				return fmt.Errorf("ccg: format output %v", err)
			}
		}
	}

	return nil
}

type astVisitor func(ast.Node) astVisitor

func (v astVisitor) Visit(node ast.Node) ast.Visitor {
	return v(node)
}
