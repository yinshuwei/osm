package osm

import (
	"fmt"
	"reflect"
)

func resultStruct(logPrefix string, o *osmBase, id, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("sql '%s' error : struct类型Query，查询结果类型应为struct的指针，而您传入的并不是指针", id)
	}
	value := reflect.Indirect(pointValue)
	valueElem := value
	isStructPtr := value.Kind() == reflect.Ptr
	if isStructPtr {
		valueElem = reflect.New(value.Type().Elem()).Elem()
	}
	if valueElem.Kind() != reflect.Struct {
		return 0, fmt.Errorf("sql '%s' error : struct类型Query，查询结果类型应为struct的指针，而您传入的并不是struct", id)
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
		fields := make([]*structFieldInfo, columnsCount) // struct成员的名字，与sql中的列对应
		values := make([]reflect.Value, columnsCount)

		structType := valueElem.Type()
		tagMap := make(map[string]*structFieldInfo)
		nameMap := make(map[string]*structFieldInfo)
		getStructFieldMap(structType, tagMap, nameMap, false)

		for i, col := range columns {
			field := findFiled(tagMap, nameMap, col)
			fields[i] = field
			if field != nil {
				if field.a {
					values[i] = valueElem.FieldByName(field.n)
				} else {
					values[i] = valueElem.Field(field.i)
				}
			} else {
				a := ""
				values[i] = reflect.ValueOf(&a).Elem()
			}
		}
		err = o.scanRow(logPrefix, rows, fields, values)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
		if isStructPtr {
			value.Set(valueElem.Addr())
		}
	}

	return 1, nil
}
