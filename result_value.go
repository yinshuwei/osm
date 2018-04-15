package osm

import (
	"fmt"
	"reflect"
)

func resultValue(o *osmBase, id, sql string, sqlParams []interface{}, containers []interface{}) (int64, error) {
	lenContainers := len(containers)
	values := make([]reflect.Value, lenContainers)
	elementTypes := make([]reflect.Type, lenContainers)
	isPtrs := make([]bool, lenContainers)
	for i, container := range containers {
		pointValue := reflect.ValueOf(container)
		if pointValue.Kind() != reflect.Ptr {
			return 0, fmt.Errorf("sql '%s' error : value类型Query，查询结果类型应为指针，而您传入的第%d个并不是指针", id, i+1)
		}
		value := reflect.Indirect(pointValue)
		values[i] = value
		elementTypes[i] = value.Type()
		isPtrs[i] = elementTypes[i].Kind() == reflect.Ptr
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

		err = scanRow(rows, isPtrs, elementTypes, values)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
	}

	return 1, nil
}
