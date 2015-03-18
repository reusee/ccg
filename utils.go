package ccg

import "go/ast"

type sliceSorter struct {
	l    int
	less func(i, j int) bool
	swap func(i, j int)
}

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
