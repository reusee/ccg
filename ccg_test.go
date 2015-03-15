package ccg

import (
	"bytes"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestCopySet(t *testing.T) {
	buf := new(bytes.Buffer)
	Copy(Config{
		From: "github.com/reusee/ccg/set",
		Params: map[string]string{
			"T": "int",
		},
		Renames: map[string]string{
			"New": "NewIntSet",
			"Set": "IntSet",
		},
		Writer: buf,
	})
	if !bytes.Equal(buf.Bytes(), []byte(
		`type IntSet map[int]struct{}

func NewIntSet() IntSet {
	return IntSet(make(map[int]struct{}))
}

func (s IntSet) Add(t int) {
	s[t] = struct{}{}
}

func (s IntSet) Del(t int) {
	delete(s, t)
}

func (s IntSet) In(t int) (ok bool) {
	_, ok = s[t]
	return
}`)) {
		pt("generated: %s\n", buf.Bytes())
		t.Fail()
	}
}

func TestOverride(t *testing.T) {
	f, err := parser.ParseFile(new(token.FileSet), "test", `
package foo
type IntSet int
	`, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := new(bytes.Buffer)
	Copy(Config{
		From: "github.com/reusee/ccg/set",
		Params: map[string]string{
			"T": "int",
		},
		Renames: map[string]string{
			"New": "NewIntSet",
			"Set": "IntSet",
		},
		Writer: buf,
		Decls:  f.Decls,
	})
	if !bytes.Equal(buf.Bytes(), []byte(
		`type IntSet map[int]struct{}

func NewIntSet() IntSet {
	return IntSet(make(map[int]struct{}))
}

func (s IntSet) Add(t int) {
	s[t] = struct{}{}
}

func (s IntSet) Del(t int) {
	delete(s, t)
}

func (s IntSet) In(t int) (ok bool) {
	_, ok = s[t]
	return
}`)) {
		pt("generated: %s\n", buf.Bytes())
		t.Fatal("copy")
	}
}

func TestNonExistsPackage(t *testing.T) {
	err := Copy(Config{
		Package: "non-exists",
	})
	if err == nil || !strings.HasPrefix(err.Error(), "ccg: load package") {
		t.Fail()
	}
}
