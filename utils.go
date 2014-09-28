package osm

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

const (
	format_Date     = "2006-01-02"
	format_DateTime = "2006-01-02 15:04:05"
)

type strTo string

// set string
func (f *strTo) Set(v string) {
	if v != "" {
		*f = strTo(v)
	} else {
		f.Clear()
	}
}

// clean string
func (f *strTo) Clear() {
	*f = strTo(0x1E)
}

// check string exist
func (f strTo) Exist() bool {
	return string(f) != string(0x1E)
}

// string to bool
func (f strTo) Bool() (bool, error) {
	return strconv.ParseBool(f.String())
}

// string to float32
func (f strTo) Float32() (float32, error) {
	v, err := strconv.ParseFloat(f.String(), 32)
	return float32(v), err
}

// string to float64
func (f strTo) Float64() (float64, error) {
	return strconv.ParseFloat(f.String(), 64)
}

// string to int
func (f strTo) Int() (int, error) {
	v, err := strconv.ParseInt(f.String(), 10, 32)
	return int(v), err
}

// string to int8
func (f strTo) Int8() (int8, error) {
	v, err := strconv.ParseInt(f.String(), 10, 8)
	return int8(v), err
}

// string to int16
func (f strTo) Int16() (int16, error) {
	v, err := strconv.ParseInt(f.String(), 10, 16)
	return int16(v), err
}

// string to int32
func (f strTo) Int32() (int32, error) {
	v, err := strconv.ParseInt(f.String(), 10, 32)
	return int32(v), err
}

// string to int64
func (f strTo) Int64() (int64, error) {
	v, err := strconv.ParseInt(f.String(), 10, 64)
	return int64(v), err
}

// string to uint
func (f strTo) Uint() (uint, error) {
	v, err := strconv.ParseUint(f.String(), 10, 32)
	return uint(v), err
}

// string to uint8
func (f strTo) Uint8() (uint8, error) {
	v, err := strconv.ParseUint(f.String(), 10, 8)
	return uint8(v), err
}

// string to uint16
func (f strTo) Uint16() (uint16, error) {
	v, err := strconv.ParseUint(f.String(), 10, 16)
	return uint16(v), err
}

// string to uint31
func (f strTo) Uint32() (uint32, error) {
	v, err := strconv.ParseUint(f.String(), 10, 32)
	return uint32(v), err
}

// string to uint64
func (f strTo) Uint64() (uint64, error) {
	v, err := strconv.ParseUint(f.String(), 10, 64)
	return uint64(v), err
}

// string to string
func (f strTo) String() string {
	if f.Exist() {
		return string(f)
	}
	return ""
}

// interface to string
func toStr(value interface{}, args ...int) (s string) {
	switch v := value.(type) {
	case bool:
		s = strconv.FormatBool(v)
	case float32:
		s = strconv.FormatFloat(float64(v), 'f', argInt(args).Get(0, -1), argInt(args).Get(1, 32))
	case float64:
		s = strconv.FormatFloat(v, 'f', argInt(args).Get(0, -1), argInt(args).Get(1, 64))
	case int:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int8:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int16:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int32:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int64:
		s = strconv.FormatInt(v, argInt(args).Get(0, 10))
	case uint:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint8:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint16:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint32:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint64:
		s = strconv.FormatUint(v, argInt(args).Get(0, 10))
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		s = fmt.Sprintf("%v", v)
	}
	return s
}

// type argString []string

// // get string by index from string slice
// func (a argString) Get(i int, args ...string) (r string) {
// 	if i >= 0 && i < len(a) {
// 		r = a[i]
// 	} else if len(args) > 0 {
// 		r = args[0]
// 	}
// 	return
// }

type argInt []int

// get int by index from int slice
func (a argInt) Get(i int, args ...int) (r int) {
	if i >= 0 && i < len(a) {
		r = a[i]
	}
	if len(args) > 0 {
		r = args[0]
	}
	return
}

// type argAny []interface{}

// // get interface by index from interface slice
// func (a argAny) Get(i int, args ...interface{}) (r interface{}) {
// 	if i >= 0 && i < len(a) {
// 		r = a[i]
// 	}
// 	if len(args) > 0 {
// 		r = args[0]
// 	}
// 	return
// }

// // parse time to string with location
// func timeParse(dateString, format string) (time.Time, error) {
// 	tp, err := time.Parse(format, dateString)
// 	return tp, err
// }

// format time string
func timeFormat(t time.Time, format string) string {
	return t.Format(format)
}

// // get pointer indirect type
// func indirectType(v reflect.Type) reflect.Type {
// 	switch v.Kind() {
// 	case reflect.Ptr:
// 		return indirectType(v.Elem())
// 	default:
// 		return v
// 	}
// 	return v
// }

// set data to reflect.Value
func setDataToValue(value reflect.Value, data interface{}) {
	switch value.Kind() {
	case reflect.Bool:
		if data == nil {
			value.SetBool(false)
		} else if v, ok := data.(bool); ok {
			value.SetBool(v)
		} else {
			v, _ := strTo(toStr(data)).Bool()
			value.SetBool(v)
		}

	case reflect.String:
		if data == nil {
			value.SetString("")
		} else {
			value.SetString(toStr(data))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if data == nil {
			value.SetInt(0)
		} else {
			val := reflect.ValueOf(data)
			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value.SetInt(val.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value.SetInt(int64(val.Uint()))
			default:
				v, _ := strTo(toStr(data)).Int64()
				value.SetInt(v)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if data == nil {
			value.SetUint(0)
		} else {
			val := reflect.ValueOf(data)
			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value.SetUint(uint64(val.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value.SetUint(val.Uint())
			default:
				v, _ := strTo(toStr(data)).Uint64()
				value.SetUint(v)
			}
		}
	case reflect.Float64, reflect.Float32:
		if data == nil {
			value.SetFloat(0)
		} else {
			val := reflect.ValueOf(data)
			switch val.Kind() {
			case reflect.Float64:
				value.SetFloat(val.Float())
			default:
				v, _ := strTo(toStr(data)).Float64()
				value.SetFloat(v)
			}
		}

	case reflect.Struct:
		if data == nil {
			value.Set(reflect.Zero(value.Type()))

		} else if _, ok := value.Interface().(time.Time); ok {
			var str string
			switch d := data.(type) {
			case time.Time:
				value.Set(reflect.ValueOf(d))
			case []byte:
				str = string(d)
			case string:
				str = d
			}
			if str != "" {
				if len(str) >= 19 {
					str = str[:19]
					t, err := time.Parse(format_DateTime, str)
					if err == nil {
						value.Set(reflect.ValueOf(t))
					}
				} else if len(str) >= 10 {
					str = str[:10]
					t, err := time.Parse(format_Date, str)
					if err == nil {
						value.Set(reflect.ValueOf(t))
					}
				}
			}
		}
	}
}

// // camel string, xx_yy to XxYy
func toGoName(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i /**&& s[i+1] >= 'a' && s[i+1] <= 'z'**/ {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:len(data)])
}

// func toGoName(name string) string {
// 	names := strings.Split(name, "_")

// 	newName := ""
// 	for _, ntemp := range names {
// 		newName += strings.ToUpper(ntemp[0:1]) + strings.ToLower(ntemp[1:len(ntemp)])
// 	}
// 	return newName
// }
