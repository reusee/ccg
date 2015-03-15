package ccg

import (
	crand "crypto/rand"
	"encoding/binary"
	"go/ast"
	"math/rand"
	"sort"
)

func init() {
	var seed int64
	binary.Read(crand.Reader, binary.LittleEndian, &seed)
	rand.Seed(seed)
}

type AstDecls []ast.Decl

func (s AstDecls) Reduce(initial interface{}, fn func(value interface{}, elem ast.Decl) interface{}) (ret interface{}) {
	ret = initial
	for _, elem := range s {
		ret = fn(ret, elem)
	}
	return
}

func (s AstDecls) Map(fn func(ast.Decl) ast.Decl) (ret AstDecls) {
	for _, elem := range s {
		ret = append(ret, fn(elem))
	}
	return
}

func (s AstDecls) Filter(filter func(ast.Decl) bool) (ret AstDecls) {
	for _, elem := range s {
		if filter(elem) {
			ret = append(ret, elem)
		}
	}
	return
}

func (s AstDecls) All(predict func(ast.Decl) bool) (ret bool) {
	ret = true
	for _, elem := range s {
		ret = predict(elem) && ret
	}
	return
}

func (s AstDecls) Any(predict func(ast.Decl) bool) (ret bool) {
	for _, elem := range s {
		ret = predict(elem) || ret
	}
	return
}

func (s AstDecls) Each(fn func(e ast.Decl)) {
	for _, elem := range s {
		fn(elem)
	}
}

func (s AstDecls) Shuffle() {
	for i := len(s) - 1; i >= 1; i-- {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

func (s AstDecls) Sort(cmp func(a, b ast.Decl) bool) {
	sort.Sort(sliceSorter{
		l: len(s),
		less: func(i, j int) bool {
			return cmp(s[i], s[j])
		},
		swap: func(i, j int) {
			s[i], s[j] = s[j], s[i]
		},
	})
}

type AstSpecs []ast.Spec

func (s AstSpecs) Reduce(initial interface{}, fn func(value interface{}, elem ast.Spec) interface{}) (ret interface{}) {
	ret = initial
	for _, elem := range s {
		ret = fn(ret, elem)
	}
	return
}

func (s AstSpecs) Map(fn func(ast.Spec) ast.Spec) (ret AstSpecs) {
	for _, elem := range s {
		ret = append(ret, fn(elem))
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

func (s AstSpecs) All(predict func(ast.Spec) bool) (ret bool) {
	ret = true
	for _, elem := range s {
		ret = predict(elem) && ret
	}
	return
}

func (s AstSpecs) Any(predict func(ast.Spec) bool) (ret bool) {
	for _, elem := range s {
		ret = predict(elem) || ret
	}
	return
}

func (s AstSpecs) Each(fn func(e ast.Spec)) {
	for _, elem := range s {
		fn(elem)
	}
}

func (s AstSpecs) Shuffle() {
	for i := len(s) - 1; i >= 1; i-- {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

func (s AstSpecs) Sort(cmp func(a, b ast.Spec) bool) {
	sort.Sort(sliceSorter{
		l: len(s),
		less: func(i, j int) bool {
			return cmp(s[i], s[j])
		},
		swap: func(i, j int) {
			s[i], s[j] = s[j], s[i]
		},
	})
}

type sliceSorter struct {
	l    int
	less func(i, j int) bool
	swap func(i, j int)
}

func (t sliceSorter) Len() int {
	return t.l
}

func (t sliceSorter) Less(i, j int) bool {
	return t.less(i, j)
}

func (t sliceSorter) Swap(i, j int) {
	t.swap(i, j)
}
