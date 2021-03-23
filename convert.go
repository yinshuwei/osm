// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Type conversions for Scan.

package osm

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

func setValue(isPtr bool, dest reflect.Value, value interface{}, destType reflect.Type) {
	if isPtr {
		data := reflect.New(destType)
		data.Elem().Set(reflect.ValueOf(value))
		dest.Set(data)
	} else {
		dest.Set(reflect.ValueOf(value))
	}
}

func setValueConvert(isPtr bool, dest reflect.Value, value interface{}, destType reflect.Type) {
	if isPtr {
		data := reflect.New(destType)
		data.Elem().Set(reflect.ValueOf(value).Convert(destType))
		dest.Set(data)
	} else {
		dest.Set(reflect.ValueOf(value).Convert(destType))
	}
}

// convertAssign copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
func convertAssign(dest reflect.Value, src interface{}, destIsPtr bool, destType reflect.Type) error {
	switch s := src.(type) {
	case string:
		switch destType.Kind() {
		case reflect.Slice:
			if destType.Elem().Kind() == reflect.Uint8 {
				setValue(destIsPtr, dest, []byte(s), destType)
				return nil
			}
		}
	case []byte:
		switch destType.Kind() {
		case reflect.String:
			setValue(destIsPtr, dest, string(s), destType)
			return nil
		}
	case time.Time:
		switch destType.Kind() {
		case reflect.String:
			setValue(destIsPtr, dest, s.Format(time.RFC3339Nano), destType)
			return nil
		case reflect.Slice:
			if destType.Elem().Kind() == reflect.Uint8 {
				setValue(destIsPtr, dest, []byte(s.Format(time.RFC3339Nano)), destType)
				return nil
			}
		}
		src = s.Local()
	case nil:
		if destIsPtr {
			dest.Set(reflect.New(destType))
		} else {
			dest.Set(reflect.New(destType).Elem())
		}
		return nil
	}

	srcType := reflect.TypeOf(src)
	if srcType.AssignableTo(destType) || srcType.Kind() == destType.Kind() {
		setValue(destIsPtr, dest, src, destType)
		return nil
	}
	if srcType.ConvertibleTo(destType) {
		if destType.Kind() != reflect.String {
			setValueConvert(destIsPtr, dest, src, destType)
			return nil
		}
	}

	var sv reflect.Value
	switch destType.Kind() {
	case reflect.String:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			setValue(destIsPtr, dest, asString(src), destType)
			return nil
		}
	case reflect.Slice:
		if destType.Elem().Kind() == reflect.Uint8 {
			sv = reflect.ValueOf(src)
			if b, ok := asBytes(nil, sv); ok {
				setValue(destIsPtr, dest, b, destType)
				return nil
			}
		}
	case reflect.Bool:
		bv, err := driver.Bool.ConvertValue(src)
		if err != nil {
			errorZapLogger.Error("convertAssign Bool", zap.Error(err))
			bv = false
		}
		setValue(destIsPtr, dest, bv.(bool), destType)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, destType.Bits())
		if err != nil {
			errorZapLogger.Error("convertAssign Int", zap.Error(err))
			errorZapLogger.Error("", zap.Error(err))
		}
		dest.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := asString(src)
		u64, err := strconv.ParseUint(s, 10, destType.Bits())
		if err != nil {
			errorZapLogger.Error("convertAssign Uint", zap.Error(err))
		}
		dest.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		s := asString(src)
		f64, err := strconv.ParseFloat(s, destType.Bits())
		if err != nil {
			errorZapLogger.Error("convertAssign Float", zap.Error(err))
		}
		dest.SetFloat(f64)
		return nil
	case reflect.Struct:
		if destType.String() == "time.Time" {
			str := ""
			switch s := src.(type) {
			case string:
				str = s
			case []byte:
				str = string(s)
			case nil:
				return nil
			}
			if str != "" {
				var t time.Time
				var err error
				if len(str) >= 19 {
					if strings.Contains(str, "T") {
						t, err = time.Parse(time.RFC3339Nano, str)
					} else {
						str = str[:19]
						t, err = time.ParseInLocation(formatDateTime, str, time.Local)
					}
				} else if len(str) >= 10 {
					str = str[:10]
					t, err = time.ParseInLocation(formatDate, str, time.Local)
				}
				if err == nil {
					t = t.Local()
					setValue(destIsPtr, dest, t, destType)
				} else {
					errorZapLogger.Error("convertAssign Time", zap.Error(err))
				}
			}
			return nil
		}
	}

	errorZapLogger.Error(fmt.Sprintf("unsupported Scan, storing driver.Value type %T into type %T", src, dest))
	return nil
}

func asString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

func asBytes(buf []byte, rv reflect.Value) (b []byte, ok bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(buf, rv.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.AppendUint(buf, rv.Uint(), 10), true
	case reflect.Float32:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 32), true
	case reflect.Float64:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 64), true
	case reflect.Bool:
		return strconv.AppendBool(buf, rv.Bool()), true
	case reflect.String:
		s := rv.String()
		return append(buf, s...), true
	}
	return
}
