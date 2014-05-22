package osm

import (
	"fmt"
	"reflect"
)

func resultMap(o *osmBase, sql string, sqlParams []interface{}, container interface{}) (int64, error) {

	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("Query()() all args must be use ptr")
	}

	value := reflect.Indirect(pointValue)

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var rowsCount int64

	if rows.Next() {

		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}

		fieldValueMap := make(map[string]interface{}, len(columns))

		refs := make([]interface{}, len(columns))
		for i, col := range columns {
			var ref interface{}
			fieldValueMap[toGoName(col)] = &ref
			refs[i] = &ref
		}

		if err := rows.Scan(refs...); err != nil {
			return 0, err
		}

		valueNew := make(map[string]Data, len(refs))
		for k, v := range fieldValueMap {
			vv := reflect.ValueOf(v).Elem().Interface()
			valueNew[k] = Data{d: vv}
		}

		rowsCount++

		value.Set(reflect.ValueOf(valueNew))
	}

	return rowsCount, nil
}
