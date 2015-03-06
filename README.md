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
