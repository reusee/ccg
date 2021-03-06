package ccg

//go:generate myccg -u AstDecls.Filter -o utils.go slice ast.Decl AstDecls
//go:generate myccg -u AstSpecs.Filter -o utils.go slice ast.Spec AstSpecs
//go:generate myccg -u ObjectSet.Add,ObjectSet.In,NewObjectSet -o utils.go set types.Object ObjectSet NewObjectSet
//go:generate myccg -u StrSet.Add,StrSet.In,NewStrSet -o utils.go set string StrSet NewStrSet
//go:generate myccg err ccg -o utils.go

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/imports"
)

var (
	pt = fmt.Printf
	sp = fmt.Sprintf
)

type Config struct {
	// generation options
	From     string
	Params   map[string]string
	Renames  map[string]string
	Existing []*ast.File
	FileSet  *token.FileSet
	Uses     []string

	// output options
	Writer     io.Writer
	Package    string
	OutputFile string
}

func Copy(config Config) (ret error) {
	// load package
	loadConf := loader.Config{
		Fset:       config.FileSet,
		ParserMode: parser.ParseComments,
	}
	loadConf.Import(config.From)
	program, err := loadConf.Load()
	if err != nil {
		return me(err, "load package")
	}
	info := program.Imported[config.From]

	// utils functions
	formatNode := func(node interface{}) (string, error) {
		buf := new(bytes.Buffer)
		err := format.Node(buf, loadConf.Fset, node)
		if err != nil { //NOCOVER
			return "", me(err, "format node")
		}
		return string(buf.Bytes()), nil
	}

	// remove param declarations
	for _, f := range info.Files {
		f.Decls = filterDecls(f.Decls, func(node interface{}) bool {
			switch node := node.(type) {
			case *ast.TypeSpec:
				name := node.Name.Name
				_, exists := config.Params[name]
				return !exists
			case valueInfo:
				_, exists := config.Params[node.Name.Name]
				return !exists
			}
			return true
		})
	}

	// collect objects to rename
	renamed := map[string]string{}
	objects := make(map[types.Object]string)
	collectObjects := func(mapping map[string]string) error {
		for from, to := range mapping {
			obj := info.Pkg.Scope().Lookup(from)
			if obj == nil {
				return fmt.Errorf("name not found %s", from)
			}
			if t, ok := obj.Type().(*types.Basic); ok && t.Kind() == types.String {
				to = "`" + to + "`"
			}
			objects[obj] = to
			renamed[to] = from
		}
		return nil
	}
	if err := collectObjects(config.Params); err != nil {
		return me(err, "process")
	}
	if err := collectObjects(config.Renames); err != nil {
		return me(err, "process")
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
	existingDecls := make(map[string]func(interface{}))
	decls := []ast.Decl{}
	used := NewObjectSet()
	initFuncs := NewStrSet()
	var cmap ast.CommentMap
	mergeComments := func(f *ast.File) {
		if cmap == nil {
			cmap = ast.NewCommentMap(loadConf.Fset, f, f.Comments)
		} else {
			cm := ast.NewCommentMap(loadConf.Fset, f, f.Comments)
			for key, value := range cm {
				cmap[key] = value
			}
		}
	}
	collectExisting := func(f *ast.File) error {
		mergeComments(f)
		for i, decl := range f.Decls {
			decls = append(decls, decl)
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				case token.VAR, token.CONST:
					for _, spec := range decl.Specs {
						spec := spec.(*ast.ValueSpec)
						for i, name := range spec.Names {
							i := i
							spec := spec
							existingDecls[name.Name] = func(expr interface{}) {
								spec.Values[i] = expr.(ast.Expr)
							}
							used.Add(info.ObjectOf(name))
						}
					}
				case token.TYPE:
					for i, spec := range decl.Specs {
						spec := spec.(*ast.TypeSpec)
						i := i
						decl := decl
						existingDecls[spec.Name.Name] = func(expr interface{}) {
							decl.Specs[i].(*ast.TypeSpec).Type = expr.(ast.Expr)
						}
						used.Add(info.ObjectOf(spec.Name))
					}
				case token.IMPORT:
					for i, spec := range decl.Specs {
						spec := spec.(*ast.ImportSpec)
						i := i
						decl := decl
						var name string
						if spec.Name == nil {
							name = spec.Path.Value
						} else {
							name = spec.Name.Name
						}
						existingDecls[name] = func(path interface{}) {
							decl.Specs[i].(*ast.ImportSpec).Path = path.(*ast.BasicLit)
						}
					}
				}
			case *ast.FuncDecl:
				name := getFuncDeclName(decl)
				if name == "init" {
					src, err := formatNode(decl)
					if err != nil { //NOCOVER
						return me(err, "format init func")
					}
					initFuncs.Add(src)
					continue
				}
				i := i
				existingDecls[name] = func(fndecl interface{}) {
					decl := fndecl.(*ast.FuncDecl)
					decls[i] = decl
					used.Add(info.ObjectOf(decl.Name))
				}
				used.Add(info.ObjectOf(decl.Name))
			}
		}
		return nil
	}
	for _, f := range config.Existing {
		if err := collectExisting(f); err != nil { //NOCOVER
			return err
		}
	}

	// collect output declarations
	for _, f := range info.Files {
		mergeComments(f)
		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				case token.VAR, token.CONST:
					newDecl := &ast.GenDecl{
						Tok: decl.Tok,
					}
					for _, spec := range decl.Specs {
						spec := spec.(*ast.ValueSpec)
						for i, name := range spec.Names {
							if mutator, ok := existingDecls[name.Name]; ok {
								mutator(spec.Values[i])
							} else {
								newDecl.Specs = append(newDecl.Specs, spec)
							}
						}
					}
					if len(newDecl.Specs) > 0 {
						decls = append(decls, newDecl)
					}
				case token.TYPE:
					newDecl := &ast.GenDecl{
						Tok: token.TYPE,
					}
					for _, spec := range decl.Specs {
						spec := spec.(*ast.TypeSpec)
						name := spec.Name.Name
						if mutator, ok := existingDecls[name]; ok {
							mutator(spec.Type)
						} else {
							newDecl.Specs = append(newDecl.Specs, spec)
						}
					}
					if len(newDecl.Specs) > 0 {
						decls = append(decls, newDecl)
					}
				case token.IMPORT:
					newDecl := &ast.GenDecl{
						Tok: token.IMPORT,
					}
					for _, spec := range decl.Specs {
						spec := spec.(*ast.ImportSpec)
						var name string
						if spec.Name == nil {
							name = spec.Path.Value
						} else {
							name = spec.Name.Name
						}
						if mutator, ok := existingDecls[name]; ok {
							mutator(spec.Path)
						} else {
							newDecl.Specs = append(newDecl.Specs, spec)
						}
					}
					if len(newDecl.Specs) > 0 {
						decls = append(decls, newDecl)
					}
				}
			case *ast.FuncDecl:
				name := getFuncDeclName(decl)
				if name == "init" {
					src, err := formatNode(decl)
					if err != nil { //NOCOVER
						return me(err, "format init func")
					}
					if !initFuncs.In(src) { // not duplicated
						decls = append(decls, decl)
					}
					continue
				}
				if mutator, ok := existingDecls[name]; ok {
					mutator(decl)
				} else {
					decls = append(decls, decl)
				}
			}
		}
	}

	// get function dependencies
	deps := make(map[types.Object]ObjectSet)
	for _, decl := range decls {
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

	// get objects being used
	for _, use := range config.Uses {
		parts := strings.Split(use, ".")
		switch len(parts) {
		case 2: // method
			var ty types.Object
			if from, ok := renamed[parts[0]]; ok { // renamed type, use original type name
				ty = info.Pkg.Scope().Lookup(from)
			} else {
				ty = info.Pkg.Scope().Lookup(parts[0])
			}
			typeName, ok := ty.(*types.TypeName)
			if !ok {
				return fmt.Errorf("%s is not a type", parts[0])
			}
			obj, _, _ := types.LookupFieldOrMethod(typeName.Type(), true, info.Pkg, parts[1])
			used.Add(obj)
		case 1: // non-method
			var obj types.Object
			if from, ok := renamed[parts[0]]; ok { // renamed function
				obj = info.Pkg.Scope().Lookup(from)
			} else {
				obj = info.Pkg.Scope().Lookup(parts[0])
			}
			used.Add(obj)
		default:
			return fmt.Errorf("invalid use spec: %s", use)
		}
	}

	// filter
	if len(config.Uses) > 0 {
		// calculate uses closure
		for {
			l := len(used)
			for use := range used {
				if deps, ok := deps[use]; ok {
					for dep := range deps {
						used.Add(dep)
					}
				}
			}
			if len(used) == l {
				break
			}
		}
		// filter
		decls = filterDecls(decls, func(node interface{}) bool {
			switch node := node.(type) {
			case *ast.FuncDecl:
				return used.In(info.ObjectOf(node.Name))
			case *ast.TypeSpec:
				return used.In(info.ObjectOf(node.Name))
			case valueInfo:
				return used.In(info.ObjectOf(node.Name))
			}
			return true
		})
	}

	// decls tidy ups
	newDecls := []ast.Decl{}
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
		// move import decls to the beginning
		if decl, ok := decl.(*ast.GenDecl); ok && decl.Tok == token.IMPORT {
			importDecls = append(importDecls, decl)
			continue
		}
		newDecls = append(newDecls, decl)
	}
	decls = append(importDecls, newDecls...)

	// output
	if config.Writer == nil { //NOCOVER
		config.Writer = os.Stdout
	}
	if config.OutputFile != "" && config.Package == "" { // detect package name
		buildPkg, err := build.Default.ImportDir(filepath.Dir(config.OutputFile), 0)
		if err != nil {
			return me(err, "detect package")
		}
		config.Package = buildPkg.Name
	}
	var src interface{}
	if config.Package != "" {
		f := &ast.File{
			Name:  ast.NewIdent(config.Package),
			Decls: decls,
		}
		f.Comments = cmap.Filter(f).Comments()
		src = f
	} else {
		src = decls
	}
	buf := new(bytes.Buffer)
	if err := format.Node(buf, program.Fset, src); err != nil { //NOCOVER
		return me(err, "format")
	}
	var bs []byte
	if config.Package != "" {
		bs, err = imports.Process("", buf.Bytes(), nil)
		if err != nil { //NOCOVER
			return me(err, "imports")
		}
	} else {
		bs = buf.Bytes()
	}
	config.Writer.Write(bs)

	return nil
}

type astVisitor func(ast.Node) astVisitor

func (v astVisitor) Visit(node ast.Node) ast.Visitor {
	return v(node)
}

type valueInfo struct {
	Name  *ast.Ident
	Value ast.Expr
	Type  ast.Expr
}

func filterDecls(decls []ast.Decl, fn func(interface{}) bool) []ast.Decl {
	decls = AstDecls(decls).Filter(func(decl ast.Decl) (ret bool) {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			ret = fn(decl)
		case *ast.GenDecl:
			switch decl.Tok {
			case token.VAR, token.CONST:
				decl.Specs = AstSpecs(decl.Specs).Filter(func(sp ast.Spec) bool {
					spec := sp.(*ast.ValueSpec)
					names := []*ast.Ident{}
					values := []ast.Expr{}
					for i, name := range spec.Names {
						var value ast.Expr
						if i < len(spec.Values) {
							value = spec.Values[i]
						}
						if fn(valueInfo{name, value, spec.Type}) {
							names = append(names, name)
							if value != nil {
								values = append(values, value)
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
			case token.TYPE, token.IMPORT:
				decl.Specs = AstSpecs(decl.Specs).Filter(func(spec ast.Spec) bool {
					return fn(spec)
				})
				ret = len(decl.Specs) > 0
			}
		}
		return
	})
	return decls
}

func getFuncDeclName(decl *ast.FuncDecl) (name string) {
	name = decl.Name.Name
	if decl.Recv != nil {
		expr := decl.Recv.List[0].Type
		for _, ok := expr.(*ast.Ident); !ok; _, ok = expr.(*ast.Ident) {
			switch e := expr.(type) {
			case *ast.StarExpr:
				expr = e.X
			default: //NOCOVER
				panic(sp("unknown receiver node type %T", expr))
			}
		}
		name = expr.(*ast.Ident).Name + "." + name
	}
	return
}
