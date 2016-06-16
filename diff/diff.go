// Copyright 2014-2016 The project AUTHORS. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package diff implements diff functions to compare objects.
package diff

import (
	"encoding/json"
	"errors"
	"reflect"
)

type Diff map[string]interface{}

func (d Diff) JSON() []byte {
	bs, err := json.MarshalIndent(d, "", "   ")
	if err != nil {
		return []byte{}
	}
	return bs
}

type Change struct {
	OldVal interface{} `json:"old_value"`
	NewVal interface{} `json:"new_value"`
}

type ChangeType string

const (
	AddType ChangeType = "ADD"
	DelType ChangeType = "DEL"
)

type CollectionChange struct {
	Val  interface{} `json:"value"`
	Type ChangeType  `json:"type"`
}

func ComputeDiff(x, y interface{}, recursive bool) (Diff, error) {
	vx, vy := reflect.ValueOf(x), reflect.ValueOf(y)
	tx, ty := vx.Type(), vy.Type()

	if !tx.AssignableTo(ty) {
		return nil, errors.New("input objects do not share the same type")
	}

	// since x and y share the same type, there is no need to check them both
	if vx.Kind() != reflect.Struct {
		return nil, errors.New("input values are not struct")
	}

	xNumFields := vx.NumField()

	delta := make(Diff)

	for i := 0; i < xNumFields; i++ {
		fx := vx.Field(i)
		typ := tx.Field(i)
		fy := vy.FieldByName(typ.Name)

		if d := handleValue(fx, fy); d != nil {
			delta[typ.Name] = d
		}
	}

	return delta, nil
}

func handleValue(fx, fy reflect.Value) interface{} {
	switch fx.Kind() {

	case reflect.Struct:
		delta := make(Diff)
		numFields := fx.NumField()
		for i := 0; i < numFields; i++ {
			newFx := fx.Field(i)
			typ := fx.Type().Field(i)
			newFy := fy.FieldByName(typ.Name)
			delta[typ.Name] = handleValue(newFx, newFy)
		}
		return delta

	case reflect.Array, reflect.Slice:
		xLen, yLen := fx.Len(), fy.Len()

		changes := make(map[int]CollectionChange)
		if xLen == 0 {
			if yLen == 0 {
				return nil
			}
			for i := 0; i < yLen; i++ {
				changes[i] = CollectionChange{Val: fy.Index(i).Interface(), Type: AddType}
			}
		} else if yLen == 0 {
			changes := make(map[int]CollectionChange)
			for i := 0; i < xLen; i++ {
				changes[i] = CollectionChange{Val: fx.Index(i).Interface(), Type: DelType}
			}
		} else {
			for i := 0; i < xLen; i++ {
				// TODO(gilliek): implement
			}
		}
		return changes

	case reflect.Map:
		return nil

	case reflect.Ptr:
		if fx.IsNil() {
			if fy.IsNil() {
				return nil
			}
			return Change{OldVal: nil, NewVal: fy.Elem().Interface()}
		} else if fy.IsNil() {
			return Change{OldVal: fx.Elem().Interface(), NewVal: nil}
		}
		return handleValue(fx.Elem(), fy.Elem())

	case reflect.Interface, reflect.Func, reflect.Chan, reflect.Invalid, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
		// TODO(gilliek): support complex numbers
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ix, iy := fx.Int(), fy.Int()
		if ix != iy {
			return Change{OldVal: ix, NewVal: iy}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uix, uiy := fx.Uint(), fy.Uint()
		if uix != uiy {
			return Change{OldVal: uix, NewVal: uiy}
		}
	case reflect.Float32, reflect.Float64:
		flx, fly := fx.Float(), fy.Float()
		if flx-fly < 0.000001 {
			return Change{OldVal: flx, NewVal: fly}
		}
	case reflect.String:
		sx, sy := fx.String(), fy.String()
		if sx != sy {
			return Change{OldVal: sx, NewVal: sy}
		}
	}

	return nil
}