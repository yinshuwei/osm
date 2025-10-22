package osm

import (
	"path"
	"runtime"
	"strconv"
	"time"
)

// SelectResult 查询结果对象，支持链式调用
type SelectResult struct {
	osmBase   *osmBase
	sql       string
	sqlParams []interface{}
	logPrefix string
	err       error
}

// Select 执行查询sql，返回SelectResult对象用于链式调用
//
// 用法示例:
//
//	var users []User
//	_, err = o.Select(`SELECT * FROM users WHERE id > #{Id}`, 1).Structs(&users)
//
//	var email string
//	_, err = o.Select(`SELECT email FROM users WHERE id = #{Id}`, 1).String()
func (o *osmBase) Select(sql string, params ...interface{}) *SelectResult {
	logPrefix := ""
	_, file, lineNo, ok := runtime.Caller(1)
	if ok {
		fileName := path.Base(file)
		logPrefix = fileName + ":" + strconv.Itoa(lineNo) + ", "
	}

	now := time.Now()
	go func(start time.Time) {
		if time.Since(start) > o.options.SlowLogDuration {
			o.options.WarnLogger.Log(logPrefix+"slow sql", map[string]string{"sql": sql, "cost": time.Since(start).String()})
		}
	}(now)

	result := &SelectResult{
		osmBase:   o,
		logPrefix: logPrefix,
	}

	result.sql, result.sqlParams, result.err = o.readSQLParamsBySQL(logPrefix, sql, params...)
	return result
}

// Struct 查询单行数据并存入struct
//
// 用法:
//
//	var user User
//	_, err = o.Select(`SELECT * FROM users WHERE id = #{Id}`, 1).Struct(&user)
func (sr *SelectResult) Struct(container interface{}) (int64, error) {
	if sr.err != nil {
		return 0, sr.err
	}
	return resultStruct(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, container)
}

// Structs 查询多行数据并存入struct切片
//
// 用法:
//
//	var users []User
//	_, err = o.Select(`SELECT * FROM users`).Structs(&users)
func (sr *SelectResult) Structs(container interface{}) (int64, error) {
	if sr.err != nil {
		return 0, sr.err
	}
	return resultStructs(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, container)
}

// Kvs 查询多行两列数据并存入map
//
// 用法:
//
//	var idEmailMap = map[int64]string{}
//	_, err = o.Select(`SELECT id, email FROM users`).Kvs(&idEmailMap)
func (sr *SelectResult) Kvs(container interface{}) (int64, error) {
	if sr.err != nil {
		return 0, sr.err
	}
	return resultKvs(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, container)
}

// ColumnsAndData 查询多行数据，返回列名和数据
//
// 用法:
//
//	var columns []string
//	var datas [][]string
//	_, err := o.Select(`SELECT id, email FROM users`).ColumnsAndData(&columns, &datas)
func (sr *SelectResult) ColumnsAndData() ([]string, [][]string, error) {
	if sr.err != nil {
		return nil, nil, sr.err
	}
	var columns []string
	var datas [][]string
	_, err := resultStrings(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, &columns, &datas)
	if err != nil {
		return nil, nil, err
	}
	return columns, datas, nil
}

// String 查询单个字符串值
//
// 用法:
//
//	email, err := o.Select(`SELECT email FROM users WHERE id = #{Id}`, 1).String()
func (sr *SelectResult) String() (string, error) {
	results, err := sr.Strings()
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", nil
	}
	return results[0], nil
}

// Strings 查询多个字符串值
//
// 用法:
//
//	emails, err := o.Select(`SELECT email FROM users`).Strings()
func (sr *SelectResult) Strings() ([]string, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []string
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}

// Int 查询单个int值
//
// 用法:
//
//	count, err := o.Select(`SELECT COUNT(*) FROM users`).Int()
func (sr *SelectResult) Int() (int, error) {
	results, err := sr.Ints()
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0], nil
}

// Ints 查询多个int值
//
// 用法:
//
//	ids, err := o.Select(`SELECT age FROM users`).Ints()
func (sr *SelectResult) Ints() ([]int, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []int
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}

// Int64 查询单个int64值
//
// 用法:
//
//	id, err := o.Select(`SELECT id FROM users WHERE email = #{Email}`, "test@example.com").Int64()
func (sr *SelectResult) Int64() (int64, error) {
	results, err := sr.Int64s()
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0], nil
}

// Int64s 查询多个int64值
//
// 用法:
//
//	ids, err := o.Select(`SELECT id FROM users`).Int64s()
func (sr *SelectResult) Int64s() ([]int64, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []int64
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}

// Float64 查询单个float64值
//
// 用法:
//
//	avg, err := o.Select(`SELECT AVG(score) FROM users`).Float64()
func (sr *SelectResult) Float64() (float64, error) {
	results, err := sr.Float64s()
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0], nil
}

// Float64s 查询多个float64值
//
// 用法:
//
//	scores, err := o.Select(`SELECT score FROM users`).Float64s()
func (sr *SelectResult) Float64s() ([]float64, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []float64
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}

// Int32 查询单个int32值
//
// 用法:
//
//	count, err := o.Select(`SELECT count FROM table WHERE id = #{Id}`, 1).Int32()
func (sr *SelectResult) Int32() (int32, error) {
	results, err := sr.Int32s()
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0], nil
}

// Int32s 查询多个int32值
//
// 用法:
//
//	counts, err := o.Select(`SELECT count FROM table`).Int32s()
func (sr *SelectResult) Int32s() ([]int32, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []int32
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}

// Float32 查询单个float32值
//
// 用法:
//
//	price, err := o.Select(`SELECT price FROM products WHERE id = #{Id}`, 1).Float32()
func (sr *SelectResult) Float32() (float32, error) {
	results, err := sr.Float32s()
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0], nil
}

// Float32s 查询多个float32值
//
// 用法:
//
//	prices, err := o.Select(`SELECT price FROM products`).Float32s()
func (sr *SelectResult) Float32s() ([]float32, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []float32
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}

// Uint 查询单个uint值
//
// 用法:
//
//	count, err := o.Select(`SELECT COUNT(*) FROM users`).Uint()
func (sr *SelectResult) Uint() (uint, error) {
	results, err := sr.Uints()
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0], nil
}

// Uints 查询多个uint值
//
// 用法:
//
//	counts, err := o.Select(`SELECT count FROM table`).Uints()
func (sr *SelectResult) Uints() ([]uint, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []uint
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}

// Uint64 查询单个uint64值
//
// 用法:
//
//	id, err := o.Select(`SELECT id FROM users WHERE email = #{Email}`, "test@example.com").Uint64()
func (sr *SelectResult) Uint64() (uint64, error) {
	results, err := sr.Uint64s()
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0], nil
}

// Uint64s 查询多个uint64值
//
// 用法:
//
//	ids, err := o.Select(`SELECT id FROM users`).Uint64s()
func (sr *SelectResult) Uint64s() ([]uint64, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []uint64
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}

// Bool 查询单个布尔值
//
// 用法:
//
//	isActive, err := o.Select(`SELECT is_active FROM users WHERE id = #{Id}`, 1).Bool()
func (sr *SelectResult) Bool() (bool, error) {
	results, err := sr.Bools()
	if err != nil {
		return false, err
	}
	if len(results) == 0 {
		return false, nil
	}
	return results[0], nil
}

// Bools 查询多个布尔值
//
// 用法:
//
//	statuses, err := o.Select(`SELECT is_active FROM users`).Bools()
func (sr *SelectResult) Bools() ([]bool, error) {
	if sr.err != nil {
		return nil, sr.err
	}
	var result []bool
	_, err := resultValues(sr.logPrefix, sr.osmBase, sr.sql, sr.sql, sr.sqlParams, []interface{}{&result})
	return result, err
}
