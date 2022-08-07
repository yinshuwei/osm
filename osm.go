package osm

// osm (Object Sql Mapping) 极简sql工具，支持MySQL和PostgreSQL。

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	dbTypeMysql    = 0
	dbTypePostgres = 1
	dbTypeMssql    = 2
)

var (
	errorLogger Logger = &DefaultLogger{}
	infoLogger  Logger = &DefaultLogger{}

	// ShowSQL 显示执行的sql，用于调试，使用logger打印
	showSQL = false

	// SlowLogDuration 慢查询时间阈值
	slowLogDuration = 500 * time.Millisecond
)

type dbRunner interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type osmBase struct {
	db     dbRunner
	dbType int
}

// Osm 对象，通过Struct、Map、Array、value等对象以及Sql Map来操作数据库。可以开启事务。
type Osm struct {
	osmBase
}

// Tx 与Osm对象一样，不过是在事务中进行操作
type Tx struct {
	osmBase
}

func ConfLogger(_infoLogger, _errorLogger Logger, _showSQL bool, _slowLogDuration time.Duration) {
	infoLogger = _infoLogger
	errorLogger = _errorLogger
	showSQL = _showSQL
	slowLogDuration = _slowLogDuration
}

type Options struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// New 创建一个新的Osm，这个过程会打开数据库连接。
//
//driverName是数据库驱动名称如"mysql".
//dataSource是数据库连接信息如"root:root@/51jczj?charset=utf8".
//options是数据连接的参数，MaxIdleConns, MaxOpenConns, ConnMaxLifetime, ConnMaxIdleTime
//
//如：
//  o, err := osm.New("mysql", "root:root@/51jczj?charset=utf8", 50, 100)
func New(driverName, dataSource string, options Options) (*Osm, error) {
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
				errorLogger.Warn("osm Ping fail", map[string]string{"error": err.Error()})
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

	if options.MaxIdleConns > 0 {
		db.SetMaxIdleConns(options.MaxIdleConns)
	}

	if options.MaxOpenConns > 0 {
		db.SetMaxOpenConns(options.MaxOpenConns)
	}

	if options.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(options.ConnMaxLifetime)
	}

	if options.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(options.ConnMaxIdleTime)
	}

	return osm, nil
}

// Begin 打开事务
//
//如：
//  tx, err := o.Begin()
func (o *Osm) Begin() (*Tx, error) {
	tx := new(Tx)
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
