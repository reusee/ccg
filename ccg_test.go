package main

import (
	"bytes"
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
		pt("%s\n", buf.Bytes())
		t.Fail()
	}
}
