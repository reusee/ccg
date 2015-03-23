package foo

type T int

func (t T) Foo() {
	t.Baz()
}

func (t T) Baz() {}
