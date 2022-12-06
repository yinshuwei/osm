# osm

osm(Object Sql Mapping) 是用 go 编写的极简 sql 工具，目前已在生产环境中使用，支持 MySQL、PostgreSQL 和 SQL Server。

设计的目的就是提供一种简单查询接口：

```go
_, err = o.SelectXXX(sql, params...)(&result...)
```

## 特性

- 不依赖标准库以外的三方库

- 灵活的 SQL 参数 #{ParamName}

  - 可以按参数顺序匹配
  - 可以匹配 map[string]interface{}
  - 可以匹配 struct
  - 可以使用 in

- 灵活的 SQL 结果接收

  - value (&username, &email) 查出的结果为单行,并存入不定长的变量上(...)
  - values (&usernameList, &emailList) 查出的结果为多行,并存入不定长的变量上(...，每个都为 array)
  - struct (&user) 查出的结果为单行,并存入 struct
  - structs (&users) 查出的结果为多行,并存入 struct array
  - kvs (&emailUsernameMap) 查出的结果为多行,每行有两个字段,前者为 key,后者为 value,存入 map (双列)
  - strings (&columns, &datas) 查出的结果为多行,并存入 columns，和 datas。columns 为[]string，datas 为[][]string（常用于数据交换，如给 python 的 pandas 提供数据源）

- [默认的 struct 字段名与 SQL 列名对应关系](#field_column_mapping)

## go.mod

```
require (
	github.com/yinshuwei/osm/v2 v2.0.2
)
```

## api doc

https://pkg.go.dev/github.com/yinshuwei/osm

## Quickstart

创建数据库

```sql
    create database test;
    use test;
```

创建 user 表

```sql
    CREATE TABLE `user` (
      `id` int(11) NOT NULL AUTO_INCREMENT,
      `email` varchar(255) DEFAULT NULL,
      `nickname` varchar(45) DEFAULT NULL,
      `create_time` varchar(255) DEFAULT NULL,
      PRIMARY KEY (`id`)
    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='user table';
```

执行 SQL 示例

osm_demo.go

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/yinshuwei/osm/v2"
    "go.uber.org/zap"
)

// InfoLogger 适配zap logger
type InfoLogger struct {
	zapLogger *zap.Logger
}

// WarnLoggor 适配zap logger
type WarnLoggor struct {
	zapLogger *zap.Logger
}

// ErrorLogger 适配zap logger
type ErrorLogger struct {
	zapLogger *zap.Logger
}

func loggerFields(data map[string]string) []zap.Field {
	var fields []zap.Field
	for key, val := range data {
		fields = append(fields, zap.String(key, val))
	}
	return fields
}

func (l *ErrorLogger) Log(msg string, data map[string]string) {
	if l == nil || l.zapLogger == nil {
		return
	}
	l.zapLogger.Error(msg, loggerFields(data)...)
}

func (l *InfoLogger) Log(msg string, data map[string]string) {
	if l == nil || l.zapLogger == nil {
		return
	}
	l.zapLogger.Info(msg, loggerFields(data)...)
}

func (l *WarnLoggor) Log(msg string, data map[string]string) {
	if l == nil || l.zapLogger == nil {
		return
	}
	l.zapLogger.Warn(msg, loggerFields(data)...)
}

// User 用户Model
type User struct {
    ID         int64
    Email      string
    Nickname   string
    CreateTime time.Time
}

func main() {
	logger, _ := zap.NewDevelopment()
	o, err := osm.New("mysql", "root:123456@/test?charset=utf8", osm.Options{
		MaxIdleConns:    0,                    // int
		MaxOpenConns:    0,                    // int
		ConnMaxLifetime: 0,                    // time.Duration
		ConnMaxIdleTime: 0,                    // time.Duration
		WarnLogger:      &WarnLoggor{logger},  // Logger
		ErrorLogger:     &ErrorLogger{logger}, // Logger
		InfoLogger:      &InfoLogger{logger},  // Logger
		ShowSQL:         true,                 // bool
		SlowLogDuration: 0,                    // time.Duration
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	//添加
	user := User{
		Email:      "test@foxmail.com",
		Nickname:   "haha",
		CreateTime: time.Now(),
	}
	id, count, err := o.Insert("INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});", user)
	if err != nil {
		logger.Error("insert error", zap.Error(err))
	}
	logger.Info("test insert", zap.Int64("id", id), zap.Int64("count", count))

	//更新
	user = User{
		Email:    "test@foxmail.com",
		Nickname: "hello",
	}
	count, err = o.Update("UPDATE user SET nickname=#{Nickname} WHERE email=#{Email}", user)
	if err != nil {
		logger.Error("update error", zap.Error(err))
	}
	logger.Info("test update", zap.Int64("count", count))

	//查询
	user = User{
		Email: "test@foxmail.com",
	}
	var results []User
	count, err = o.SelectStructs("SELECT id,email,nickname,create_time FROM user WHERE email=#{Email};", user)(&results)
	if err != nil {
		logger.Error("test select", zap.Error(err))
	}
	resultBytes, _ := json.Marshal(results)
	logger.Info("test select", zap.Int64("count", count), zap.ByteString("result", resultBytes))

	//删除
	count, err = o.Delete("DELETE FROM user WHERE email=#{Email}", user)
	if err != nil {
		logger.Error("test delete", zap.Error(err))
	}
	logger.Info("test delete", zap.Int64("count", count))

	//关闭
	err = o.Close()
	if err != nil {
		logger.Error("close", zap.Error(err))
	}
}
```

结果

```log
2022-02-18T18:46:54.347+0800    INFO    osm@v1.0.12-0.20220218102134-94591d978e3d/sql.go:311    readSQLParamsBySQL showSql, sql: INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});, params: {"ID":0,"Email":"test@foxmail.com","Nickname":"haha","CreateTime":"2022-02-18T18:46:54.346897988+08:00"}, dbSql: INSERT INTO user (email,nickname,create_time) VALUES (?,?,?);, dbParams: ["test@foxmail.com","haha","2022-02-18 18:46:54"]
2022-02-18T18:46:54.348+0800    INFO    osm_demo/main.go:40     test insert     {"id": 2, "count": 1}
2022-02-18T18:46:54.348+0800    INFO    osm@v1.0.12-0.20220218102134-94591d978e3d/sql.go:311    readSQLParamsBySQL showSql, sql: UPDATE user SET nickname=#{Nickname} WHERE email=#{Email}, params: {"ID":0,"Email":"test@foxmail.com","Nickname":"hello","CreateTime":"0001-01-01T00:00:00Z"}, dbSql: UPDATE user SET nickname=? WHERE email=?, dbParams: ["hello","test@foxmail.com"]
2022-02-18T18:46:54.349+0800    INFO    osm_demo/main.go:51     test update     {"count": 1}
2022-02-18T18:46:54.350+0800    INFO    osm@v1.0.12-0.20220218102134-94591d978e3d/sql.go:311    readSQLParamsBySQL showSql, sql: SELECT id,email,nickname,create_time FROM user WHERE email=#{Email};, params: {"ID":0,"Email":"test@foxmail.com","Nickname":"","CreateTime":"0001-01-01T00:00:00Z"}, dbSql: SELECT id,email,nickname,create_time FROM user WHERE email=?;, dbParams: ["test@foxmail.com"]
2022-02-18T18:46:54.350+0800    INFO    osm_demo/main.go:63     test select     {"count": 1, "result": "[{\"ID\":2,\"Email\":\"test@foxmail.com\",\"Nickname\":\"hello\",\"CreateTime\":\"2022-02-18T18:46:54+08:00\"}]"}
2022-02-18T18:46:54.350+0800    INFO    osm@v1.0.12-0.20220218102134-94591d978e3d/sql.go:311    readSQLParamsBySQL showSql, sql: DELETE FROM user WHERE email=#{Email}, params: {"ID":0,"Email":"test@foxmail.com","Nickname":"","CreateTime":"0001-01-01T00:00:00Z"}, dbSql: DELETE FROM user WHERE email=?, dbParams: ["test@foxmail.com"]
2022-02-18T18:46:54.351+0800    INFO    osm_demo/main.go:70     test delete     {"count": 1}
```

指针类型支持 nil 示例

osm_demo2.go

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/yinshuwei/osm/v2"
	"go.uber.org/zap"
)

// InfoLogger 适配zap logger
type InfoLogger struct {
	zapLogger *zap.Logger
}

// WarnLoggor 适配zap logger
type WarnLoggor struct {
	zapLogger *zap.Logger
}

// ErrorLogger 适配zap logger
type ErrorLogger struct {
	zapLogger *zap.Logger
}

func loggerFields(data map[string]string) []zap.Field {
	var fields []zap.Field
	for key, val := range data {
		fields = append(fields, zap.String(key, val))
	}
	return fields
}

func (l *ErrorLogger) Log(msg string, data map[string]string) {
	if l == nil || l.zapLogger == nil {
		return
	}
	l.zapLogger.Error(msg, loggerFields(data)...)
}

func (l *InfoLogger) Log(msg string, data map[string]string) {
	if l == nil || l.zapLogger == nil {
		return
	}
	l.zapLogger.Info(msg, loggerFields(data)...)
}

func (l *WarnLoggor) Log(msg string, data map[string]string) {
	if l == nil || l.zapLogger == nil {
		return
	}
	l.zapLogger.Warn(msg, loggerFields(data)...)
}

// User 用户Model
type User struct {
	ID         *int64
	Email      *string
	Nickname   *string
	CreateTime *time.Time
}

func stringPoint(t string) *string {
	return &t
}

func timePoint(t time.Time) *time.Time {
	return &t
}

func main() {
	logger, _ := zap.NewDevelopment()
	o, err := osm.New("mysql", "root:123456@/test?charset=utf8", osm.Options{
		MaxIdleConns:    0,                    // int
		MaxOpenConns:    0,                    // int
		ConnMaxLifetime: 0,                    // time.Duration
		ConnMaxIdleTime: 0,                    // time.Duration
		WarnLogger:      &WarnLoggor{logger},  // Logger
		ErrorLogger:     &ErrorLogger{logger}, // Logger
		InfoLogger:      &InfoLogger{logger},  // Logger
		ShowSQL:         true,                 // bool
		SlowLogDuration: 0,                    // time.Duration
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	{ //添加
		user := User{
			Email:      stringPoint("test@foxmail.com"),
			Nickname:   nil,
			CreateTime: timePoint(time.Now()),
		}
		id, count, err := o.Insert("INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});", user)
		if err != nil {
			logger.Error("insert error", zap.Error(err))
		}
		logger.Info("test insert", zap.Int64("id", id), zap.Int64("count", count))
	}

	{ //查询
		user := User{
			Email: stringPoint("test@foxmail.com"),
		}
		var results []User
		count, err := o.SelectStructs("SELECT id,email,nickname,create_time FROM user WHERE email=#{Email};", user)(&results)
		if err != nil {
			logger.Error("test select", zap.Error(err))
		}
		resultBytes, _ := json.Marshal(results)
		logger.Info("test select", zap.Int64("count", count), zap.ByteString("result", resultBytes))
	}

	{ // 更新
		user := User{
			Email:    stringPoint("test@foxmail.com"),
			Nickname: stringPoint("hello"),
		}
		count, err := o.Update("UPDATE user SET nickname=#{Nickname} WHERE email=#{Email}", user)
		if err != nil {
			logger.Error("update error", zap.Error(err))
		}
		logger.Info("test update", zap.Int64("count", count))
	}

	{ //查询
		user := User{
			Email: stringPoint("test@foxmail.com"),
		}
		var results []User
		count, err := o.SelectStructs("SELECT id,email,nickname,create_time FROM user WHERE email=#{Email};", user)(&results)
		if err != nil {
			logger.Error("test select", zap.Error(err))
		}
		resultBytes, _ := json.Marshal(results)
		logger.Info("test select", zap.Int64("count", count), zap.ByteString("result", resultBytes))
	}

	{ //删除
		user := User{
			Email: stringPoint("test@foxmail.com"),
		}
		count, err := o.Delete("DELETE FROM user WHERE email=#{Email}", user)
		if err != nil {
			logger.Error("test delete", zap.Error(err))
		}
		logger.Info("test delete", zap.Int64("count", count))
	}

	{ //关闭
		err = o.Close()
		if err != nil {
			logger.Error("close", zap.Error(err))
		}
	}
}

```

结果

```log
2022-02-21T11:42:44.591+0800    INFO    v2@v2.0.2/sql.go:311    readSQLParamsBySQL showSql, sql: INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});, params: {"ID":null,"Email":"test@foxmail.com","Nickname":null,"CreateTime":"2022-02-21T11:42:44.591619385+08:00"}, dbSql: INSERT INTO user (email,nickname,create_time) VALUES (?,?,?);, dbParams: ["test@foxmail.com",null,"2022-02-21T11:42:44.591619385+08:00"]
2022-02-21T11:42:44.596+0800    INFO    osm_demo/main.go:48     test insert     {"id": 10, "count": 1}
2022-02-21T11:42:44.596+0800    INFO    v2@v2.0.2/sql.go:311    readSQLParamsBySQL showSql, sql: SELECT id,email,nickname,create_time FROM user WHERE email=#{Email};, params: {"ID":null,"Email":"test@foxmail.com","Nickname":null,"CreateTime":null}, dbSql: SELECT id,email,nickname,create_time FROM user WHERE email=?;, dbParams: ["test@foxmail.com"]
2022-02-21T11:42:44.597+0800    INFO    osm_demo/main.go:61     test select     {"count": 1, "result": "[{\"ID\":10,\"Email\":\"test@foxmail.com\",\"Nickname\":\"\",\"CreateTime\":\"2022-02-21T03:42:44+08:00\"}]"}
2022-02-21T11:42:44.597+0800    INFO    v2@v2.0.2/sql.go:311    readSQLParamsBySQL showSql, sql: UPDATE user SET nickname=#{Nickname} WHERE email=#{Email}, params: {"ID":null,"Email":"test@foxmail.com","Nickname":"hello","CreateTime":null}, dbSql: UPDATE user SET nickname=? WHERE email=?, dbParams: ["hello","test@foxmail.com"]
2022-02-21T11:42:44.598+0800    INFO    osm_demo/main.go:73     test update     {"count": 1}
2022-02-21T11:42:44.598+0800    INFO    v2@v2.0.2/sql.go:311    readSQLParamsBySQL showSql, sql: SELECT id,email,nickname,create_time FROM user WHERE email=#{Email};, params: {"ID":null,"Email":"test@foxmail.com","Nickname":null,"CreateTime":null}, dbSql: SELECT id,email,nickname,create_time FROM user WHERE email=?;, dbParams: ["test@foxmail.com"]
2022-02-21T11:42:44.599+0800    INFO    osm_demo/main.go:86     test select     {"count": 1, "result": "[{\"ID\":10,\"Email\":\"test@foxmail.com\",\"Nickname\":\"hello\",\"CreateTime\":\"2022-02-21T03:42:44+08:00\"}]"}
2022-02-21T11:42:44.600+0800    INFO    v2@v2.0.2/sql.go:311    readSQLParamsBySQL showSql, sql: DELETE FROM user WHERE email=#{Email}, params: {"ID":null,"Email":"test@foxmail.com","Nickname":null,"CreateTime":null}, dbSql: DELETE FROM user WHERE email=?, dbParams: ["test@foxmail.com"]
2022-02-21T11:42:44.603+0800    INFO    osm_demo/main.go:97     test delete     {"count": 1}
```

## <a id="field_column_mapping" name="field_column_mapping">struct 字段名与 SQL 列名对应关系</a>

- 正常的转换过程

  用"\_"分隔 （例：XXX_YYY -> XXX,YYY）

  每个部分全部转为首字大写其余字符小写 （例：XXX,YYY -> Xxx,Yyy）

  拼接（例：Xxx,Yyy -> XxxYyy）

- 常见缩写单词，下面这些单词两种形式都可以，struct 上可以任选其一。
  比如"UserId"和"UserID"可以正常对应到"user_id"列上。但是同一个 struct 中不可以既有"UserId"成员又有"UserID"成员，如果同时存在只会有一个成员会被赋值。
  ```
  Acl  或   ACL
  Api  或   API
  Ascii  或 ASCII
  Cpu  或   CPU
  Css  或   CSS
  Dns  或   DNS
  Eof  或   EOF
  Guid  或  GUID
  Html  或  HTML
  Http  或  HTTP
  Https  或 HTTPS
  Id  或    ID
  Ip  或    IP
  Json  或  JSON
  Lhs  或   LHS
  Qps  或   QPS
  Ram  或   RAM
  Rhs  或   RHS
  Rpc  或   RPC
  Sla  或   SLA
  Smtp  或  SMTP
  Sql  或   SQL
  Ssh  或   SSH
  Tcp  或   TCP
  Tls  或   TLS
  Ttl  或   TTL
  Udp  或   UDP
  Ui  或    UI
  Uid  或   UID
  Uuid  或  UUID
  Uri  或   URI
  Url  或   URL
  Utf8  或  UTF8
  Vm  或    VM
  Xml  或   XML
  Xmpp  或  XMPP
  Xsrf  或  XSRF
  Xss  或   XSS
  ```
