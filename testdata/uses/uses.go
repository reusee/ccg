package uses

import "fmt"

var ft = fmt.Printf

type T interface{}

type Ts []T

func (t Ts) Foo() {}

func Foo() {}

var bar = 42

var baz = 42
