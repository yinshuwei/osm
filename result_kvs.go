package osm

import (
	"fmt"
	"reflect"
)

func resultKvs(o *osmBase, id, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("sql '%s' error : kvs类型Query，查询结果类型应为map的指针，而您传入的并不是指针", id)
	}
	value := reflect.Indirect(pointValue)
	if value.Kind() != reflect.Map {
		return 0, fmt.Errorf("sql '%s' error : kvs类型Query，查询结果类型应为map的指针，而您传入的并不是map", id)
	}
	cType := value.Type()
	if value.IsNil() {
		value.Set(reflect.MakeMap(cType))
	}
	elementTypes := []reflect.Type{cType.Key(), cType.Elem()}
	isPtrs := []bool{elementTypes[0].Kind() == reflect.Ptr, elementTypes[1].Kind() == reflect.Ptr}
	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
	}
	defer rows.Close()
	var rowsCount int64
	for rows.Next() {
		if rowsCount == 0 {
			columns, err := rows.Columns()
			if err != nil {
				return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
			}
			if len(columns) != 2 {
				return 0, fmt.Errorf("sql '%s' error : kvs类型Query，SQL查询的结果需要为2列", id)
			}
		}
		objs := []reflect.Value{
			reflect.New(elementTypes[0]).Elem(),
			reflect.New(elementTypes[1]).Elem(),
		}
		err = o.scanRow(rows, isPtrs, elementTypes, objs)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
		value.SetMapIndex(objs[0], objs[1])
		rowsCount++
	}
	return rowsCount, nil
}
