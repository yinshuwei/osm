package osm

import (
	"fmt"
	"reflect"
)

func resultKvs(o *osmBase, sql string, sqlParams []interface{}, container interface{}) (int64, error) {

	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("query kvs all args must be use ptr")
	}

	value := reflect.Indirect(pointValue)
	if value.Kind() != reflect.Map {
		return 0, fmt.Errorf("query kvs args must be use map")
	}

	cType := value.Type()
	keyType := cType.Key()
	valType := cType.Elem()

	valueNew := reflect.MakeMap(cType)

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var rowsCount int64

	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}

		if len(columns) < 2 {
			return 0, fmt.Errorf("kvs len(columns)<2")
		}

		refs := make([]interface{}, len(columns))
		for i, _ := range columns {
			var ref interface{}
			refs[i] = &ref
		}

		if err := rows.Scan(refs...); err != nil {
			return 0, err
		}

		key := reflect.Indirect(reflect.New(keyType))
		val := reflect.Indirect(reflect.New(valType))

		setDataToValue(key, reflect.ValueOf(refs[0]).Elem().Interface())
		setDataToValue(val, reflect.ValueOf(refs[1]).Elem().Interface())
		valueNew.SetMapIndex(key, val)

		rowsCount++
	}

	value.Set(valueNew)

	return rowsCount, nil
}
