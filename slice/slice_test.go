package slice

import "testing"

func TestReduce(t *testing.T) {
	var ints Ts
	for i := 1; i <= 100; i++ {
		ints = append(ints, i)
	}
	sum := ints.Reduce(0, func(n interface{}, e T) interface{} {
		return n.(int) + e.(int)
	})
	if sum.(int) != 5050 {
		t.Fail()
	}
}

func TestMap(t *testing.T) {
	var ints Ts
	for i := 0; i < 100; i++ {
		ints = append(ints, i)
	}
	ints = ints.Map(func(e T) T {
		return e.(int) * 2
	})
	for i := 0; i < 100; i++ {
		if ints[i] != i*2 {
			t.Fail()
		}
	}
}

func TestFilter(t *testing.T) {
	var ints Ts
	for i := 0; i < 100; i++ {
		ints = append(ints, i)
	}
	ints = ints.Filter(func(e T) bool {
		return e.(int)%2 == 0
	})
	if len(ints) != 50 {
		t.Fail()
	}
}

func TestAll(t *testing.T) {
	var ints Ts
	for i := 0; i < 100; i++ {
		ints = append(ints, i)
	}
	if !ints.All(func(e T) bool {
		return e.(int) < 100
	}) {
		t.Fatal("less")
	}
	if ints.All(func(e T) bool {
		return e.(int)%2 == 0
	}) {
		t.Fatal("even")
	}
}

func TestAny(t *testing.T) {
	var ints Ts
	for i := 0; i < 100; i++ {
		ints = append(ints, i)
	}
	if ints.Any(func(e T) bool {
		return e.(int) < 0
	}) {
		t.Fatal("less")
	}
	if !ints.Any(func(e T) bool {
		return e.(int) > 98
	}) {
		t.Fatal("large")
	}
}

func TestEach(t *testing.T) {
	var ints Ts
	for i := 0; i < 100; i++ {
		ints = append(ints, i)
	}
	n := 0
	ints.Each(func(e T) {
		n++
	})
	if n != 100 {
		t.Fail()
	}
}

func TestShuffle(t *testing.T) {
	var ints Ts
	for i := 0; i < 100; i++ {
		ints = append(ints, i)
	}
	ints.Shuffle()
	n := 0
	for i := 0; i < 100; i++ {
		if ints[i] == i {
			n++
		}
	}
	if n == 100 {
		t.Fail()
	}
}

func TestSort(t *testing.T) {
	var ints Ts
	for i := 0; i < 100; i++ {
		ints = append(ints, i)
	}
	ints.Shuffle()
	ints.Sort(func(a, b T) bool {
		return a.(int) < b.(int)
	})
	for i := 0; i < 100; i++ {
		if ints[i] != i {
			t.Fail()
		}
	}
}
