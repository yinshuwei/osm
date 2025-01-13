package osm

import (
	"fmt"
	"reflect"
)

// resultStructs 数据库结果读入到struct切片中，struct可以是指针类型或非指针类型
func resultStructs(logPrefix string, o *osmBase, id, sql string, sqlParams []interface{}, container interface{}) (int64, error) {
	// 获得反射后结果的指针(这里应该是一个切片的指针)
	pointValue := reflect.ValueOf(container)
	if pointValue.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("sql '%s' error : structs类型Query，查询结果类型应为struct切片的指针，而您传入的并不是指针", id)
	}

	// 获得反射后结果内容(这里应该是一个切片)
	value := reflect.Indirect(pointValue)
	if value.Kind() != reflect.Slice {
		return 0, fmt.Errorf("sql '%s' error : structs类型Query，查询结果类型应为struct切片的指针，而您传入的并不是切片", id)
	}

	// 切片元素类型(这里应该是struct的类型,也可以是struct的指针类型)
	rowType := value.Type().Elem()
	isStructPtr := rowType.Kind() == reflect.Ptr // 是否为struct的指针类型
	structType := rowType
	// 如果是struct的指针类型那么我们要获取struct类型
	if isStructPtr {
		structType = rowType.Elem()
	}
	// 无论如何structType都将成为struct的类型,如果不是,程序走不下去了
	if structType.Kind() != reflect.Struct {
		return 0, fmt.Errorf("sql '%s' error : structs类型Query，查询结果类型应为struct切片的指针，而您传入的并不是struct", id)
	}

	var rowsCount int64                      // 读取的行数，用于返回
	var columnsCount int                     // 读取的列数
	var fields []*structFieldInfo            // struct成员的名字，与sql中的列对应
	tagMap := map[string]*structFieldInfo{}  // struct每个成员的tag，优先匹配
	nameMap := map[string]*structFieldInfo{} // struct每个成员的名字，不一定与sql中的列对应
	getStructFieldMap(structType, tagMap, nameMap, false)

	// 使用提供的SQL，从数据库读取数据
	rows, err := o.db.Query(sql, sqlParams...)
	if err != nil {
		return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
	}
	defer rows.Close()

	// 遍历数据
	for rows.Next() {
		// 创建建struct实列,用来装这一行数据
		valueElem := reflect.New(structType).Elem()
		// 当isPtrs没有内容时,rowsCount,columnsCount,elementTypes,isPtrs,fieldNames的结果
		if fields == nil {
			columns, err := rows.Columns()
			if err != nil {
				return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
			}
			columnsCount = len(columns)
			// 定义
			fields = make([]*structFieldInfo, columnsCount)
			// 计算
			for i, col := range columns {
				fields[i] = findFiled(tagMap, nameMap, col)
			}
		}
		// 通过fieldName,创建struct实列的成员实例切片
		values := make([]reflect.Value, columnsCount)
		for i, field := range fields {
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
		// 读取一行数据到成员实例切片中
		err = o.scanRow(logPrefix, rows, fields, values)
		if err != nil {
			return 0, fmt.Errorf("sql '%s' error : %s", id, err.Error())
		}
		// struct实列装进结果切片
		if isStructPtr {
			value.Set(reflect.Append(value, valueElem.Addr()))
		} else {
			value.Set(reflect.Append(value, valueElem))
		}
		rowsCount++
	}
	return rowsCount, nil
}
