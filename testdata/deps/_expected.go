package deps

type T int

func (t T) Foo() {
	t.Baz()
}

func (t T) Bar() {
	t.Foo()
}

func (t *T) Baz() {}

type B int
