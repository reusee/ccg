package ccg

import (
	"go/ast"

	"golang.org/x/tools/go/types"
)

import "fmt"

type AstDecls []ast.Decl

func (s AstDecls) Filter(filter func(ast.Decl) bool) (ret AstDecls) {
	for _, elem := range s {
		if filter(elem) {
			ret = append(ret, elem)
		}
	}
	return
}

type AstSpecs []ast.Spec

func (s AstSpecs) Filter(filter func(ast.Spec) bool) (ret AstSpecs) {
	for _, elem := range s {
		if filter(elem) {
			ret = append(ret, elem)
		}
	}
	return
}

type ObjectSet map[types.Object]struct{}

func NewObjectSet() ObjectSet {
	return ObjectSet(make(map[types.Object]struct{}))
}

func (s ObjectSet) Add(t types.Object) {
	s[t] = struct{}{}
}

func (s ObjectSet) In(t types.Object) (ok bool) {
	_, ok = s[t]
	return
}

type Err struct {
	Pkg  string
	Info string
	Err  error
}

func (e *Err) Error() string {
	return fmt.Sprintf("%s: %s\n%v", e.Pkg, e.Info, e.Err)
}

func makeErr(err error, info string) *Err {
	return &Err{
		Pkg:  "ccg",
		Info: info,
		Err:  err,
	}
}
