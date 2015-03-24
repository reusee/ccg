package foo

import "fmt"

import bs "bytes"

func foo() {
	fmt.Printf("foo")
	_ = bs.Contains
}
