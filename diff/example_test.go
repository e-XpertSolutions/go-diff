// Copyright 2014-2016 The project AUTHORS. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff_test

import (
	"log"

	"e-xpert_solutions/go-diff/diff"
)

type Foo struct {
	IntVal    int
	FloatVal  float32
	StringVal string
	Bar       Bar
	FooPtr    *Foo
	IntList   []int
}

type Bar struct {
	StringVal string
}

func Example() {
	f1 := Foo{
		IntVal:    42,
		FloatVal:  53.032,
		StringVal: "bar",
		Bar: Bar{
			StringVal: "ok",
		},
		FooPtr:  nil,
		IntList: []int{1, 3, 4},
	}
	f2 := Foo{
		IntVal:    42,
		FloatVal:  53.042,
		StringVal: "baraca",
		Bar: Bar{
			StringVal: "ok",
		},
		FooPtr: &Foo{
			IntVal: 42,
		},
		IntList: []int{1, 2, 4, 5},
	}
	delta, err := diff.Compute(f1, f2, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("delta = ", delta)
}