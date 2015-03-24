package ccg

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readExpected(path string) []byte {
	content, err := ioutil.ReadFile(filepath.Join(
		os.Getenv("GOPATH"), "src", "github.com/reusee/ccg/testdata/", path))
	if err != nil {
		panic(fmt.Sprintf("read file %s: %v", path, err))
	}
	return content
}

func checkResult(expected, got []byte, t *testing.T) {
	if !bytes.Equal(expected, got) {
		pt("== expected ==\n")
		pt("%s\n", expected)
		pt("== got ==\n")
		pt("%s\n", got)
		t.Fatalf("not match")
	}
}

func TestCopy(t *testing.T) {
	buf := new(bytes.Buffer)
	err := Copy(Config{
		From: "github.com/reusee/ccg/testdata/copy",
		Params: map[string]string{
			"T": "int",
		},
		Renames: map[string]string{
			"Ts":  "Ints",
			"Foo": "NewInts",
		},
		Package: "foo",
		Writer:  buf,
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("copy/_expected.go")
	checkResult(expected, buf.Bytes(), t)
}

func TestOverride(t *testing.T) {
	f, err := parser.ParseFile(new(token.FileSet), "foo", `
package foo
var bar = 5
	`, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := new(bytes.Buffer)
	err = Copy(Config{
		From:    "github.com/reusee/ccg/testdata/override",
		Writer:  buf,
		Decls:   f.Decls,
		Package: "foo",
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("override/_expected.go")
	checkResult(expected, buf.Bytes(), t)
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
		Writer:  buf,
		Package: "foo",
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("var/_expected.go")
	checkResult(expected, buf.Bytes(), t)
}

func TestNameNotFound(t *testing.T) {
	err := Copy(Config{
		From: "github.com/reusee/ccg/testdata/var",
		Params: map[string]string{
			"FOOBARBAZ": "foobarbaz",
		},
	})
	if err == nil || !strings.HasPrefix(err.Error(), "ccg: name not found") {
		t.Fail()
	}
}

func TestDeps(t *testing.T) {
	buf := new(bytes.Buffer)
	err := Copy(Config{
		From:    "github.com/reusee/ccg/testdata/deps",
		Writer:  buf,
		Uses:    []string{"T.Bar", "B"},
		Package: "deps",
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("deps/_expected.go")
	checkResult(expected, buf.Bytes(), t)
}

func TestDepsWithDecls(t *testing.T) {
	f, err := parser.ParseFile(new(token.FileSet), "foo", `
package foo
type B string
func (b B) Foo() {}

	`, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := new(bytes.Buffer)
	err = Copy(Config{
		From:    "github.com/reusee/ccg/testdata/deps",
		Writer:  buf,
		Uses:    []string{"T.Foo"},
		Package: "foo",
		Decls:   f.Decls,
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("deps/_expected2.go")
	checkResult(expected, buf.Bytes(), t)
}

func TestDeps2(t *testing.T) {
	buf := new(bytes.Buffer)
	err := Copy(Config{
		From:    "github.com/reusee/ccg/testdata/deps",
		Writer:  buf,
		Uses:    []string{"T.Foo"},
		Package: "foo",
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("deps/_expected3.go")
	checkResult(expected, buf.Bytes(), t)
}

func TestImport(t *testing.T) {
	buf := new(bytes.Buffer)
	err := Copy(Config{
		From:    "github.com/reusee/ccg/testdata/import",
		Writer:  buf,
		Package: "foo",
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("import/_expected.go")
	checkResult(expected, buf.Bytes(), t)
}
