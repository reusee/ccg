package ccg

import (
	"go/ast"

	"golang.org/x/tools/go/types"
)

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
