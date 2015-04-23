package ccg

import (
	"bytes"
	"fmt"
	"go/ast"
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

func TestCopy2(t *testing.T) {
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
		Writer: buf,
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("copy/_expected2.go")
	checkResult(bytes.TrimSpace(expected), bytes.TrimSpace(buf.Bytes()), t)
}

func TestOverride(t *testing.T) {
	f, err := parser.ParseFile(new(token.FileSet), "foo", `
package foo
import "fmt"
import ft "fmt"
var foo = fmt.Printf
var bar = 5
var baz =ft.Printf
const c = 5
	`, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := new(bytes.Buffer)
	err = Copy(Config{
		From:     "github.com/reusee/ccg/testdata/override",
		Writer:   buf,
		Existing: []*ast.File{f},
		Package:  "foo",
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
			"M": "foobar",
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
	// params
	err := Copy(Config{
		From: "github.com/reusee/ccg/testdata/var",
		Params: map[string]string{
			"FOOBARBAZ": "foobarbaz",
		},
	})
	if err == nil || !strings.HasPrefix(err.Error(), "ccg: process\nname not found") {
		t.Fail()
	}
	// renames
	err = Copy(Config{
		From: "github.com/reusee/ccg/testdata/var",
		Renames: map[string]string{
			"FOOBARBAZ": "foobarbaz",
		},
	})
	if err == nil || !strings.HasPrefix(err.Error(), "ccg: process\nname not found") {
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
		From:     "github.com/reusee/ccg/testdata/deps",
		Writer:   buf,
		Uses:     []string{"T.Foo"},
		Package:  "foo",
		Existing: []*ast.File{f},
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

func TestInvalidUses(t *testing.T) {
	err := Copy(Config{
		From: "github.com/reusee/ccg/testdata/copy",
		Uses: []string{"foo.bar.baz"},
	})
	if err == nil || err.Error() != "invalid use spec: foo.bar.baz" {
		t.Fail()
	}
}

func TestRetypeWithUses(t *testing.T) {
	buf := new(bytes.Buffer)
	err := Copy(Config{
		From: "github.com/reusee/ccg/testdata/uses",
		Params: map[string]string{
			"T": "string",
		},
		Renames: map[string]string{
			"Ts":  "Strings",
			"Foo": "FOO",
		},
		Writer:  buf,
		Uses:    []string{"Strings.Foo", "FOO", "baz"},
		Package: "foo",
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("uses/_expected.go")
	checkResult(expected, buf.Bytes(), t)
}

func TestInvalidUsesType(t *testing.T) {
	err := Copy(Config{
		From: "github.com/reusee/ccg/testdata/uses",
		Uses: []string{"QWERTY.Foo"},
	})
	if err == nil || err.Error() != "QWERTY is not a type" {
		t.Fail()
	}
}

func TestInitFunction(t *testing.T) {
	f, err := parser.ParseFile(new(token.FileSet), "foo", `
package foo
func init() {
	_ = 42
}
	`, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := new(bytes.Buffer)
	err = Copy(Config{
		From:     "github.com/reusee/ccg/testdata/init",
		Writer:   buf,
		Package:  "foo",
		Existing: []*ast.File{f},
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("init/_expected.go")
	checkResult(expected, buf.Bytes(), t)
}

func TestInitFunction2(t *testing.T) {
	f, err := parser.ParseFile(new(token.FileSet), "foo", `
package foo
func init() {
	_ = "foobar"
}
	`, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := new(bytes.Buffer)
	err = Copy(Config{
		From:     "github.com/reusee/ccg/testdata/init",
		Writer:   buf,
		Package:  "foo",
		Existing: []*ast.File{f},
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	expected := readExpected("init/_expected2.go")
	checkResult(expected, buf.Bytes(), t)
}

func TestPkgNameDetect(t *testing.T) {
	buf := new(bytes.Buffer)
	err := Copy(Config{
		From:       "github.com/reusee/ccg/testdata/pkg",
		Writer:     buf,
		OutputFile: filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "reusee", "ccg", "testdata", "pkg", "gen.go"),
	})
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if string(buf.Bytes()) != "package foobarpkg\n" {
		t.Fatalf("copy: wrong package detected")
	}
}

func TestPkgNameDetectFail(t *testing.T) {
	err := Copy(Config{
		From:       "github.com/reusee/ccg/testdata/pkg",
		OutputFile: filepath.Join("non-exists-path", "non-exists.go"),
	})
	if err == nil {
		t.Fatal("should fail")
	}
	if !strings.HasPrefix(err.Error(), "ccg: detect package") {
		t.Fatalf("incorrect error, got %v", err)
	}
}
