# Introduction
ccg is a template-based code generation tool for golang.

# Install
```
go get github.com/reusee/ccg/cmd/ccg

```

# Example 0: parameterized types
First define the type

```go
package pair

type T1 interface{}
type T2 interface{}

type Pair struct {
 first  T1
 second T2
}

func (p Pair) First() T1 {
 return p.first
}

func (p Pair) Second() T2 {
 return p.second
}
```

Then put this package to a importable path, assuming $GOPATH/src/pair

Now invoke the ccg command to get type specialized codes

```bash
 ccg -f pair -t T1=int,T2=string -r Pair=IntStrPair,New=NewIntStrPair
```

The above command generates:

```go
type IntStrPair struct {
 first  int
 second string
}

func NewIntStrPair(first int, second string) IntStrPair {
 return IntStrPair{first, second}
}

func (p IntStrPair) First() int {
 return p.first
}

func (p IntStrPair) Second() string {
 return p.second
}
```

Type T1 and T2 are substituted by int and string. Pair and New are also renamed.

# Example 1: output to file / update existing file
Use option -o to write generated codes to a file instead of stdout.

```bash
 ccg -f pair -t T1=int,T2=string -r Pair=IntStrPair,New=NewIntStrPair -o foo.go
```

If the specified file is already exists, ccg will update declarations if they're present in that file, or append to if not.
Other non-generated declarations will be preserved.

This means after updating template codes, you can re-invoke the command to update generated codes.
So it's friendly to go generate

```go
//go:generate ccg -f pair -t T1=int,T2=string -r Pair=IntStrPair,New=NewIntStrPair -o foo.go
```

# Example 2: partial generation
By default, ccg will generate all declarations from template package (except params).
If this is not what you want, you can use -u option to specify what to generate

```
 ccg -f pair -t T1=int,T2=string -r Pair=IntStrPair,New=NewIntStrPair -u NewIntStrPair,IntStrPair.First
```

The above command generates:

```go
type IntStrPair struct {
  first  int
  second string
}

func NewIntStrPair(first int, second string) IntStrPair {
  return IntStrPair{first, second}
}

func (p IntStrPair) First() int {
  return p.first
}
```

Method Second is not generated. And type IntStrPair is automatically generated because it's depended by NewIntStrPair and First method.
