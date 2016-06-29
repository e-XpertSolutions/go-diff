# go-diff

[![GoDoc](https://godoc.org/github.com/e-XpertSolutions/go-diff/diff?status.png)](http://godoc.org/github.com/e-XpertSolutions/go-diff/diff)
[![Travis](https://travis-ci.org/e-XpertSolutions/go-diff.svg?branch=master)](https://travis-ci.org/e-XpertSolutions/go-diff.svg)


go-diff is a Go package that aims to provide functions to calculate differences
between two Go structures. The result is presented as a generic map containing
old and new values. See the [documentation](http://godoc.org/github.com/e-XpertSolutions/go-diff/diff) for
more details and examples.


## Installation

```
go get -u github.com/e-XpertSolutions/go-diff/diff
```


## Usage

```go
package main

import (
  "fmt"
  "log"

  "github.com/e-XpertSolutions/go-diff/diff"
)

type StructA struct {
  SomeInt int
  B *StructB
}

type StructB struct {
  SomeVal string
  SomeFloat float32
}

func main() {
  a1 := StructA{
    SomeInt: 42,
    B: &StructB{ SomeVal: "Foo", SomeFloat: 123.456 },
  }
  a2 := StructA{
    SomeInt: 24,
    B: &StructB{ SomeVal: "Bar", SomeFloat: 123.4567 },
  }

  d, err :=  diff.Compute(a1, a2)
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(string(d.PrettyJSON()))

  // Output:
  // {
  //   "B": {
  //     "SomeVal": {
  //       "old_value": "Foo",
  //       "new_value": "Bar",
  //       "type": "MOD"
  //     }
  //   },
  //   "SomeInt": {
  //     "old_value": 42,
  //     "new_value": 24,
  //     "type": "MOD"
  //   }
  // }
}
```


## Diff calculation

* **Basic types (int, float, bool and string):** Basic types are directly compared
using language defined operators (== and !=).
* **Pointers:** The values pointed are compared, not the addresses.
* **Structures:** Structures are compared recursively. If they do not contain
any exported fields, the structures are compared as strings.
* **Slices and arrays:** Slices and arrays are compared row by row.
* **Maps:** Maps are not yet supported.


## TODO

- [ ] Add support for map types
- [ ] Add support for complex numbers
- [ ] Allow the user to override comparisons through an interface
- [ ] Add an option to limit the depth of the comparisons
- [ ] Better support for slices/arrays comparison
- [ ] More tests and benchmarks
- [ ] Add XML serialization


## Contributing

We appreciate any form of contribution (feature request, bug report,
pull request, ...). We have no special requirements for Pull Request,
just follow the standard [GitHub way](https://help.github.com/articles/using-pull-requests/).


## License

The sources are release under a BSD 3-Clause License. The full terms of that
license can be found in `LICENSE` file of this repository.
