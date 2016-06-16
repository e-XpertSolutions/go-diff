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
	//yNumFields := vy.NumField()

	delta := make(Diff)

	for i := 0; i < xNumFields; i++ {
		fx := vx.Field(i)
		typ := tx.Field(i)
		fy := vy.FieldByName(typ.Name)

		switch fx.Kind() {
		case reflect.Struct:
		case reflect.Array, reflect.Slice:
		case reflect.Map:
		case reflect.Ptr:
		case reflect.Interface, reflect.Func, reflect.Chan, reflect.Invalid, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
			// TODO(gilliek): handle complex numbers
			continue
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ix, iy := fx.Int(), fy.Int()
			if ix != iy {
				delta[typ.Name] = Change{OldVal: ix, NewVal: iy}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uix, uiy := fx.Uint(), fy.Uint()
			if uix != uiy {
				delta[typ.Name] = Change{OldVal: uix, NewVal: uiy}
			}
		case reflect.Float32, reflect.Float64:
			flx, fly := fx.Float(), fy.Float()
			if flx-fly < 0.000001 {
				delta[typ.Name] = Change{OldVal: flx, NewVal: fly}
			}
		case reflect.String:
			sx, sy := fx.String(), fy.String()
			if sx != sy {
				delta[typ.Name] = Change{OldVal: sx, NewVal: sy}
			}
		}
	}

	return delta, nil
}
