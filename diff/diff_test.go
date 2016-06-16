// Copyright 2014-2016 The project AUTHORS. All rights reserved.
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
}

type Bar struct {
	StringVal string
}

func TestComputeDiff(t *testing.T) {
	f1 := Foo{
		IntVal:    42,
		FloatVal:  53.032,
		StringVal: "bar",
		Bar: Bar{
			StringVal: "ok",
		},
		FooPtr: nil,
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
	}
	delta, err := ComputeDiff(f1, f2, false)
	if err != nil {
		t.Fatal("Failed to compute diff: ", err)
	}
	t.Log(string(delta.JSON()))
}
