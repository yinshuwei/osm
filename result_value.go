package osm

import (
	"fmt"
	"reflect"
)

func resultValue(logPrefix string, o *osmBase, id, sql string, sqlParams []interface{}, containers []interface{}) (int64, error) {
	lenContainers := len(containers)
	values := make([]reflect.Value, lenContainers)
	fields := make([]*structFieldInfo, lenContainers)
	for i, container := range containers {
		pointValue := reflect.ValueOf(container)
		if pointValue.Kind() != reflect.Ptr {
			return 0, fmt.Errorf("sql '%s' error : value类型Query，查询结果类型应为指针，而您传入的第%d个并不是指针", id, i+1)
		}
		value := reflect.Indirect(pointValue)
		values[i] = value

		valueType := value.Type()
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
	if rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
		columnsCount := len(columns)
		if columnsCount != lenContainers {
			return 0, fmt.Errorf("sql '%s' error : value类型Query，查询结果的长度与SQL的长度不一致", id)
		}

		err = o.scanRow(logPrefix, rows, fields, values)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
	}

	return 1, nil
}
