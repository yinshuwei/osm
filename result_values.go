package osm

import (
	"fmt"
	"reflect"
)

func resultValues(o *osmBase, sql string, sqlParams []interface{}, containers []interface{}) (int64, error) {
	lenContainers := len(containers)
	values := make([]reflect.Value, lenContainers)
	elementTypes := make([]reflect.Type, lenContainers)
	isPtrs := make([]bool, lenContainers)
	for i, container := range containers {
		pointValue := reflect.ValueOf(container)
		if pointValue.Kind() != reflect.Ptr {
			return 0, fmt.Errorf("values类型Query，查询结果类型应为slice的指针，而您传入的第%d个并不是指针", i+1)
		}
		value := reflect.Indirect(pointValue)
		if value.Kind() != reflect.Slice {
			return 0, fmt.Errorf("values类型Query，查询结果类型应为slice的指针，而您传入的第%d个并不是slice", i+1)
		}
		values[i] = value
		elementTypes[i] = value.Type().Elem()
		isPtrs[i] = elementTypes[i].Kind() == reflect.Ptr
	}

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var rowsCount int64
	var lenColumn int
	for rows.Next() {
		if rowsCount == 0 {
			columns, err := rows.Columns()
			if err != nil {
				return 0, err
			}
			lenColumn = len(columns)
			if lenColumn != lenContainers {
				return 0, fmt.Errorf("values类型Query，查询结果的长度与SQL的长度不一致")
			}
		}
		objs := make([]reflect.Value, lenContainers)
		for i := 0; i < lenContainers; i++ {
			objs[i] = reflect.New(elementTypes[i]).Elem()
		}
		err = scanRow(rows, isPtrs, elementTypes, objs)
		if err != nil {
			return 0, err
		}
		for i := 0; i < lenContainers; i++ {
			values[i].Set(reflect.Append(values[i], objs[i]))
		}
		rowsCount++
	}

	return rowsCount, nil
}
