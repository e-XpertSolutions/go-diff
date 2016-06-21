// Copyright 2016 e-Xpert Solutions SA. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package diff implements diff functions to compare objects.
package diff

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// A Diff represents the changes between two structures.
type Diff map[string]interface{}

// HasChange report whether the Diff contains some changes.
func (d Diff) HasChange() bool {
	return len(d) > 0
}

// PrettyJSON serializes the Diff into JSON in a indented (hence human readable)
// way.
func (d Diff) PrettyJSON() []byte {
	bs, err := json.MarshalIndent(d, "", "   ")
	if err != nil {
		return []byte{}
	}
	return bs
}

// JSON serializes the Diff into JSON in a compact format (no indentation).
func (d Diff) JSON() []byte {
	bs, err := json.Marshal(d)
	if err != nil {
		return []byte{}
	}
	return bs
}

type ChangeType string

// Possible values for a ChangeType.
const (
	AddType ChangeType = "ADD" // addition
	DelType ChangeType = "DEL" // deletion
	ModType ChangeType = "MOD" // modification
)

type Change struct {
	OldVal interface{} `json:"old_value,omitempty"`
	NewVal interface{} `json:"new_value,omitempty"`
	Type   ChangeType  `json:"type,omitempty"`
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
		return handleStruct(fx, fy)
	case reflect.Array, reflect.Slice:
		return handleSlice(fx, fy)
	case reflect.Map:
		return nil

	case reflect.Ptr:
		if fx.IsNil() {
			if fy.IsNil() {
				return nil
			}
			return Change{OldVal: nil, NewVal: fy.Elem().Interface()}
		} else if fy.IsNil() {
			return Change{OldVal: fx.Elem().Interface(), NewVal: nil, Type: ModType}
		}
		return handleValue(fx.Elem(), fy.Elem())

	case reflect.Interface, reflect.Func, reflect.Chan, reflect.Invalid, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
		// TODO(gilliek): support complex numbers
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ix, iy := fx.Int(), fy.Int()
		if ix != iy {
			return Change{OldVal: ix, NewVal: iy, Type: ModType}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uix, uiy := fx.Uint(), fy.Uint()
		if uix != uiy {
			return Change{OldVal: uix, NewVal: uiy, Type: ModType}
		}
	case reflect.Float32, reflect.Float64:
		flx, fly := fx.Float(), fy.Float()
		if flx-fly < 0.000001 {
			return Change{OldVal: flx, NewVal: fly, Type: ModType}
		}
	case reflect.String:
		sx, sy := fx.String(), fy.String()
		if sx != sy {
			return Change{OldVal: sx, NewVal: sy, Type: ModType}
		}
	}

	return nil
}

func handleStruct(fx, fy reflect.Value) interface{} {
	if isFullyNonExportedStruct(fx) {
		if !isEqual(fx, fy) {
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
}

func handleSlice(fx, fy reflect.Value) interface{} {
	xLen, yLen := fx.Len(), fy.Len()

	changes := make(Diff)
	if xLen == 0 {
		if yLen == 0 {
			return nil
		}
		for i := 0; i < yLen; i++ {
			changes[strconv.Itoa(i)] = Change{NewVal: fy.Index(i).Interface(), Type: AddType}
		}
	} else if yLen == 0 {
		for i := 0; i < xLen; i++ {
			changes[strconv.Itoa(i)] = Change{OldVal: fx.Index(i).Interface(), Type: DelType}
		}
	} else {
		var maxLen int
		if xLen > yLen {
			maxLen = yLen
			for i := yLen; i < xLen; i++ {
				changes[strconv.Itoa(i)] = Change{OldVal: fx.Index(i).Interface(), Type: DelType}
			}
		} else if xLen < yLen {
			maxLen = xLen
			for i := xLen; i < yLen; i++ {
				changes[strconv.Itoa(i)] = Change{NewVal: fy.Index(i).Interface(), Type: AddType}
			}
		}
		for i := 0; i < maxLen; i++ {
			if d := handleValue(fx.Index(i), fy.Index(i)); d != nil {
				changes[strconv.Itoa(i)] = d
			}
		}
	}
	if len(changes) > 0 {
		return changes
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

func isEqual(x, y reflect.Value) bool {
	return fmt.Sprint(x.Interface()) == fmt.Sprint(y.Interface())
}
