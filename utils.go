package ccg

import "go/ast"

type AstSpecs []ast.Spec

type AstDecls []ast.Decl

func (
	s AstDecls) Filter(filter func(
	ast.Decl) bool) (ret AstDecls) {
	for _, elem := range s {
		if filter(elem) {
			ret = append(
				ret, elem)

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
