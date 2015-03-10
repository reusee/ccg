package slice

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sort"
)

func init() {
	var seed int64
	binary.Read(crand.Reader, binary.LittleEndian, &seed)
	rand.Seed(seed)
}

type T interface{}

type Ts []T

func (s Ts) Reduce(initial interface{}, fn func(value interface{}, elem T) interface{}) (ret interface{}) {
	ret = initial
	for _, elem := range s {
		ret = fn(ret, elem)
	}
	return
}

func (s Ts) Map(fn func(T) T) (ret Ts) {
	for _, elem := range s {
		ret = append(ret, fn(elem))
	}
	return
}

func (s Ts) Filter(filter func(T) bool) (ret Ts) {
	for _, elem := range s {
		if filter(elem) {
			ret = append(ret, elem)
		}
	}
	return
}

func (s Ts) All(predict func(T) bool) (ret bool) {
	ret = true
	for _, elem := range s {
		ret = predict(elem) && ret
	}
	return
}

func (s Ts) Any(predict func(T) bool) (ret bool) {
	for _, elem := range s {
		ret = predict(elem) || ret
	}
	return
}

func (s Ts) Each(fn func(e T)) {
	for _, elem := range s {
		fn(elem)
	}
}

func (s Ts) Shuffle() {
	for i := len(s) - 1; i >= 1; i-- {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

func (s Ts) Sort(cmp func(a, b T) bool) {
	sort.Sort(sliceSorter{
		l: len(s),
		less: func(i, j int) bool {
			return cmp(s[i], s[j])
		},
		swap: func(i, j int) {
			s[i], s[j] = s[j], s[i]
		},
	})
}

type sliceSorter struct {
	l    int
	less func(i, j int) bool
	swap func(i, j int)
}

func (t sliceSorter) Len() int {
	return t.l
}

func (t sliceSorter) Less(i, j int) bool {
	return t.less(i, j)
}

func (t sliceSorter) Swap(i, j int) {
	t.swap(i, j)
}
