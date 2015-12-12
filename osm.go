// osm(Object Sql Mapping)是用go编写的ORM工具，目前很简单，只能算是半成品，只支持mysql(因为我目前的项目是mysql,所以其他数据库没有测试过)。
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

const (
	DBTYPE_MYSQL    = 0
	DBTYPE_POSTGRES = 1
)

var logger *log.Logger = nil

//显示执行的sql，用于调试，使用logger打印
var ShowSql = false

type dbRunner interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type osmBase struct {
	db            dbRunner
	sqlMappersMap map[string]*sqlMapper
	dbType        int
}

//osm对象，通过Struct、Map、Array、value等对象以及Sql Map来操作数据库。可以开启事务。
type Osm struct {
	osmBase
}

//与Osm对象一样，不过是在事务中进行操作
type OsmTx struct {
	osmBase
}

//创建一个新的Osm，这个过程会打开数据库连接。
//
//driverName是数据库驱动名称如"mysql".
//dataSource是数据库连接信息如"root:root@/51jczj?charset=utf8".
//xmlPaths是sql xml的路径如[]string{"test.xml"}.
//params是数据连接的参数，可以是0个1个或2个数字，第一个表示MaxIdleConns，第二个表示MaxOpenConns.
//
//如：
//  o, err := osm.New("mysql", "root:root@/51jczj?charset=utf8", []string{"test.xml"})
func New(driverName, dataSource string, xmlPaths []string, params ...int) (osm *Osm, err error) {
	if logger == nil {
		logger = log.New(utils.LogOutput, "[osm] ", utils.LogFlag)
	}

	osm = new(Osm)
	db, err := sql.Open(driverName, dataSource)

	if err != nil {
		if db != nil {
			db.Close()
		}
		err = fmt.Errorf("create osm error : %s", err.Error())
		return
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		err = fmt.Errorf("create osm error : %s", err.Error())
		return
	}

	switch driverName {
	case "postgres":
		osm.dbType = DBTYPE_POSTGRES
	default:
		osm.dbType = DBTYPE_MYSQL
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

//打开事务
//
//如：
//  tx, err := o.Begin()
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

//与数据库断开连接，释放连接资源
//
//如：
//  err := o.Close()
func (o *Osm) Close() (err error) {
	if o.db == nil {
		err = fmt.Errorf("db no opened")
	} else {
		sqlDb, ok := o.db.(*sql.DB)
		if ok {
			err = sqlDb.Close()
			o.db = nil
		} else {
			err = fmt.Errorf("db no opened")
		}
	}
	return
}

//提交事务
//
//如：
//  err := tx.Commit()
func (o *OsmTx) Commit() error {
	if o.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := o.db.(*sql.Tx)
	if ok {
		return sqlTx.Commit()
	} else {
		return fmt.Errorf("tx no runing")
	}
}

//事务回滚
//
//如：
//  err := tx.Rollback()
func (o *OsmTx) Rollback() error {
	if o.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := o.db.(*sql.Tx)
	if ok {
		return sqlTx.Rollback()
	} else {
		return fmt.Errorf("tx no runing")
	}
}

//执行删除sql
//
//xml
//  <osm>
//  ...
//    <delete id="deleteUser">DELETE FROM user where id = #{Id};</delete>
//  ...
//  </osm>
//代码
//  user := User{Id: 3}
//  count,err := o.Delete("deleteUser", user)
//删除id为3的用户数据
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

//执行更新sql
//
//xml
//  <osm>
//  ...
//    <update id="updateUserEmail">UPDATE user SET email=#{Email} where id = #{Id};</update>
//  ...
//  </osm>
//代码
//  user := User{Id: 3, Email: "test@foxmail.com"}
//  count,err := o.Update("updateUserEmail", user)
//将id为3的用户email更新为"test@foxmail.com"
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

//执行添加sql
//
//xml
//  <osm>
//  ...
//    <insert id="insertUser">INSERT INTO user(email) VALUES(#{Email});</insert>
//  ...
//  </osm>
//代码
//  user := User{Email: "test@foxmail.com"}
//  insertId,count,err := o.Insert("insertUser", user)
//添加一个用户数据，email为"test@foxmail.com"
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

//执行查询sql
//
//查询结果分为8种，分别是:
//	"value"   : 查出的结果为单行,并存入不定长的变量上(...)
//	"struct"  : 查出的结果为单行,并存入struct
//	"structs" : 查出的结果为多行,并存入struct array
//	"map"     : 查出的结果为单行,并存入map
//	"maps"    : 查出的结果为多行,并存入map array
//	"array"   : 查出的结果为单行,并存入array
//	"arrays"  : 查出的结果为多行,并存入array array
//	"kvs"     : 查出的结果为多行,每行有两个字段,前者为key,后者为value,存入map
//xml
//  <select id="searchArchives" result="struct">
//   <![CDATA[
//   SELECT id,email,create_time FROM user WHERE id=#{Id};
//   ]]>
//  </select>
//result上面8种的一种,查询结果会将列名转换为属性名，如"create_time"列,在结果中存放在CreateTime属性中
//
//上面的结果为User{Id: "1", Email: "test@foxmail.com", CreateTime: "2014-06-01 12:32:40"}
func (o *osmBase) Select(id string, params ...interface{}) func(containers ...interface{}) (int64, error) {
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
		err = fmt.Errorf("Select \"%s\" error ,id not fond ", id)
		return
	}
	resultType = sm.result

	if sm.sqlType != sqlType {
		err = fmt.Errorf("Select type Error")
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

		if ShowSql {
			logger.Printf(`sql:"%s", params:"%+v"`, sqlOrg, param)
		}

		sqlTemp := sqlOrg
		errorIndex := 0
		signIndex := 1
		for strings.Contains(sqlTemp, "#{") || strings.Contains(sqlTemp, "}") {
			si := strings.Index(sqlTemp, "#{")
			ei := strings.Index(sqlTemp, "}")
			if si != -1 && si < ei {
				sqls = append(sqls, sqlTemp[0:si])
				sqlTemp = sqlTemp[si+2:]
				startFlag++
				errorIndex += si + 2
			} else if (ei != -1 && si != -1 && ei < si) || (ei != -1 && si == -1) {
				if o.dbType == DBTYPE_POSTGRES {
					sqls = append(sqls, fmt.Sprintf("$%d", signIndex))
					signIndex++
				} else {
					sqls = append(sqls, "?")
				}
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
				logger.Printf("sql read error \"%v\"", markSqlError(sqlOrg, errorIndex))
				return
			}

		}
		sqls = append(sqls, sqlTemp)
		//sql end

		if startFlag != 0 {
			logger.Printf("sql read error \"%v\"", markSqlError(sqlOrg, errorIndex))
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
		if ShowSql {
			logger.Printf(`sql:"%s"`, sqlOrg)
		}
	}
	return
}
