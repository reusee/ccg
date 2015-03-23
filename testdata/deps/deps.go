package deps

type T int

func (t T) Foo() {}

func (t T) Bar() {
	t.Foo()
}
