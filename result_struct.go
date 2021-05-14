package osm

import (
	"fmt"
	"reflect"
)

func resultStruct(o *osmBase, id, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
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
		elementTypes := make([]reflect.Type, columnsCount)
		isPtrs := make([]bool, columnsCount)
		values := make([]reflect.Value, columnsCount)

		structType := valueElem.Type()
		allFieldNameTypeMap := map[string]*reflect.Type{} // struct每个成员的名字，不一定与sql中的列对应
		getStructFieldMap(structType, allFieldNameTypeMap)

		for i, col := range columns {
			filedName, t := findFiled(allFieldNameTypeMap, col)
			if filedName != "" && t != nil {
				f := valueElem.FieldByName(filedName)
				elementTypes[i] = f.Type()
				isPtrs[i] = elementTypes[i].Kind() == reflect.Ptr
				values[i] = f
			} else {
				a := ""
				elementTypes[i] = reflect.TypeOf(a)
				values[i] = reflect.ValueOf(&a).Elem()
			}
		}
		err = scanRow(rows, isPtrs, elementTypes, values)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
		if isStructPtr {
			value.Set(valueElem.Addr())
		}
	}

	return 1, nil
}

func getStructFieldMap(t reflect.Type, m map[string]*reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		t := t.Field(i)
		if t.Anonymous && t.Type.Kind() == reflect.Struct {
			getStructFieldMap(t.Type, m)
			continue
		}
		m[t.Name] = &(t.Type)
	}
}
