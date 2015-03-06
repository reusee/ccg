# ccg
copying code generation

# install
```
go get github.com/reusee/ccg/cmd/ccg
```

# example
```
ccg -from github.com/reusee/ccg/sorter -params T=[]byte -renames Sorter=BytesSorter
```
output
```
type BytesSorter struct {
  Slice [][]byte
  Cmp   func(a, b []byte) bool
}

func (s BytesSorter) Len() int {
  return len(s.Slice)
}

func (s BytesSorter) Less(i, j int) bool {
  return s.Cmp(s.Slice[i], s.Slice[j])
}

func (s BytesSorter) Swap(i, j int) {
  s.Slice[i], s.Slice[j] = s.Slice[j], s.Slice[i]
}
```

```
ccg -from github.com/reusee/ccg/set -params T=string -renames Set=StrSet,New=NewStrSet -package foobar
```
output
```
package foobar

type StrSet map[string]struct{}

func NewStrSet() StrSet {
  return StrSet(make(map[string]struct{}))
}

func (s StrSet) Add(t string) {
  s[t] = struct{}{}
}

func (s StrSet) Del(t string) {
  delete(s, t)
}

func (s StrSet) In(t string) (ok bool) {
  _, ok = s[t]
  return
}
```
