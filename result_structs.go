package osm

import (
	"fmt"
	"reflect"
)

func resultStructs(o *osmBase, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("structs类型Query，查询结果类型应为struct数组的指针，而您传入的并不是指针")
	}

	value := reflect.Indirect(pointValue)
	if value.Kind() != reflect.Slice {
		return 0, fmt.Errorf("structs类型Query，查询结果类型应为struct数组的指针，而您传入的并不是数组")
	}

	rowType := value.Type().Elem()
	isStructPtr := rowType.Kind() == reflect.Ptr
	structType := rowType
	if isStructPtr {
		structType = rowType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return 0, fmt.Errorf("structs类型Query，查询结果类型应为struct数组的指针，而您传入的并不是struct")
	}

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var rowsCount int64
	var lenColumn int
	var elementTypes []reflect.Type
	var isPtrs []bool
	var fileNames []string
	for rows.Next() {
		valueElem := reflect.New(structType).Elem()
		if isPtrs == nil {
			columns, err := rows.Columns()
			if err != nil {
				return 0, err
			}
			lenColumn = len(columns)
			elementTypes = make([]reflect.Type, lenColumn)
			isPtrs = make([]bool, lenColumn)
			fileNames = make([]string, lenColumn)
			for i, col := range columns {
				fileNames[i] = toGoName(col)
				f := valueElem.FieldByName(fileNames[i])
				elementTypes[i] = f.Type()
				isPtrs[i] = elementTypes[i].Kind() == reflect.Ptr
			}
		}
		values := make([]reflect.Value, lenColumn)
		for i, fileName := range fileNames {
			f := valueElem.FieldByName(fileName)
			values[i] = f
		}
		err = scanRow(rows, isPtrs, elementTypes, values)
		if err != nil {
			return 0, err
		}
		if isStructPtr {
			value.Set(reflect.Append(value, valueElem.Addr()))
		} else {
			value.Set(reflect.Append(value, valueElem))
		}
		rowsCount++
	}
	return rowsCount, nil
}
