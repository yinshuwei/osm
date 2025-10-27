package osm

import (
	"fmt"
	"reflect"
)

var stringType = reflect.TypeOf("")

func checkColumns(columns interface{}) *reflect.Value {
	pointValue := reflect.ValueOf(columns)
	if pointValue.Kind() != reflect.Ptr {
		return nil
	}
	value := reflect.Indirect(pointValue)
	if value.Kind() != reflect.Slice {
		return nil
	}
	eleType := value.Type().Elem()
	kind := eleType.Kind()
	if kind != reflect.String {
		return nil
	}
	return &value
}

func checkDatas(datas interface{}) *reflect.Value {
	pointValue := reflect.ValueOf(datas)
	if pointValue.Kind() != reflect.Ptr {
		return nil
	}
	value := reflect.Indirect(pointValue)
	if value.Kind() != reflect.Slice {
		return nil
	}
	eleType := value.Type().Elem()
	kind := eleType.Kind()
	if kind != reflect.Slice {
		return nil
	}

	if eleType.Elem().Kind() != reflect.String {
		return nil
	}

	return &value
}

// resultStrings 数据库结果读入到columns，和datas。columns为[]string，datas为[][]string。
func resultStrings(logPrefix string, o *osmBase, id, sql string, sqlParams []interface{}, columnsContainer, datasContainer interface{}) (int64, error) {
	columnsValue := checkColumns(columnsContainer)
	if columnsValue == nil {
		return 0, fmt.Errorf("sql '%s' error : strings类型Query，查询结果类型第一个为[]string的指针，第二个为[][]string的指针", id)
	}
	datasValue := checkDatas(datasContainer)
	if datasValue == nil {
		return 0, fmt.Errorf("sql '%s' error : strings类型Query，查询结果类型第一个为[]string的指针，第二个为[][]string的指针", id)
	}

	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
	}
	defer rows.Close()

	var rowsCount int64
	var columnsCount int
	var fields []*structFieldInfo
	for rows.Next() {
		if rowsCount == 0 {
			columns, err1 := rows.Columns()
			if err1 != nil {
				return 0, fmt.Errorf("sql '%s' error : %s", id, err1.Error())
			}
			columnsCount = len(columns)
			for _, column := range columns {
				fields = append(fields, &structFieldInfo{0, "", &stringType, false, false})
				(*columnsValue).Set(reflect.Append(*columnsValue, reflect.ValueOf(column)))
			}
		}
		objs := make([]reflect.Value, columnsCount)
		for i := 0; i < columnsCount; i++ {
			objs[i] = reflect.New(stringType).Elem()
		}
		err = o.scanRow(logPrefix, rows, fields, objs)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
		var data []string
		for i := 0; i < columnsCount; i++ {
			data = append(data, objs[i].String())
		}
		(*datasValue).Set(reflect.Append(*datasValue, reflect.ValueOf(data)))
		rowsCount++
	}

	return rowsCount, nil
}
