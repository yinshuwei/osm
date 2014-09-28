package osm

//Data组合了interface{},从sql中查询到的数据都可以放入在Data对象中,可以通过它的转换方法来还原数据
//
//如
//  var userMaps []map[string]osm.Data
//  o.Select("selectUserMaps", user)(&userMaps)
//  for _, uMap := range userMaps {
//    fmt.Println(uMap["Id"].Int64(), uMap["Email"].String())
//  }
//以上是用maps类型进行数据查询的例子,当然这里可以用[]map[string]inferface{}类型做为结果,只是Data类型多了一些转换方法而已
type Data struct {
	d interface{}
}

// osmData to string
func (data Data) String() string {
	bsv, ok := data.d.([]byte)
	if ok {
		return string(bsv)
	}
	return ""
}

// osmData to bool
func (data Data) Bool() bool {
	if data.d == nil {
		return false
	}
	return data.d.(bool)
}

// osmData to float32
func (data Data) Float32() float32 {
	if data.d == nil {
		return 0.0
	}
	return data.d.(float32)
}

// osmData to float64
func (data Data) Float64() float64 {
	if data.d == nil {
		return 0.0
	}
	return data.d.(float64)
}

// osmData to int
func (data Data) Int() int {
	if data.d == nil {
		return 0
	}
	return data.d.(int)
}

// osmData to int8
func (data Data) Int8() int8 {
	if data.d == nil {
		return 0
	}
	return data.d.(int8)
}

// osmData to int16
func (data Data) Int16() int16 {
	if data.d == nil {
		return 0
	}
	return data.d.(int16)
}

// osmData to int32
func (data Data) Int32() int32 {
	if data.d == nil {
		return 0
	}
	return data.d.(int32)
}

// osmData to int64
func (data Data) Int64() int64 {
	if data.d == nil {
		return 0
	}
	return data.d.(int64)
}

// osmData to uint
func (data Data) Uint() uint {
	if data.d == nil {
		return 0
	}
	return data.d.(uint)
}

// osmData to uint8
func (data Data) Uint8() uint8 {
	if data.d == nil {
		return 0
	}
	return data.d.(uint8)
}

// osmData to uint16
func (data Data) Uint16() uint16 {
	if data.d == nil {
		return 0
	}
	return data.d.(uint16)
}

// osmData to uint31
func (data Data) Uint32() uint32 {
	if data.d == nil {
		return 0
	}
	return data.d.(uint32)
}

// osmData to uint64
func (data Data) Uint64() uint64 {
	if data.d == nil {
		return 0
	}
	return data.d.(uint64)
}

// osmData to interface{}
func (data Data) Data() interface{} {
	return data.d
}
