package osm

import (
	"reflect"
	"testing"
	"time"
)

var toGoNamesTestDatas = [][]string{
	{"_", "", ""},
	{"_ID_", "Id", "ID"},
	{"Card_ID_", "CardId", "CardID"},
	{"_ID_name_", "IdName", "IDName"},
	{"id", "Id", "ID"},
	{"hahah_url_id_aaaaa_xss_bb", "HahahUrlIdAaaaaXssBb", "HahahURLIDAaaaaXSSBb"},
	{"foo_bar", "FooBar", "FooBar"},
	{"foo_bar_baz", "FooBarBaz", "FooBarBaz"},
	{"Foo_bar", "FooBar", "FooBar"},
	{"foo_WiFi", "FooWifi", "FooWifi"},
	{"Id", "Id", "ID"},
	{"foo_id", "FooId", "FooID"},
	{"fooId", "Fooid", "Fooid"},
	{"_Leading", "Leading", "Leading"},
	{"___Leading", "Leading", "Leading"},
	{"trailing_", "Trailing", "Trailing"},
	{"trailing___", "Trailing", "Trailing"},
	{"a_b", "AB", "AB"},
	{"a__b", "AB", "AB"},
	{"a___b", "AB", "AB"},
	{"Rpc1150", "Rpc1150", "Rpc1150"},
	{"case3_1", "Case31", "Case31"},
	{"case3__1", "Case31", "Case31"},
	{"IEEE802_16bit", "Ieee80216bit", "Ieee80216bit"},
	{"IEEE802_16Bit", "Ieee80216bit", "Ieee80216bit"},
	{"Uid", "Uid", "UID"},
	{"UUId", "Uuid", "UUID"},
	{"Uid_121_abd", "Uid121Abd", "UID121Abd"},
	{"a_UUId_b", "AUuidB", "AUUIDB"},
	{"AAA__Uid", "AaaUid", "AaaUID"},
	{"AA_DDD_UUId_12", "AaDddUuid12", "AaDddUUID12"},
}

func TestAll(t *testing.T) {
	TestToGoNames(t)
}

func TestToGoNames(t *testing.T) {
	for _, words := range toGoNamesTestDatas {
		a, b := toGoNames(words[0])
		if a != words[1] {
			t.Errorf("普通方式转换错误,\"%s\"->\"%s\"", words[0], a)
		}
		if b != words[2] {
			t.Errorf("特珠字符大写方式转换错误,\"%s\"->\"%s\"", words[0], b)
		}
	}
}

func TestGetFieldMap(t *testing.T) {
	type A struct {
		Name string `db:"111name222"`
	}

	type B struct {
		ID int
		A
	}
	type C struct {
		B
		Age int64
	}

	b := reflect.TypeOf(&C{}).Elem()
	tagMap := make(map[string]*structFieldInfo)
	nameMap := make(map[string]*structFieldInfo)

	getStructFieldMap(b, tagMap, nameMap, false)
	for k, v := range tagMap {
		t.Log(k, v.i, v.n, v.t, v.a, v.isPtr)
	}
	for k, v := range nameMap {
		t.Log(k, v.i, v.n, v.t, v.a, v.isPtr)
	}
}

func TestIsValueKind(t *testing.T) {
	tests := []struct {
		kind reflect.Kind
		want bool
	}{
		{reflect.Bool, true},
		{reflect.Int, true},
		{reflect.Int8, true},
		{reflect.Int16, true},
		{reflect.Int32, true},
		{reflect.Int64, true},
		{reflect.Uint, true},
		{reflect.Uint8, true},
		{reflect.Uint16, true},
		{reflect.Uint32, true},
		{reflect.Uint64, true},
		{reflect.Uintptr, true},
		{reflect.Float32, true},
		{reflect.Float64, true},
		{reflect.Complex64, true},
		{reflect.Complex128, true},
		{reflect.String, true},
		{reflect.Struct, true},
		{reflect.Slice, false},
		{reflect.Map, false},
		{reflect.Array, false},
		{reflect.Ptr, false},
		{reflect.Interface, false},
		{reflect.Func, false},
		{reflect.Chan, false},
	}

	for _, tc := range tests {
		got := isValueKind(tc.kind)
		if got != tc.want {
			t.Errorf("isValueKind(%s) = %v, want %v", tc.kind, got, tc.want)
		}
	}
}

func TestIsNativeParamType(t *testing.T) {
	tests := []struct {
		kind reflect.Kind
		want bool
	}{
		{reflect.Bool, true},
		{reflect.Int, true},
		{reflect.Int8, true},
		{reflect.Int16, true},
		{reflect.Int32, true},
		{reflect.Int64, true},
		{reflect.Uint, true},
		{reflect.Uint8, true},
		{reflect.Uint16, true},
		{reflect.Uint32, true},
		{reflect.Uint64, true},
		{reflect.Uintptr, true},
		{reflect.Float32, true},
		{reflect.Float64, true},
		{reflect.Complex64, true},
		{reflect.Complex128, true},
		{reflect.String, true},
		{reflect.Struct, false},
		{reflect.Slice, false},
		{reflect.Map, false},
		{reflect.Array, false},
		{reflect.Ptr, false},
		{reflect.Interface, false},
	}

	for _, tc := range tests {
		got := isNativeParamType(tc.kind)
		if got != tc.want {
			t.Errorf("isNativeParamType(%s) = %v, want %v", tc.kind, got, tc.want)
		}
	}
}

func TestFindField(t *testing.T) {
	strType := reflect.TypeOf("")
	intType := reflect.TypeOf(0)

	tagMap := map[string]*structFieldInfo{
		"db_name": {i: 0, n: "Name", t: &strType},
		"db_age":  {i: 1, n: "Age", t: &intType},
	}

	nameMap := map[string]*structFieldInfo{
		"Name": {i: 0, n: "Name", t: &strType},
		"Age":  {i: 1, n: "Age", t: &intType},
	}

	t.Run("tag match", func(t *testing.T) {
		f := findField(tagMap, nameMap, "db_name")
		if f == nil || f.n != "Name" {
			t.Errorf("expected Name, got nil")
		}
	})

	t.Run("name match", func(t *testing.T) {
		f := findField(tagMap, nameMap, "Name")
		if f == nil || f.n != "Name" {
			t.Errorf("expected Name, got nil")
		}
	})

	t.Run("no match returns nil", func(t *testing.T) {
		f := findField(tagMap, nameMap, "nonexistent")
		if f != nil {
			t.Errorf("expected nil, got %v", f)
		}
	})
}

func TestTimeFormat(t *testing.T) {
	now := time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)

	got := timeFormat(now, formatDate)
	if got != "2024-06-15" {
		t.Errorf("timeFormat date: got %q, want %q", got, "2024-06-15")
	}

	got = timeFormat(now, formatDateTime)
	if got != "2024-06-15 10:30:45" {
		t.Errorf("timeFormat datetime: got %q, want %q", got, "2024-06-15 10:30:45")
	}
}

func TestSetDataToParamName(t *testing.T) {
	t.Run("non-IN scalar", func(t *testing.T) {
		frag := &sqlFragment{content: "id", isParam: true, isIn: false}
		setDataToParamName(frag, reflect.ValueOf(42))
		if frag.paramValue != 42 {
			t.Errorf("got %v, want 42", frag.paramValue)
		}
	})

	t.Run("IN slice", func(t *testing.T) {
		frag := &sqlFragment{content: "ids", isParam: true, isIn: true}
		setDataToParamName(frag, reflect.ValueOf([]int{1, 2, 3}))
		if len(frag.paramValues) != 3 || frag.paramValues[0] != 1 || frag.paramValues[1] != 2 || frag.paramValues[2] != 3 {
			t.Errorf("got %v", frag.paramValues)
		}
	})

	t.Run("IN single value", func(t *testing.T) {
		frag := &sqlFragment{content: "id", isParam: true, isIn: true}
		setDataToParamName(frag, reflect.ValueOf(42))
		if len(frag.paramValues) != 1 || frag.paramValues[0] != 42 {
			t.Errorf("got %v", frag.paramValues)
		}
	})

	t.Run("time.Time non-IN", func(t *testing.T) {
		now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
		frag := &sqlFragment{content: "t", isParam: true, isIn: false}
		setDataToParamName(frag, reflect.ValueOf(now))
		if frag.paramValue.(string) != "2024-06-15 10:30:00" {
			t.Errorf("got %v, want formatted time", frag.paramValue)
		}
	})
}
