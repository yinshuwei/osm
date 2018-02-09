package osm

import (
	"fmt"
	"reflect"
)

func resultStruct(o *osmBase, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("struct类型Query，查询结果类型应为struct的指针，而您传入的并不是指针")
	}
	value := reflect.Indirect(pointValue)
	valueElem := value
	isStructPtr := value.Kind() == reflect.Ptr
	if isStructPtr {
		valueElem = reflect.New(value.Type().Elem()).Elem()
	}
	if valueElem.Kind() != reflect.Struct {
		return 0, fmt.Errorf("struct类型Query，查询结果类型应为struct的指针，而您传入的并不是struct")
	}

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}
		columnsCount := len(columns)
		elementTypes := make([]reflect.Type, columnsCount)
		isPtrs := make([]bool, columnsCount)
		values := make([]reflect.Value, columnsCount)
		allFieldNameTypeMap := map[string]*reflect.Type{} // struct每个成员的名字，不一定与sql中的列对应
		structType := valueElem.Type()
		for i := 0; i < structType.NumField(); i++ {
			t := structType.Field(i)
			allFieldNameTypeMap[t.Name] = &(t.Type)
		}
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
			return 0, err
		}
		if isStructPtr {
			value.Set(valueElem.Addr())
		}
	}

	return 1, nil
}
