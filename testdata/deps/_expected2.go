package foo

type B int

func (b B) Foo() {}

type T int

func (t T) Foo() {
	t.Baz()
}

func (t T) Baz() {}
