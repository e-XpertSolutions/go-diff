// Copyright 2016 e-Xpert Solutions SA. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package diff implements diff functions to compare objects.
package diff

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Tolerance used to compare floating point numbers.
const Tolerance = 0.000001

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

// ChangeType describes the nature of a Change (addition, deletion,
// modification, ...).
type ChangeType string

// Possible values for a ChangeType.
const (
	AddType ChangeType = "ADD" // addition
	DelType ChangeType = "DEL" // deletion
	ModType ChangeType = "MOD" // modification
)

// A Change represents a change between an old and a new value.
type Change struct {
	OldVal interface{} `json:"old_value,omitempty"`
	NewVal interface{} `json:"new_value,omitempty"`
	Val    interface{} `json:"value,omitempty"`
	Type   ChangeType  `json:"type,omitempty"`
}

// Compute computes the differences between to objects x and y.
//
// x and y must be both structures and have to share the same type.
func Compute(x, y interface{}) (Diff, error) {
	return Engine{}.Compute(x, y)
}

// An Engine provides a flexible diff calculator.
type Engine struct {
	ExcludeFieldList []string
	MaxDepth         int // XXX(gilliek): not yet implemented
}

// IsIgnored reports whether a field is ignored by the Engine configuration.
func (e Engine) IsIgnored(field string) bool {
	for _, f := range e.ExcludeFieldList {
		if f == field {
			return true
		}
	}
	return false
}

// Compute computes the differences between to objects x and y using the
// parameters defined in the Engine.
//
// x and y must be both structures and have share the same type.
func (e Engine) Compute(x, y interface{}) (Diff, error) {
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

		// skip non-exported fields and the ones that are excluded
		if !isExported(typ.Name) || e.IsIgnored(typ.Name) {
			continue
		}

		fy := vy.FieldByName(typ.Name)

		if d := e.compareValues(fx, fy); d != nil {
			delta[typ.Name] = d
		}
	}

	return delta, nil
}

func (e Engine) compareValues(fx, fy reflect.Value) interface{} {
	switch fx.Kind() {

	// Structures, slices/arrays and maps must be recursively visited.
	case reflect.Struct:
		return e.compareStructs(fx, fy)
	case reflect.Array, reflect.Slice:
		return e.compareSlices(fx, fy)
	case reflect.Map:
		// TODO(gilliek): add support for map
		return nil

	// For pointers, only the values pointed are compared.
	case reflect.Ptr:
		if fx.IsNil() {
			if fy.IsNil() {
				return nil
			}
			return Change{OldVal: nil, NewVal: fy.Elem().Interface(), Type: ModType}
		} else if fy.IsNil() {
			return Change{OldVal: fx.Elem().Interface(), NewVal: nil, Type: ModType}
		}
		return e.compareValues(fx.Elem(), fy.Elem())

	// "basic" types are directly compared.
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
		if math.Abs(flx-fly) > Tolerance {
			return Change{OldVal: flx, NewVal: fly, Type: ModType}
		}
	case reflect.String:
		sx, sy := fx.String(), fy.String()
		if sx != sy {
			return Change{OldVal: sx, NewVal: sy, Type: ModType}
		}
	case reflect.Complex64, reflect.Complex128:
		// TODO(gilliek): add support for complex numbers
		return nil
	}

	return nil
}

func (e Engine) compareStructs(fx, fy reflect.Value) interface{} {
	if isFullyNonExportedStruct(fx) {
		if !isEqual(fx, fy) {
			return Change{OldVal: fx.Interface(), NewVal: fy.Interface(), Type: ModType}
		}
		return nil
	}

	delta := make(Diff)
	numFields := fx.NumField()
	for i := 0; i < numFields; i++ {
		newFx := fx.Field(i)
		typ := fx.Type().Field(i)

		// skip non-exported fields and the ones that are excluded
		if !isExported(typ.Name) || e.IsIgnored(typ.Name) {
			continue
		}

		newFy := fy.FieldByName(typ.Name)

		if d := e.compareValues(newFx, newFy); d != nil {
			delta[typ.Name] = d
		}
	}
	if len(delta) > 0 {
		return delta
	}
	return nil
}

func (e Engine) compareSlices(fx, fy reflect.Value) interface{} {
	xLen, yLen := fx.Len(), fy.Len()
	changes := make(map[string]Change)
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
		} else {
			maxLen = xLen
		}
		for i := 0; i < maxLen; i++ {
			if d := e.compareValues(fx.Index(i), fy.Index(i)); d != nil {
				changes[strconv.Itoa(i)] = Change{Val: d, Type: ModType}
			}
		}
	}
	if len(changes) > 0 {
		return changes
	}
	return nil
}

// isExported reports whether a field name is exported based on its name.
func isExported(fieldName string) bool {
	if fieldName == "" {
		return false
	}
	firstLetter := string(fieldName[0])
	return firstLetter != strings.ToLower(firstLetter)
}

// isFullyNonExportedStruct reports whether a structure only contains exported
// fields.
//
// If s is not a struct, it returns false.
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

// isEqual reports whether the two values x and y are equal. x and y must be
// two structures of the same type and being fully non-exported, meaning that
// they only contain non exported fields (e.g. time.Time).  Since it is not
// possible to compare such structures, this function transforms x and y into
// strings and compare them as strings.
func isEqual(x, y reflect.Value) bool {
	return fmt.Sprint(x.Interface()) == fmt.Sprint(y.Interface())
}
