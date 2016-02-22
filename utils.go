package ccg

import (
	"go/ast"
	"go/types"
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
	Prev error
}

func (e *Err) Error() string {
	if e.Prev == nil {
		return fmt.Sprintf("%s: %s", e.Pkg, e.Info)
	}
	return fmt.Sprintf("%s: %s\n%v", e.Pkg, e.Info, e.Prev)
}

type StrSet map[string]struct{}

func NewStrSet() StrSet {
	return StrSet(make(map[string]struct{}))
}

func (s StrSet) Add(t string) {
	s[t] = struct{}{}
}

func (s StrSet) In(t string) (ok bool) {
	_, ok = s[t]
	return
}

func me(err error, format string, args ...interface{}) *Err {
	if len(args) > 0 {
		return &Err{
			Pkg:  `ccg`,
			Info: fmt.Sprintf(format, args...),
			Prev: err,
		}
	}
	return &Err{
		Pkg:  `ccg`,
		Info: format,
		Prev: err,
	}
}

func ce(err error, format string, args ...interface{}) {
	if err != nil {
		panic(me(err, format, args...))
	}
}

func ct(err *error) {
	if p := recover(); p != nil {
		if e, ok := p.(error); ok {
			*err = e
		} else {
			panic(p)
		}
	}
}
