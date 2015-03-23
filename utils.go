package ccg

import (
	"go/ast"

	"golang.org/x/tools/go/types"
)

type ObjectSet map[types.Object]struct{}

type AstSpecs []ast.Spec

type AstDecls []ast.Decl

func (s AstDecls) Filter(filter func(ast.Decl) bool) (ret AstDecls) {
	for _, elem := range s {
		if filter(elem) {
			ret = append(ret, elem)
		}
	}
	return
}

func (s AstSpecs) Filter(filter func(ast.Spec) bool) (ret AstSpecs) {
	for _, elem := range s {
		if filter(elem) {
			ret = append(ret, elem)
		}
	}
	return
}

func NewObjectSet() ObjectSet {
	return ObjectSet(make(map[types.Object]struct{}))
}

func (s ObjectSet) Add(t types.Object) {
	s[t] = struct{}{}
}

func (s ObjectSet) Del(t types.Object) {
	delete(s, t)
}

func (s ObjectSet) In(t types.Object) (ok bool) {
	_, ok = s[t]
	return
}

func (s ObjectSet) Each(fn func(types.Object)) {
	for e := range s {
		fn(e)
	}
}

func (s ObjectSet) Slice() (ret []types.Object) {
	for e := range s {
		ret = append(ret, e)
	}
	return
}
