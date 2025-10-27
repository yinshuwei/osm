package osm

import (
	"fmt"
	"reflect"
)

func resultKvs(logPrefix string, o *osmBase, id, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
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

	kType := cType.Key()
	vType := cType.Elem()
	fields := []*structFieldInfo{
		{0, "", &kType, false, kType.Kind() == reflect.Ptr},
		{0, "", &vType, false, vType.Kind() == reflect.Ptr},
	}

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
	}
	defer rows.Close()
	var rowsCount int64
	for rows.Next() {
		if rowsCount == 0 {
			columns, err1 := rows.Columns()
			if err1 != nil {
				return 0, fmt.Errorf("sql '%s' error : %s", id, err1.Error())
			}
			if len(columns) != 2 {
				return 0, fmt.Errorf("sql '%s' error : kvs类型Query，SQL查询的结果需要为2列", id)
			}
		}
		objs := []reflect.Value{
			reflect.New(*(fields[0].t)).Elem(),
			reflect.New(*(fields[1].t)).Elem(),
		}
		err = o.scanRow(logPrefix, rows, fields, objs)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
		value.SetMapIndex(objs[0], objs[1])
		rowsCount++
	}
	return rowsCount, nil
}
