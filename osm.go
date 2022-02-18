package osm

// osm (Object Sql Mapping)是用go编写的ORM工具，目前很简单，只能算是半成品，只支持mysql(因为我目前的项目是mysql,所以其他数据库没有测试过)。
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

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	dbTypeMysql    = 0
	dbTypePostgres = 1
	dbTypeMssql    = 2
)

var (
	errorLogger Logger = log.Default()
	infoLogger  Logger = log.Default()

	// ShowSQL 显示执行的sql，用于调试，使用logger打印
	showSQL = false
)

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

// Osm 对象，通过Struct、Map、Array、value等对象以及Sql Map来操作数据库。可以开启事务。
type Osm struct {
	osmBase
}

// Tx 与Osm对象一样，不过是在事务中进行操作
type Tx struct {
	osmBase
}

func ConfLogger(_infoLogger, _errorLogger Logger, _showSQL bool) {
	infoLogger = _infoLogger
	errorLogger = _errorLogger
	showSQL = _showSQL
}

// New 创建一个新的Osm，这个过程会打开数据库连接。
//
//driverName是数据库驱动名称如"mysql".
//dataSource是数据库连接信息如"root:root@/51jczj?charset=utf8".
//xmlPaths是sql xml的路径如[]string{"test.xml"}.
//params是数据连接的参数，可以是0个1个或2个数字，第一个表示MaxIdleConns，第二个表示MaxOpenConns.
//
//如：
//  o, err := osm.New("mysql", "root:root@/51jczj?charset=utf8", []string{"test.xml"})
func New(driverName, dataSource string, xmlPaths []string, params ...int) (*Osm, error) {
	osm := new(Osm)
	db, err := sql.Open(driverName, dataSource)

	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, fmt.Errorf("create osm error : %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create osm error : %s", err.Error())
	}

	go func() {
		for {
			err := db.Ping()
			if err != nil {
				errorLogger.Printf("osm Ping fail: %s", err.Error())
			}
			time.Sleep(time.Minute)
		}
	}()

	switch driverName {
	case "postgres":
		osm.dbType = dbTypePostgres
	case "mssql":
		osm.dbType = dbTypeMssql
	default:
		osm.dbType = dbTypeMysql
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

	osmXMLPaths := []string{}
	for _, xmlPath := range xmlPaths {
		var pathInfo os.FileInfo
		pathInfo, err = os.Stat(xmlPath)

		if err != nil {
			return nil, err
		}

		if pathInfo.IsDir() {
			if strings.LastIndex(xmlPath, "/") != (len([]rune(xmlPath)) - 1) {
				xmlPath += "/"
			}
			fs, _ := ioutil.ReadDir(xmlPath)
			for _, fileInfo := range fs {
				fileName := fileInfo.Name()
				if strings.LastIndex(fileName, ".xml") == (len([]rune(fileName)) - 4) {
					osmXMLPaths = append(osmXMLPaths, xmlPath+fileName)
				}
			}
		} else {
			osmXMLPaths = append(osmXMLPaths, xmlPath)
		}
	}

	for _, osmXMLPath := range osmXMLPaths {
		sqlMappers, err := readMappers(osmXMLPath)
		if err != nil {
			return nil, fmt.Errorf("read sqlMappers error : %s", err.Error())
		}
		for _, sm := range sqlMappers {
			osm.sqlMappersMap[sm.id] = sm
		}
	}

	return osm, nil
}

// Begin 打开事务
//
//如：
//  tx, err := o.Begin()
func (o *Osm) Begin() (*Tx, error) {
	tx := new(Tx)
	tx.sqlMappersMap = o.sqlMappersMap
	tx.dbType = o.dbType

	if o.db == nil {
		return nil, fmt.Errorf("db no opened")
	}
	sqlDb, ok := o.db.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("db no opened")
	}

	var err error
	tx.db, err = sqlDb.Begin()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Close 与数据库断开连接，释放连接资源
//
//如：
//  err := o.Close()
func (o *Osm) Close() error {
	if o.db == nil {
		return fmt.Errorf("db no opened")
	}
	sqlDb, ok := o.db.(*sql.DB)
	if !ok {
		return fmt.Errorf("db no opened")
	}

	o.db = nil
	return sqlDb.Close()
}

// Commit 提交事务
//
//如：
//  err := tx.Commit()
func (o *Tx) Commit() error {
	if o.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := o.db.(*sql.Tx)
	if !ok {
		return fmt.Errorf("tx no runing")
	}
	return sqlTx.Commit()
}

// Rollback 事务回滚
//
//如：
//  err := tx.Rollback()
func (o *Tx) Rollback() error {
	if o.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := o.db.(*sql.Tx)
	if !ok {
		return fmt.Errorf("tx no runing")
	}
	return sqlTx.Rollback()
}

// Delete 通过id在xml中找到删除sql并执行
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
	sql, sqlParams, _, err := o.readSQLParams(id, typeDelete, params...)
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

// Update 通过id在xml中找到更新sql并执行
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
	sql, sqlParams, _, err := o.readSQLParams(id, typeUpdate, params...)
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

// UpdateMulti 批量通过id在xml中找到更新sql并执行
//
//xml
//  <osm>
//  ...
//    <update id="updateUserEmail">
//       UPDATE user SET email=#{Email} where id = #{Id};
//       UPDATE user SET email=#{Email} where id = #{Id2};
//    </update>
//  ...
//  </osm>
//代码
//  user := User{Id: 3, Id2: 4, Email: "test@foxmail.com"}
//  err := o.UpdateMulti("updateUserEmail", user)
//将id为3和4的用户email更新为"test@foxmail.com"
func (o *osmBase) UpdateMulti(id string, params ...interface{}) error {
	sql, sqlParams, _, err := o.readSQLParams(id, typeUpdate, params...)
	if err != nil {
		return err
	}
	_, err = o.db.Exec(sql, sqlParams...)
	return err
}

// Insert 通过id在xml中找到添加sql并执行
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
	sql, sqlParams, _, err := o.readSQLParams(id, typeInsert, params...)
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

	var insertID int64
	if o.dbType == dbTypeMysql {
		insertID, err = result.LastInsertId()
		if err != nil {
			errorLogger.Printf("lastInsertId read error: %s", err.Error())
		}
	}

	count, err := result.RowsAffected()
	return insertID, count, err
}

//通过id在xml中找到查询sql并执行
//
//查询结果分为8种，分别是:
//	"value"   : 查出的结果为单行,并存入不定长的变量上(...)，可以是指针，如var r1,r2 string、var r1,r2 *string
//	"values"  : 查出的结果为多行,并存入不定长的变量上(...，每个都为array)，元素可以是指针，如var r1,r2 []string、var r1,r2 []*string都允许
//	"struct"  : 查出的结果为单行,并存入struct，可以是指针，如var r User、var r *User
//	"structs" : 查出的结果为多行,并存入struct array，元素可以是指针，如var r []User、var r []*User
//	"kvs"     : 查出的结果为多行,每行有两个字段,前者为key,后者为value,存入map (双列)，Key、Value可以是指针，如var r map[string]time.Time、var r map[*string]time.Time、var r map[string]*time.Time
//	"strings" : 查出的结果为多行,并存入columns，和datas。columns为[]string，datas为[][]string
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
	sql, sqlParams, resultType, err := o.readSQLParams(id, typeSelect, params...)

	if err != nil {
		return func(containers ...interface{}) (int64, error) {
			return 0, err
		}
	}
	callback := func(containers ...interface{}) (int64, error) {
		switch resultType {
		case resultTypeStructs:
			if len(containers) != 1 {
				return 0, fmt.Errorf("sql '%s' error : resultTypeStructs ,len(containers) != 1", id)
			}
			return resultStructs(o, id, sql, sqlParams, containers[0])
		case resultTypeStruct:
			if len(containers) != 1 {
				return 0, fmt.Errorf("sql '%s' error : resultTypeStruct ,len(containers) != 1", id)
			}
			return resultStruct(o, id, sql, sqlParams, containers[0])
		case resultTypeValue:
			if len(containers) == 0 {
				return 0, fmt.Errorf("sql '%s' error : resultTypeValue ,len(containers) == 0", id)
			}
			return resultValue(o, id, sql, sqlParams, containers)
		case resultTypeValues:
			if len(containers) == 0 {
				return 0, fmt.Errorf("sql '%s' error : resultTypeValues ,len(containers) == 0", id)
			}
			return resultValues(o, id, sql, sqlParams, containers)
		case resultTypeKvs:
			if len(containers) != 1 {
				return 0, fmt.Errorf("sql '%s' error : resultTypeKvs ,len(containers) != 1", id)
			}
			return resultKvs(o, id, sql, sqlParams, containers[0])
		case resultTypeStrings:
			if len(containers) != 2 {
				return 0, fmt.Errorf("sql '%s' error : resultTypeKvs ,len(containers) != 2", id)
			}
			return resultStrings(o, id, sql, sqlParams, containers[0], containers[1])
		}
		return 0, fmt.Errorf("sql '%s' error : sql resultTypeType no in ['value','struct','values','structs','kvs']", id)

	}
	return callback
}

type sqlFragment struct {
	content     string
	paramValue  interface{}
	paramValues []interface{}
	isParam     bool
	isIn        bool
}

func setDataToParamName(paramName *sqlFragment, v reflect.Value) {
	if paramName.isIn {
		v = reflect.ValueOf(v.Interface())
		kind := v.Kind()
		if kind == reflect.Array || kind == reflect.Slice {
			for j := 0; j < v.Len(); j++ {
				vv := v.Index(j)
				if vv.Type().String() == "time.Time" {
					paramName.paramValues = append(paramName.paramValues, timeFormat(vv.Interface().(time.Time), formatDateTime))
				} else {
					paramName.paramValues = append(paramName.paramValues, vv.Interface())
				}
			}
		} else {
			if v.Type().String() == "time.Time" {
				paramName.paramValues = append(paramName.paramValues, timeFormat(v.Interface().(time.Time), formatDateTime))
			} else {
				paramName.paramValues = append(paramName.paramValues, v.Interface())
			}
		}
	} else {
		if v.Type().String() == "time.Time" {
			paramName.paramValue = timeFormat(v.Interface().(time.Time), formatDateTime)
		} else {
			paramName.paramValue = v.Interface()
		}
	}
}

func sqlIsIn(lastSQLText string) bool {
	lastSQLText = strings.TrimSpace(lastSQLText)
	lenLastSQLText := len(lastSQLText)
	if lenLastSQLText > 3 {
		return strings.EqualFold(lastSQLText[lenLastSQLText-3:], " IN")
	}
	return false
}

func (o *osmBase) readSQLParams(id string, sqlType int, params ...interface{}) (string, []interface{}, string, error) {
	sm, ok := o.sqlMappersMap[id]
	if !ok || sm == nil {
		return "", nil, "", fmt.Errorf("sql '%s' error : id not found ", id)
	}

	if sm.sqlType != sqlType {
		return "", nil, "", fmt.Errorf("sql '%s' error : Select type Error", id)
	}

	sql := ""
	sqlParams := []interface{}{}
	resultType := sm.result

	var param interface{}
	paramsSize := len(params)
	if paramsSize > 0 {
		if paramsSize == 1 {
			param = params[0]
		} else {
			param = params
		}

		//sql start
		sqls := []*sqlFragment{}
		paramNames := []*sqlFragment{}
		var buf bytes.Buffer

		err := sm.sqlTemplate.Execute(&buf, param)
		if err != nil {
			return "", nil, "", err
		}
		sqlOrg := buf.String()
		defer func() {
			if showSQL {
				params, _ := json.Marshal(param)
				sqlParams, _ := json.Marshal(sqlParams)
				infoLogger.Printf("readSQLParams showSql, id: %s , sql: %s, params: %s, dbSql: %s, dbParams: %s", id, sqlOrg, string(params), sql, string(sqlParams))
			}
		}()
		sqlTemp := sqlOrg
		errorIndex := 0
		for strings.Contains(sqlTemp, "#{") {
			si := strings.Index(sqlTemp, "#{")
			lastSQLText := sqlTemp[0:si]
			sqls = append(sqls, &sqlFragment{
				content: lastSQLText,
			})
			sqlTemp = sqlTemp[si+2:]
			errorIndex += si + 2

			ei := strings.Index(sqlTemp, "}")
			if ei != -1 {
				pni := &sqlFragment{
					content: strings.TrimSpace(sqlTemp[0:ei]),
					isParam: true,
					isIn:    sqlIsIn(lastSQLText),
				}
				sqls = append(sqls, pni)
				paramNames = append(paramNames, pni)
				sqlTemp = sqlTemp[ei+1:]
				errorIndex += ei + 1
			} else {
				return "", nil, "", fmt.Errorf("sql read error \"%v\"", markSQLError(sqlOrg, errorIndex))
			}
		}
		sqls = append(sqls, &sqlFragment{
			content: sqlTemp,
		})
		//sql end

		v := reflect.ValueOf(param)

		kind := v.Kind()
		switch {
		case kind == reflect.Array || kind == reflect.Slice:
			if len(paramNames) == 1 && paramNames[0].isIn {
				setDataToParamName(paramNames[0], v)
			} else {
				for i := 0; i < v.Len() && i < len(paramNames); i++ {
					vv := v.Index(i)
					if vv.IsValid() {
						setDataToParamName(paramNames[i], v.Index(i))
					}
				}
			}
		case kind == reflect.Map:
			for _, paramName := range paramNames {
				vv := v.MapIndex(reflect.ValueOf(paramName.content))
				if vv.IsValid() {
					setDataToParamName(paramName, vv)
				} else {
					return "", nil, "", fmt.Errorf("sql '%s' error : Key '%s' no exist", sm.id, paramName.content)
				}
			}
		case kind == reflect.Struct:
			for _, paramName := range paramNames {
				firstChar := paramName.content[0]
				if firstChar < 'A' || firstChar > 'Z' {
					return "", nil, "", fmt.Errorf("sql '%s' error : Field '%s' unexported", sm.id, paramName.content)
				}
				vv := v.FieldByName(paramName.content)
				if vv.IsValid() {
					setDataToParamName(paramName, vv)
				} else {
					return "", nil, "", fmt.Errorf("sql '%s' error : Field '%s' no exist", sm.id, paramName.content)
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
			for _, paramName := range paramNames {
				setDataToParamName(paramName, v)
			}
		default:
		}

		var sqlTexts []string
		signIndex := 1
		for _, sql := range sqls {
			if sql.isParam {
				if sql.isIn {
					sqlTexts = append(sqlTexts, "(")
					for index, pv := range sql.paramValues {
						if index > 0 {
							sqlTexts = append(sqlTexts, ",")
						}
						if o.dbType == dbTypeMysql {
							sqlTexts = append(sqlTexts, "?")
						} else {
							sqlTexts = append(sqlTexts, "$"+strconv.Itoa(signIndex))
							signIndex++
						}
						sqlParams = append(sqlParams, pv)
					}
					sqlTexts = append(sqlTexts, ")")
				} else {
					if o.dbType == dbTypeMysql {
						sqlTexts = append(sqlTexts, "?")
					} else {
						sqlTexts = append(sqlTexts, "$"+strconv.Itoa(signIndex))
						signIndex++
					}
					sqlParams = append(sqlParams, sql.paramValue)
				}
			} else {
				sqlTexts = append(sqlTexts, sql.content)
			}
		}

		sql = strings.Join(sqlTexts, "")
	} else {
		sql = sm.sql
		if showSQL {
			infoLogger.Printf("readSQLParams showSql, id: %s, sql: %s", id, sql)
		}
	}
	return sql, sqlParams, resultType, nil
}
