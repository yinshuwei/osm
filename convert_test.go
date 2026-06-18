package osm

import (
	"reflect"
	"testing"
	"time"
)

func TestAsString(t *testing.T) {
	tests := []struct {
		input interface{}
		want  string
	}{
		{"hello", "hello"},
		{[]byte("world"), "world"},
		{42, "42"},
		{int8(8), "8"},
		{int16(16), "16"},
		{int32(32), "32"},
		{int64(64), "64"},
		{uint(1), "1"},
		{uint8(8), "8"},
		{uint16(16), "16"},
		{uint32(32), "32"},
		{uint64(64), "64"},
		{3.14, "3.14"},
		{float32(1.5), "1.5"},
		{true, "true"},
		{false, "false"},
		{nil, "<nil>"},
	}

	for _, tc := range tests {
		got := asString(tc.input)
		if got != tc.want {
			t.Errorf("asString(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestTrimZeroDecimal(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123.00", "123"},
		{"123.0", "123"},
		{"123.10", "123.10"},
		{"123.456", "123.456"},
		{"123", "123"},
		{"0.00", "0"},
		{"", ""},
		{".00", ""},
	}

	for _, tc := range tests {
		got := trimZeroDecimal(tc.input)
		if got != tc.want {
			t.Errorf("trimZeroDecimal(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSetValue(t *testing.T) {
	strType := reflect.TypeOf("")
	intType := reflect.TypeOf(0)

	t.Run("non-ptr", func(t *testing.T) {
		dest := reflect.New(strType).Elem()
		setValue(false, dest, "hello", strType)
		if dest.String() != "hello" {
			t.Errorf("got %v, want hello", dest.String())
		}
	})

	t.Run("ptr", func(t *testing.T) {
		dest := reflect.New(reflect.PointerTo(strType)).Elem()
		setValue(true, dest, "hello", strType)
		if dest.Elem().String() != "hello" {
			t.Errorf("got %v, want hello", dest.Elem().String())
		}
	})

	t.Run("int-to-float", func(t *testing.T) {
		dest := reflect.New(intType).Elem()
		setValueConvert(false, dest, int64(42), intType)
		if dest.Int() != 42 {
			t.Errorf("got %v, want 42", dest.Int())
		}
	})
}

func TestConvertAssign(t *testing.T) {
	opts := &Options{}
	opts.tidy()
	o := &osmBase{options: opts}

	t.Run("string to string", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf("")).Elem()
		err := o.convertAssign("test", dest, "hello", false, reflect.TypeOf(""))
		if err != nil {
			t.Fatal(err)
		}
		if dest.String() != "hello" {
			t.Errorf("got %q, want %q", dest.String(), "hello")
		}
	})

	t.Run("string to []byte", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf([]byte{})).Elem()
		err := o.convertAssign("test", dest, "hello", false, reflect.TypeOf([]byte{}))
		if err != nil {
			t.Fatal(err)
		}
		if string(dest.Bytes()) != "hello" {
			t.Errorf("got %q, want %q", string(dest.Bytes()), "hello")
		}
	})

	t.Run("[]byte to string", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf("")).Elem()
		err := o.convertAssign("test", dest, []byte("world"), false, reflect.TypeOf(""))
		if err != nil {
			t.Fatal(err)
		}
		if dest.String() != "world" {
			t.Errorf("got %q, want %q", dest.String(), "world")
		}
	})

	t.Run("int to int64", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf(int64(0))).Elem()
		err := o.convertAssign("test", dest, 42, false, reflect.TypeOf(int64(0)))
		if err != nil {
			t.Fatal(err)
		}
		if dest.Int() != 42 {
			t.Errorf("got %d, want 42", dest.Int())
		}
	})

	t.Run("float to float64", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf(float64(0))).Elem()
		err := o.convertAssign("test", dest, 3.14, false, reflect.TypeOf(float64(0)))
		if err != nil {
			t.Fatal(err)
		}
		if dest.Float() != 3.14 {
			t.Errorf("got %f, want 3.14", dest.Float())
		}
	})

	t.Run("string to bool", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf(false)).Elem()
		err := o.convertAssign("test", dest, "true", false, reflect.TypeOf(false))
		if err != nil {
			t.Fatal(err)
		}
		if dest.Bool() != true {
			t.Errorf("got %v, want true", dest.Bool())
		}
	})

	t.Run("time.Time to string", func(t *testing.T) {
		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		dest := reflect.New(reflect.TypeOf("")).Elem()
		err := o.convertAssign("test", dest, now, false, reflect.TypeOf(""))
		if err != nil {
			t.Fatal(err)
		}
		if dest.String() == "" {
			t.Error("expected non-empty time string")
		}
	})

	t.Run("string to time.Time", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf(time.Time{})).Elem()
		err := o.convertAssign("test", dest, "2024-01-15 10:30:00", false, reflect.TypeOf(time.Time{}))
		if err != nil {
			t.Fatal(err)
		}
		if dest.Interface().(time.Time).IsZero() {
			t.Error("expected non-zero time")
		}
	})

	t.Run("nil to ptr", func(t *testing.T) {
		strType := reflect.TypeOf("")
		ptrType := reflect.PointerTo(strType)
		dest := reflect.New(ptrType).Elem()
		err := o.convertAssign("test", dest, nil, true, ptrType)
		if err != nil {
			t.Fatal(err)
		}
		if dest.IsNil() {
			t.Error("expected non-nil pointer after nil conversion")
		}
		if dest.Elem().String() != "" {
			t.Errorf("expected empty string, got %q", dest.Elem().String())
		}
	})

	t.Run("parse error returns error", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf(int64(0))).Elem()
		err := o.convertAssign("test", dest, "not_a_number", false, reflect.TypeOf(int64(0)))
		if err == nil {
			t.Error("expected error for invalid int")
		}
	})

	t.Run("parse error for uint returns error", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf(uint64(0))).Elem()
		err := o.convertAssign("test", dest, "-1", false, reflect.TypeOf(uint64(0)))
		if err == nil {
			t.Error("expected error for negative uint")
		}
	})

	t.Run("parse error for float returns error", func(t *testing.T) {
		dest := reflect.New(reflect.TypeOf(float64(0))).Elem()
		err := o.convertAssign("test", dest, "not_float", false, reflect.TypeOf(float64(0)))
		if err == nil {
			t.Error("expected error for invalid float")
		}
	})
}

func TestAsBytes(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		b, ok := asBytes(nil, reflect.ValueOf(42))
		if !ok || string(b) != "42" {
			t.Errorf("got %q, want 42", string(b))
		}
	})

	t.Run("string", func(t *testing.T) {
		b, ok := asBytes(nil, reflect.ValueOf("hello"))
		if !ok || string(b) != "hello" {
			t.Errorf("got %q, want hello", string(b))
		}
	})

	t.Run("bool", func(t *testing.T) {
		b, ok := asBytes(nil, reflect.ValueOf(true))
		if !ok || string(b) != "true" {
			t.Errorf("got %q, want true", string(b))
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, ok := asBytes(nil, reflect.ValueOf([]int{1, 2, 3}))
		if ok {
			t.Error("expected false for unsupported type")
		}
	})
}
