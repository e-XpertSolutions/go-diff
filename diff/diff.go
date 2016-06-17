// Copyright 2014-2016 The project AUTHORS. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package diff implements diff functions to compare objects.
package diff

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type Diff map[string]interface{}

func (d Diff) PrettyJSON() []byte {
	bs, err := json.MarshalIndent(d, "", "   ")
	if err != nil {
		return []byte{}
	}
	return bs
}

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
	AddType ChangeType = "ADD" // addition
	DelType ChangeType = "DEL" // deletion
	ModType ChangeType = "MOD" // modification
)

type CollectionChange struct {
	Val  interface{} `json:"value"`
	Type ChangeType  `json:"type"`
}

func Compute(x, y interface{}, recursive bool) (Diff, error) {
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
		if !isExported(typ.Name) { // skip non-exported fields
			continue
		}
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
		if isFullyNonExportedStruct(fx) {
			if fx.String() != fy.String() {
				return Change{OldVal: fx.Interface(), NewVal: fy.Interface()}
			}
			return nil
		}

		delta := make(Diff)
		numFields := fx.NumField()
		for i := 0; i < numFields; i++ {
			newFx := fx.Field(i)
			typ := fx.Type().Field(i)
			if !isExported(typ.Name) { // skip non-exported fields
				continue
			}
			newFy := fy.FieldByName(typ.Name)

			if d := handleValue(newFx, newFy); d != nil {
				delta[typ.Name] = d
			}
		}
		if len(delta) > 0 {
			return delta
		}
		return nil

	case reflect.Array, reflect.Slice:
		xLen, yLen := fx.Len(), fy.Len()

		changes := make(map[string]CollectionChange)
		if xLen == 0 {
			if yLen == 0 {
				return nil
			}
			for i := 0; i < yLen; i++ {
				changes[strconv.Itoa(i)] = CollectionChange{Val: fy.Index(i).Interface(), Type: AddType}
			}
		} else if yLen == 0 {
			for i := 0; i < xLen; i++ {
				changes[strconv.Itoa(i)] = CollectionChange{Val: fx.Index(i).Interface(), Type: DelType}
			}
		} else {
			var maxLen int
			if xLen > yLen {
				maxLen = yLen
				for i := yLen; i < xLen; i++ {
					changes[strconv.Itoa(i)] = CollectionChange{Val: fx.Index(i).Interface(), Type: DelType}
				}
			} else if xLen < yLen {
				maxLen = xLen
				for i := xLen; i < yLen; i++ {
					changes[strconv.Itoa(i)] = CollectionChange{Val: fy.Index(i).Interface(), Type: AddType}
				}
			}
			for i := 0; i < maxLen; i++ {
				if d := handleValue(fx.Index(i), fy.Index(i)); d != nil {
					changes[strconv.Itoa(i)] = CollectionChange{Val: d, Type: ModType}
				}
			}
		}
		if len(changes) > 0 {
			return changes
		}
		return nil

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

func isExported(fieldName string) bool {
	if fieldName == "" {
		return false
	}
	firstLetter := string(fieldName[0])
	return firstLetter != strings.ToLower(firstLetter)
}

func isFullyNonExportedStruct(s reflect.Value) bool {
	if s.Kind() != reflect.Struct {
		return false
	}
	numFields := s.NumField()
	for i := 0; i < numFields; i++ {
		typ := s.Type().Field(i)
		if isExported(typ.Name) {
			return false
		}
	}
	return true
}
