// Copyright 2016 e-Xpert Solutions SA. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"reflect"
	"testing"
	"time"
)

func TestCompute(t *testing.T) {
	type (
		Bar struct {
			StringVal string
		}
		Foo struct {
			IntVal    int
			FloatVal  float32
			StringVal string
			Bar       Bar
			FooPtr    *Foo
			IntList   []int
			BarList   []Bar
		}
	)
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
	expectedJSON := `{"BarList":{"0":{"value":{"StringVal":{"old_value":"aaa","new_value":"ccc","type":"MOD"}},"type":"MOD"},"1":{"value":{"StringVal":{"old_value":"bbb","new_value":"ddd","type":"MOD"}},"type":"MOD"}},"FloatVal":{"old_value":53.03200149536133,"new_value":53.04199981689453,"type":"MOD"},"FooPtr":{"new_value":{"IntVal":42,"FloatVal":0,"StringVal":"","Bar":{"StringVal":""},"FooPtr":null,"IntList":null,"BarList":null},"type":"MOD"},"IntList":{"1":{"value":{"old_value":3,"new_value":2,"type":"MOD"},"type":"MOD"},"3":{"new_value":5,"type":"ADD"}},"StringVal":{"old_value":"bar","new_value":"baraca","type":"MOD"}}`
	if found := string(delta.JSON()); found != expectedJSON {
		t.Errorf("Compute(...): found '%s', expected '%s'", found, expectedJSON)
	}
	t.Log(string(delta.PrettyJSON()))
}

func TestIsFullyNonExportedStruct(t *testing.T) {
	type Foo struct {
		a int
		b int
	}
	f := Foo{a: 42, b: 42}
	if !isFullyNonExportedStruct(reflect.ValueOf(f)) {
		t.Errorf("isFullyNonExportedStruct(Foo{a: 42, b:42}): found 'false', expected 'true'")
	}

	type Bar struct {
		A int
		b int
	}
	b := Bar{A: 42, b: 42}
	if isFullyNonExportedStruct(reflect.ValueOf(b)) {
		t.Errorf("isFullyNonExportedStruct(Bar{A: 42, b:42}): found 'true', expected 'false'")
	}
}

func TestIsEqual(t *testing.T) {
	d1 := time.Date(2016, time.Month(6), 22, 10, 58, 52, 42, time.Local)
	d2 := time.Date(2016, time.Month(6), 22, 10, 58, 52, 42, time.Local)
	x, y := reflect.ValueOf(d1), reflect.ValueOf(d2)
	if equal := isEqual(x, y); !equal {
		t.Errorf("isEqual('%v', '%v'): found 'false', expected 'true'", d1, d2)
	}

	d1 = time.Date(2016, time.Month(6), 22, 10, 58, 52, 42, time.Local)
	d2 = time.Date(2016, time.Month(6), 22, 10, 58, 52, 24, time.Local)
	x, y = reflect.ValueOf(d1), reflect.ValueOf(d2)
	if equal := isEqual(x, y); equal {
		t.Errorf("isEqual('%v', '%v'): found 'true', expected 'false'", d1, d2)
	}
}

func TestExtractJSONName(t *testing.T) {
	tests := map[string]string{
		"-":             "",
		",omitempty":    "",
		"foo":           "foo",
		"foo,omitempty": "foo",
	}
	for input, expected := range tests {
		if found := extractJSONName(input); found != expected {
			t.Errorf("extractJSONName(%q): found %q, expected %q",
				input, found, expected)
		}
	}
}
