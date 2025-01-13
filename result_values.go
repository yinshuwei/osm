package osm

import (
	"fmt"
	"reflect"
)

func resultValues(logPrefix string, o *osmBase, id, sql string, sqlParams []interface{}, containers []interface{}) (int64, error) {
	lenContainers := len(containers)
	values := make([]reflect.Value, lenContainers)
	// elementTypes := make([]reflect.Type, lenContainers)
	// isPtrs := make([]bool, lenContainers)
	fields := make([]*structFieldInfo, lenContainers)
	for i, container := range containers {
		pointValue := reflect.ValueOf(container)
		if pointValue.Kind() != reflect.Ptr {
			return 0, fmt.Errorf("sql '%s' error : values类型Query，查询结果类型应为slice的指针，而您传入的第%d个并不是指针", id, i+1)
		}
		value := reflect.Indirect(pointValue)
		if value.Kind() != reflect.Slice {
			return 0, fmt.Errorf("sql '%s' error : values类型Query，查询结果类型应为slice的指针，而您传入的第%d个并不是slice", id, i+1)
		}
		values[i] = value
		valueType := value.Type().Elem()
		kind := valueType.Kind()
		field := &structFieldInfo{t: &valueType, isPtr: kind == reflect.Ptr}
		if field.isPtr {
			kind = valueType.Elem().Kind()
		}
		fields[i] = field
		if !isValueKind(kind) {
			return 0, fmt.Errorf("sql '%s' error : value类型Query，查询结果类型应为Bool,Int,Int8,Int16,Int32,Int64,Uint,Uint8,Uint16,Uint32,Uint64,Uintptr,Float32,Float64,Complex64,Complex128,String,Time，而您传入的第%d个并不是", id, i+1)
		}
	}

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
	}
	defer rows.Close()
	var rowsCount int64
	var columnsCount int
	for rows.Next() {
		if rowsCount == 0 {
			columns, err := rows.Columns()
			if err != nil {
				return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
			}
			columnsCount = len(columns)
			if columnsCount != lenContainers {
				return 0, fmt.Errorf("sql '%s' error : values类型Query，查询结果的长度与SQL的长度不一致", id)
			}
		}
		objs := make([]reflect.Value, lenContainers)
		for i := 0; i < lenContainers; i++ {
			objs[i] = reflect.New(*(fields[i].t)).Elem()
		}
		err = o.scanRow(logPrefix, rows, fields, objs)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
		for i := 0; i < lenContainers; i++ {
			values[i].Set(reflect.Append(values[i], objs[i]))
		}
		rowsCount++
	}

	return rowsCount, nil
}
