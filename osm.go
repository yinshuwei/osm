// 是一个go的ORM工具，目前很简单，只能算是半成品，只支持mysql(因为我目前的项目是mysql,所以其他数据库没有测试过)。
//
// 以前是使用MyBatis开发java服务端，它的sql mapping很灵活，把sql独立出来，程序通过输入与输出来完成所有的数据库操作。
//
// osm就是对MyBatis的简单模仿。当然动态sql的生成是使用go和template包，所以sql mapping的格式与MyBatis的不同。sql xml 格式如下：
//  <?xml version="1.0" encoding="utf-8"?>
//  <osm>
//   <select id="selectUsers" result="structs">
//     SELECT id,email
//     FROM user
//     {{if ne .Email ""}} where email=#{Email} {{end}}
//     order by id
//   </select>
//  </osm>
//
package osm

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/yinshuwei/utils"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

var logger *log.Logger = nil

type dbRunner interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type osmBase struct {
	db            dbRunner
	sqlMappersMap map[string]*sqlMapper
}

type Osm struct {
	osmBase
}

type OsmTx struct {
	osmBase
}

func NewOsm(driverName, dataSource string, xmlPaths []string, params ...int) (osm *Osm, err error) {
	if logger == nil {
		logger = log.New(utils.LogOutput, "[**osm**] ", utils.LogFlag)
	}

	osm = new(Osm)
	db, err := sql.Open(driverName, dataSource)

	if err != nil {
		if db != nil {
			db.Close()
		}
		err = fmt.Errorf("create osm %s", err.Error())
		return
	}

	osm.db = db
	osm.sqlMappersMap = make(map[string]*sqlMapper)

	for i, v := range params {
		switch i {
		case 0:
			db.SetMaxIdleConns(v)
		case 1:
			db.SetMaxOpenConns(v)
		}
	}

	osmXmlPaths := make([]string, 0)
	for _, xmlPath := range xmlPaths {
		var pathInfo os.FileInfo
		pathInfo, err = os.Stat(xmlPath)

		if err != nil {
			return
		}

		if pathInfo.IsDir() {
			if strings.LastIndex(xmlPath, "/") != (len([]rune(xmlPath)) - 1) {
				xmlPath += "/"
			}
			fs, _ := ioutil.ReadDir(xmlPath)
			for _, fileInfo := range fs {
				fileName := fileInfo.Name()
				if strings.LastIndex(fileName, ".xml") == (len([]rune(fileName)) - 4) {
					osmXmlPaths = append(osmXmlPaths, xmlPath+fileName)
				}
			}
		} else {
			osmXmlPaths = append(osmXmlPaths, xmlPath)
		}
	}

	for _, osmXmlPath := range osmXmlPaths {
		sqlMappers, err := readMappers(osmXmlPath)
		if err == nil {
			for _, sm := range sqlMappers {
				osm.sqlMappersMap[sm.id] = sm
			}
		} else {
			err = fmt.Errorf("read sqlMappers %s", err.Error())
		}
	}

	return
}

func (o *Osm) Begin() (tx *OsmTx, err error) {
	tx = new(OsmTx)
	tx.sqlMappersMap = o.sqlMappersMap

	if o.db == nil {
		err = fmt.Errorf("db no opened")
	} else {
		sqlDb, ok := o.db.(*sql.DB)
		if ok {
			tx.db, err = sqlDb.Begin()
		} else {
			err = fmt.Errorf("db no opened")
		}
	}

	return
}

func (tx *OsmTx) Commit() error {
	if tx.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := tx.db.(*sql.Tx)
	if ok {
		return sqlTx.Commit()
	} else {
		return fmt.Errorf("tx no runing")
	}
}

func (tx *OsmTx) Rollback() error {
	if tx.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := tx.db.(*sql.Tx)
	if ok {
		return sqlTx.Rollback()
	} else {
		return fmt.Errorf("tx no runing")
	}
}

func (o *osmBase) Delete(id string, params ...interface{}) (int64, error) {
	sql, sqlParams, _, err := o.readSqlParams(id, type_delete, params...)
	if err != nil {
		return 0, err
	}
	stmt, err := o.db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	result, err := stmt.Exec(sqlParams...)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	return result.RowsAffected()
}

func (o *osmBase) Update(id string, params ...interface{}) (int64, error) {
	sql, sqlParams, _, err := o.readSqlParams(id, type_update, params...)
	if err != nil {
		return 0, err
	}
	stmt, err := o.db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	result, err := stmt.Exec(sqlParams...)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	return result.RowsAffected()
}

func (o *osmBase) Insert(id string, params ...interface{}) (int64, int64, error) {
	sql, sqlParams, _, err := o.readSqlParams(id, type_insert, params...)
	if err != nil {
		return 0, 0, err
	}
	stmt, err := o.db.Prepare(sql)
	if err != nil {
		return 0, 0, err
	}

	result, err := stmt.Exec(sqlParams...)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()
	insertId, err := result.LastInsertId()
	if err != nil {
		return insertId, 0, err
	}
	count, err := result.RowsAffected()
	return insertId, count, err
}

func (o *osmBase) Query(id string, params ...interface{}) func(containers ...interface{}) (int64, error) {
	sql, sqlParams, resultType, err := o.readSqlParams(id, type_select, params...)

	if err != nil {
		return func(containers ...interface{}) (int64, error) {
			return 0, err
		}
	}
	callback := func(containers ...interface{}) (int64, error) {
		var err error = nil
		switch resultType {
		case result_structs:
			if len(containers) == 1 {
				return resultStructs(o, sql, sqlParams, containers[0])
			} else {
				err = fmt.Errorf("result_structs ,len(containers) != 1")
			}
		case result_struct:
			if len(containers) == 1 {
				return resultStruct(o, sql, sqlParams, containers[0])
			} else {
				err = fmt.Errorf("result_struct ,len(containers) != 1")
			}
		case result_maps:
			if len(containers) == 1 {
				return resultMaps(o, sql, sqlParams, containers[0])
			} else {
				err = fmt.Errorf("result_maps ,len(containers) != 1")
			}
		case result_map:
			if len(containers) == 1 {
				return resultMap(o, sql, sqlParams, containers[0])
			} else {
				err = fmt.Errorf("result_map ,len(containers) != 1")
			}
		case result_arrays:
			if len(containers) == 1 {
				return resultArrays(o, sql, sqlParams, containers[0])
			} else {
				err = fmt.Errorf("result_arrays ,len(containers) != 1")
			}
		case result_array:
			if len(containers) == 1 {
				return resultArray(o, sql, sqlParams, containers[0])
			} else {
				err = fmt.Errorf("result_array ,len(containers) != 1")
			}
		case result_value:
			if len(containers) > 0 {
				return resultValue(o, sql, sqlParams, containers)
			} else {
				err = fmt.Errorf("result_value ,len(containers) < 1")
			}
		case result_kvs:
			if len(containers) == 1 {
				return resultKvs(o, sql, sqlParams, containers[0])
			} else {
				err = fmt.Errorf("result_kvs ,len(containers) != 1")
			}
		}

		if err == nil {
			err = fmt.Errorf("sql result_type no in ['value','struct','','','','']")
		}
		return 0, err
	}
	return callback
}

func (o *osmBase) readSqlParams(id string, sqlType int, params ...interface{}) (sql string, sqlParams []interface{}, resultType string, err error) {
	sqlParams = make([]interface{}, 0)
	sm, ok := o.sqlMappersMap[id]
	err = nil

	if !ok {
		err = fmt.Errorf("Query \"%s\" error ,id not fond ", id)
		return
	}
	resultType = sm.result

	if sm.sqlType != sqlType {
		err = fmt.Errorf("Query type Error")
		return
	}

	var param interface{}
	paramsSize := len(params)
	if paramsSize > 0 {
		if paramsSize == 1 {
			param = params[0]
		} else {
			param = params
		}

		//sql start
		sqls := make([]string, 0)
		paramNames := make([]string, 0)
		startFlag := 0

		var buf bytes.Buffer

		sm.sqlTemplate.Execute(&buf, param)
		sqlOrg := buf.String()
		logger.Println("\n\tsql:\n", sqlOrg, "\n\tparams:\n", param, "\n")

		sqlTemp := sqlOrg

		errorIndex := 0

		for strings.Contains(sqlTemp, "#{") || strings.Contains(sqlTemp, "}") {
			si := strings.Index(sqlTemp, "#{")
			ei := strings.Index(sqlTemp, "}")
			if si != -1 && si < ei {
				sqls = append(sqls, sqlTemp[0:si])
				sqlTemp = sqlTemp[si+2:]
				startFlag++
				errorIndex += si + 2
			} else if (ei != -1 && si != -1 && ei < si) || (ei != -1 && si == -1) {
				sqls = append(sqls, "?")
				paramNames = append(paramNames, sqlTemp[0:ei])
				sqlTemp = sqlTemp[ei+1:]
				startFlag--
				errorIndex += ei + 1
			} else {
				if ei > -1 {
					errorIndex += ei
				} else {
					errorIndex += si
				}
				logger.Printf("sql read error \"%v\"\n", markSqlError(sqlOrg, errorIndex))
				return
			}

		}
		sqls = append(sqls, sqlTemp)
		//sql end

		if startFlag != 0 {
			logger.Printf("sql read error \"%v\"\n", markSqlError(sqlOrg, errorIndex))
			return
		}
		sql = strings.Join(sqls, "")

		v := reflect.ValueOf(param)

		kind := v.Kind()
		switch {
		case kind == reflect.Array || kind == reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				vv := v.Index(i)
				sqlParams = append(sqlParams, vv.Interface())
			}
		case kind == reflect.Map:
			for _, paramName := range paramNames {
				if ok {
					vv := v.MapIndex(reflect.ValueOf(paramName))
					sqlParams = append(sqlParams, vv.Interface())
				} else {
					err = fmt.Errorf("array type not map[string]interface{} of %s", param)
					return
				}
			}
		case kind == reflect.Struct:
			for _, paramName := range paramNames {
				vv := v.FieldByName(paramName)
				if vv.IsValid() {
					if vv.Type().String() == "time.Time" {
						sqlParams = append(sqlParams, timeFormat(vv.Interface().(time.Time), format_DateTime))
					} else {
						sqlParams = append(sqlParams, vv.Interface())
					}
				} else {
					sqlParams = append(sqlParams, nil)
					err = fmt.Errorf("sql '%s' error : '%s' no exist", sm.id, paramName)
					return
				}
			}
		case kind == reflect.Bool ||
			kind == reflect.Int ||
			kind == reflect.Int8 ||
			kind == reflect.Int16 ||
			kind == reflect.Int32 ||
			kind == reflect.Int64 ||
			kind == reflect.Uint ||
			kind == reflect.Uint8 ||
			kind == reflect.Uint16 ||
			kind == reflect.Uint32 ||
			kind == reflect.Uint64 ||
			kind == reflect.Uintptr ||
			kind == reflect.Float32 ||
			kind == reflect.Float64 ||
			kind == reflect.Complex64 ||
			kind == reflect.Complex128 ||
			kind == reflect.String:
			sqlParams = append(sqlParams, param)
		default:
		}
	} else {
		sql = sm.sql
		logger.Println(sql)
	}
	return
}
