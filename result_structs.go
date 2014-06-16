package osm

import (
	"fmt"
	"reflect"
)

func resultStructs(o *osmBase, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("Select()() all args must be use ptr")
	}

	value := reflect.Indirect(pointValue)
	valueNew := value
	elementType := value.Type().Elem()

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

		if rowsCount == 0 && !value.IsNil() {
			value.Set(reflect.New(value.Type()).Elem())
		}

		var obj reflect.Value
		if elementType.Kind() == reflect.Ptr {
			obj = reflect.New(elementType.Elem())
		} else {
			obj = reflect.New(elementType)
		}

		if obj.Kind() == reflect.Ptr {
			obj = obj.Elem()
		}

		for i := 0; i < obj.NumField(); i++ {
			f := obj.Field(i)
			fe := obj.Type().Field(i)

			if v, ok := fieldValueMap[fe.Name]; ok {
				vv := reflect.ValueOf(v).Elem().Interface()
				setDataToValue(f, vv)
			}
		}

		if elementType.Kind() == reflect.Ptr {
			obj = obj.Addr()
		}

		valueNew = reflect.Append(valueNew, obj)

		rowsCount++
	}

	value.Set(valueNew)

	return rowsCount, nil
}
