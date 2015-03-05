package sorter

type T interface{}

type Sorter struct {
	Slice []T
	Cmp   func(a, b T) bool
}

func (s Sorter) Len() int {
	return len(s.Slice)
}

func (s Sorter) Less(i, j int) bool {
	return s.Cmp(s.Slice[i], s.Slice[j])
}

func (s Sorter) Swap(i, j int) {
	s.Slice[i], s.Slice[j] = s.Slice[j], s.Slice[i]
}
