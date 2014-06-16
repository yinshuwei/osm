package osm

import (
	"fmt"
	"reflect"
)

func resultStruct(o *osmBase, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		panic(fmt.Errorf("Select()() all args must be use ptr"))
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

		columnsMp := make(map[string]interface{}, len(columns))

		refs := make([]interface{}, len(columns))
		for i, col := range columns {
			var ref interface{}
			columnsMp[toGoName(col)] = &ref
			refs[i] = &ref
		}

		if err := rows.Scan(refs...); err != nil {
			return 0, err
		}

		for fieldName, v := range columnsMp {
			f := value.FieldByName(fieldName)

			vv := reflect.ValueOf(v).Elem().Interface()
			setDataToValue(f, vv)
		}

		rowsCount++
	}

	return rowsCount, nil
}
