package set

type T interface{}

type Set map[T]struct{}

func New() Set {
	return Set(make(map[T]struct{}))
}

func (s Set) Add(t T) {
	s[t] = struct{}{}
}

func (s Set) Del(t T) {
	delete(s, t)
}

func (s Set) In(t T) (ok bool) {
	_, ok = s[t]
	return
}
