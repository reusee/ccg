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
	err := Copy(Config{
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
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
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
	f, err := parser.ParseFile(new(token.FileSet), "foo", `
package foo
type IntSet int
var foo = 42
func NewIntSet() {}
func (s IntSet) Add() {}
	`, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := new(bytes.Buffer)
	err = Copy(Config{
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
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), []byte(
		`type IntSet map[int]struct{}

var foo = 42

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

func TestOverride2(t *testing.T) {
	f, err := parser.ParseFile(new(token.FileSet), "foo", `
package foo
var bar = 42
`, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := new(bytes.Buffer)
	err = Copy(Config{
		From:   "github.com/reusee/ccg/testdata/override",
		Decls:  f.Decls,
		Writer: buf,
	})
	if !bytes.Equal(buf.Bytes(), []byte(
		`var bar = 5`)) {
		pt("generated: %s\n", buf.Bytes())
		t.Fatalf("copy")
	}
}

func TestNonExistsPackage(t *testing.T) {
	err := Copy(Config{
		From: "non-exists",
	})
	if err == nil || !strings.HasPrefix(err.Error(), "ccg: load package") {
		t.Fail()
	}
}

func TestVar(t *testing.T) {
	buf := new(bytes.Buffer)
	err := Copy(Config{
		From: "github.com/reusee/ccg/testdata/var",
		Params: map[string]string{
			"N": "42",
		},
		Writer: buf,
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	//TODO
}

func TestNameNotFound(t *testing.T) {
	err := Copy(Config{
		From: "github.com/reusee/ccg/set",
		Params: map[string]string{
			"FOOBARBAZ": "foobarbaz",
		},
	})
	if err == nil || !strings.HasPrefix(err.Error(), "ccg: name not found") {
		t.Fail()
	}
	err = Copy(Config{
		From: "github.com/reusee/ccg/set",
		Renames: map[string]string{
			"FOOBARBAZ": "foobarbaz",
		},
	})
	if err == nil || !strings.HasPrefix(err.Error(), "ccg: name not found") {
		t.Fail()
	}
}
