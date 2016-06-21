// Copyright 2016 e-Xpert Solutions SA. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import "testing"

type Foo struct {
	IntVal    int
	FloatVal  float32
	StringVal string
	Bar       Bar
	FooPtr    *Foo
	IntList   []int
	BarList   []Bar
}

type Bar struct {
	StringVal string
}

func TestCompute(t *testing.T) {
	f1 := Foo{
		IntVal:    42,
		FloatVal:  53.032,
		StringVal: "bar",
		Bar: Bar{
			StringVal: "ok",
		},
		FooPtr:  nil,
		IntList: []int{1, 3, 4},
		BarList: []Bar{{StringVal: "aaa"}, {StringVal: "bbb"}},
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
		BarList: []Bar{{StringVal: "ccc"}, {StringVal: "ddd"}},
	}
	delta, err := Compute(f1, f2)
	if err != nil {
		t.Fatal("Failed to compute diff: ", err)
	}
	t.Log(string(delta.PrettyJSON()))
}
