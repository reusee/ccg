package sorter

import (
	"sort"
	"testing"
)

func TestAll(t *testing.T) {
	ints := []T{1, 3, 9, 4, 2}
	sort.Sort(Sorter{ints, func(a, b T) bool {
		return a.(int) < b.(int)
	}})
	if ints[0] != 1 || ints[1] != 2 || ints[2] != 3 || ints[3] != 4 || ints[4] != 9 {
		t.Fail()
	}
}
