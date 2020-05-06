package osm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// DeleteBySQL 执行删除sql
//
//代码
//   count, err := o.DeleteBySQL(`DELETE FROM res_user WHERE id in #{Ids};`, []int64{1, 2})
//   if err != nil {
// 	   log.Println(err)
//   }
//   log.Println("count:", count)
//结果
//
//   count: 2
//删除id为1和2的用户数据
func (o *osmBase) DeleteBySQL(sql string, params ...interface{}) (int64, error) {
	sql, sqlParams, err := o.readSQLParamsBySQL(sql, params...)
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

// UpdateBySQL 执行更新sql
//
//代码
//   count, err := o.UpdateBySQL(`UPDATE res_user SET email=#{Email} WHERE id=#{ID};`, "test2@foxmail.com", 3)
//   if err != nil {
// 	  log.Println(err)
//   }
//   log.Println("count:", count)
//结果
//   count: 1
//
//将id为1的用户email更新为"test2@foxmail.com"
func (o *osmBase) UpdateBySQL(sql string, params ...interface{}) (int64, error) {
	sql, sqlParams, err := o.readSQLParamsBySQL(sql, params...)
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

// InsertBySQL 执行添加sql
//
//代码
//   insertResUser := ResUser{
// 	  Email: "test@foxmail.com",
//   }
//   insertID, count, err := o.InsertBySQL("INSERT INTO res_user (email) VALUES(#{Email});", insertResUser)
//   if err != nil {
// 	  log.Println(err)
//   }
//   log.Println("insertID:", insertID, "count:", count)
//结果
//   insertID: 3 count: 1
//
//添加一个用户数据，email为"test@foxmail.com"
func (o *osmBase) InsertBySQL(sql string, params ...interface{}) (int64, int64, error) {
	sql, sqlParams, err := o.readSQLParamsBySQL(sql, params...)
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
			logger.Println(err)
		}
	}

	count, err := result.RowsAffected()
	return insertID, count, err
}

// SelectValue 执行查询sql
//
//查出的结果为单行,并存入不定长的变量上(...)，可以是指针，如var r1,r2 string、var r1,r2 *string
//
//代码
//   var email string
//   _, err = o.SelectValue(`SELECT email FROM res_user WHERE id=#{Id};`, 1)(&email)
//   if err != nil {
// 	   log.Println(err)
//   }
//   log.Printf("email: %s \n", email)
//结果
//   email: test@foxmail.com
func (o *osmBase) SelectValue(sql string, params ...interface{}) func(containers ...interface{}) (int64, error) {
	return o.selectBySQL(sql, resultTypeValue, params)
}

// SelectValues 执行查询sql
//
// 查出的结果为多行,并存入不定长的变量上(...，每个都为array)，元素可以是指针，如var r1,r2 []string、var r1,r2 []*string都允许
//
//代码
//   var emails []string
//   _, err = o.SelectValues(`SELECT email FROM res_user WHERE id in #{Ids};`, []int64{1, 2})(&emails)
//   if err != nil {
// 	   log.Println(err)
//   }
//   log.Printf("emails: %v \n", emails)
//结果
//   emails: [test@foxmail.com test@foxmail.com]
func (o *osmBase) SelectValues(sql string, params ...interface{}) func(containers ...interface{}) (int64, error) {
	return o.selectBySQL(sql, resultTypeValues, params)
}

// SelectStruct 执行查询sql
//
// 查出的结果为单行,并存入struct，可以是指针，如var r User、var r *User
//
//代码
//   var users []ResUser
//   _, err = o.SelectStruct(`SELECT id,email,create_time FROM res_user WHERE id=#{Id};`, 1)(&users)
//   if err != nil {
// 	   log.Println(err)
//   }
//   log.Printf("user: %#v \n", users)
//结果
//   user: ResUser{ID:1, Email:"test@foxmail.com", Mobile:"", Nickname:""}
func (o *osmBase) SelectStruct(sql string, params ...interface{}) func(containers ...interface{}) (int64, error) {
	return o.selectBySQL(sql, resultTypeStruct, params)
}

// SelectStructs 执行查询sql
//
// 查出的结果为多行,并存入struct array，元素可以是指针，如var r []User、var r []*User
//
//代码
//   var users []ResUser
//   _, err = o.SelectStructs(`SELECT id,email,create_time FROM res_user WHERE id=#{Id};`, 1)(&users)
//   if err != nil {
// 	   log.Println(err)
//   }
//   log.Printf("users: %#v \n", users)
//结果
//   users: []ResUser{ResUser{ID:1, Email:"test@foxmail.com", Mobile:"", Nickname:""}}
func (o *osmBase) SelectStructs(sql string, params ...interface{}) func(containers ...interface{}) (int64, error) {
	return o.selectBySQL(sql, resultTypeStructs, params)
}

// SelectKVS 执行查询sql
//
// 查出的结果为多行,每行有两个字段,前者为key,后者为value,存入map (双列)，Key、Value可以是指针，如var r map[string]time.Time、var r map[*string]time.Time、var r map[string]*time.Time
//
//代码
//   var idEmailMap = map[int64]string{}
//   _, err = o.SelectKVS(`SELECT id,email FROM res_user WHERE id in #{Ids};`, []int64{1, 2})(&idEmailMap)
//   if err != nil {
// 	  log.Println(err)
//   }
//   log.Printf("idEmailMap: %v \n", idEmailMap)
//结果
//   idEmailMap: map[1:test@foxmail.com 2:test@foxmail.com]
func (o *osmBase) SelectKVS(sql string, params ...interface{}) func(containers ...interface{}) (int64, error) {
	return o.selectBySQL(sql, resultTypeKvs, params)
}

func (o *osmBase) selectBySQL(sql, resultType string, params []interface{}) func(containers ...interface{}) (int64, error) {
	sql, sqlParams, err := o.readSQLParamsBySQL(sql, params...)

	if err != nil {
		return func(containers ...interface{}) (int64, error) {
			return 0, err
		}
	}
	callback := func(containers ...interface{}) (int64, error) {
		var err error
		switch resultType {
		case resultTypeStructs:
			if len(containers) == 1 {
				return resultStructs(o, sql, sql, sqlParams, containers[0])
			}
			err = fmt.Errorf("sql '%s' error : resultTypeStructs ,len(containers) != 1", sql)
		case resultTypeStruct:
			if len(containers) == 1 {
				return resultStruct(o, sql, sql, sqlParams, containers[0])
			}
			err = fmt.Errorf("sql '%s' error : resultTypeStruct ,len(containers) != 1", sql)
		case resultTypeValue:
			if len(containers) > 0 {
				return resultValue(o, sql, sql, sqlParams, containers)
			}
			err = fmt.Errorf("sql '%s' error : resultTypeValue ,len(containers) < 1", sql)
		case resultTypeValues:
			if len(containers) > 0 {
				return resultValues(o, sql, sql, sqlParams, containers)
			}
			err = fmt.Errorf("sql '%s' error : resultTypeValues ,len(containers) < 1", sql)
		case resultTypeKvs:
			if len(containers) == 1 {
				return resultKvs(o, sql, sql, sqlParams, containers[0])
			}
			err = fmt.Errorf("sql '%s' error : resultTypeKvs ,len(containers) != 1", sql)
		}

		if err == nil {
			err = fmt.Errorf("sql '%s' error : sql resultTypeType no in ['value','struct','values','structs','kvs']", sql)
		}
		return 0, err
	}
	return callback
}

func (o *osmBase) readSQLParamsBySQL(sqlOrg string, params ...interface{}) (sql string, sqlParams []interface{}, err error) {
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
		defer func() {
			if ShowSQL {
				go logger.Printf(`sql:"%s", params:"%+v", dbSQL:"%s", dbParams:"%+v"`, sqlOrg, param, sql, sqlParams)
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
				logger.Printf("sql read error \"%v\"", markSQLError(sqlOrg, errorIndex))
				return
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
					err = fmt.Errorf("sql '%s' error : Key '%s' no exist", sqlOrg, paramName.content)
					return
				}
			}
		case kind == reflect.Struct:
			for _, paramName := range paramNames {
				firstChar := paramName.content[0]
				if firstChar < 'A' || firstChar > 'Z' {
					err = fmt.Errorf("sql '%s' error : Field '%s' unexported", sqlOrg, paramName.content)
					return
				}
				vv := v.FieldByName(paramName.content)
				if vv.IsValid() {
					setDataToParamName(paramName, vv)
				} else {
					err = fmt.Errorf("sql '%s' error : Field '%s' no exist", sqlOrg, paramName.content)
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
		sql = sqlOrg
		if ShowSQL {
			go logger.Printf(`sql:"%s"`, sqlOrg)
		}
	}
	return
}
