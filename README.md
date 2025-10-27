# OSM - Object SQL Mapping

[English](./README_EN.md) | 简体中文

osm (Object SQL Mapping) 是用 Go 编写的轻量级 SQL 工具库，已在生产环境中广泛使用。

**支持的数据库:** MySQL、PostgreSQL、SQL Server

## ✨ 核心理念

提供极简且优雅的 SQL 操作接口，让数据库操作更加简单直观：

```go
// 链式调用风格
users, err := o.Select("SELECT * FROM users WHERE age > #{Age}", 18).Structs(&users)

// 传统风格
count, err := o.SelectStructs("SELECT * FROM users WHERE age > #{Age}", 18)(&users)
```

## 🚀 核心特性

### 零依赖
- 仅依赖 Go 标准库，无第三方依赖
- 轻量级设计，易于集成和维护

### 灵活的参数绑定

使用 `#{ParamName}` 语法进行参数绑定，支持多种参数类型：

- **顺序参数**: 按参数顺序自动匹配
- **Map 参数**: 支持 `map[string]interface{}`
- **Struct 参数**: 直接使用结构体作为参数
- **IN 查询**: 原生支持 SQL IN 语句

### 丰富的结果处理

支持多种数据接收方式，满足不同场景需求：

| 方法类型 | 说明 | 使用场景 |
|---------|------|---------|
| `Value` / `Values` | 单行/多行多列值 | 查询多列不同类型的值 |
| `Struct` / `Structs` | 单行/多行结构体 | 对象映射 |
| `String` / `Strings` | 单个/多个字符串 | 简单字段查询 |
| `Int` / `Ints` | 单个/多个整数 | 统计查询 |
| `Float64` / `Float64s` | 单个/多个浮点数 | 数值计算 |
| `Bool` / `Bools` | 单个/多个布尔值 | 状态标识 |
| `Kvs` | 键值对映射 | 双列数据 → Map |
| `ColumnsAndData` | 列名 + 数据行 | 数据交换/导出 |

### 智能的 Struct 映射

- 优先读取 `db` 标签
- 智能的字段名转换（支持常见缩写词，如 ID、URL、HTTP 等）
- 支持嵌套结构体
- 支持指针类型（可表示 NULL）
- [查看完整的字段映射规则](#field_column_mapping)

### SQL 占位符替换

支持在 SQL 语句中使用占位符，在运行时自动替换为配置的值。这对于以下场景特别有用：

- **表前缀替换**: 在表名前添加统一前缀
- **数据库 Schema 切换**: 根据环境动态切换数据库 schema
- **环境标识**: 在 SQL 中插入环境相关的标识

**配置示例:**

```go
o, err := osm.New("mysql", "root:123456@/test?charset=utf8", osm.Options{
    SQLReplacements: map[string]string{
        "[TablePrefix]": "data_",   // 表前缀
        "[Schema]":      "prod",     // 数据库schema
        "[Env]":         "prod",     // 环境标识
    },
})

// SQL 中的占位符会被自动替换
// SELECT * FROM [TablePrefix]users
// 实际执行: SELECT * FROM data_users
```

**使用示例:**

```go
// 单表查询
o.Select("SELECT * FROM [TablePrefix]users WHERE id = #{Id}", 1)

// 多表 JOIN
o.Select("SELECT * FROM [Schema].[TablePrefix]users u JOIN [TablePrefix]orders o ON u.id = o.user_id")

// 环境相关的条件
o.Select("SELECT * FROM [TablePrefix]config WHERE env = '[Env]'")
```

**特性:**
- ✅ 性能高效：仅增加约 174ns 的开销
- ✅ 零配置零开销：未配置时完全不影响性能
- ✅ 支持多占位符：可同时替换多个不同的占位符
- ✅ 支持重复占位符：同一个 SQL 中可以多次使用同一个占位符
- ✅ 执行前替换：替换发生在参数解析之前，不影响 `#{...}` 参数绑定

## 📦 安装

```bash
go get github.com/yinshuwei/osm/v2
```

**go.mod:**
```go
require (
    github.com/yinshuwei/osm/v2 v2.0.8
)
```

## 📖 API 文档

完整文档请访问: https://pkg.go.dev/github.com/yinshuwei/osm/v2

## 🔗 链式调用 API

osm 支持优雅的链式调用，通过 `Select()` 方法返回 `SelectResult` 对象，可灵活选择结果处理方式。

### 快速开始

```go
// 查询结构体列表
var users []User
_, err := o.Select("SELECT * FROM users WHERE id > #{Id}", 1).Structs(&users)

// 查询单个值
count, err := o.Select("SELECT COUNT(*) FROM users").Int()

// 查询字符串
email, err := o.Select("SELECT email FROM users WHERE id = #{Id}", 1).String()

// 查询多列不同类型的值
var id int64
var username string
_, err := o.Select("SELECT id, username FROM users WHERE id = #{Id}", 1).Value(&id, &username)
```

### 完整方法列表

#### 1. Struct 和 Structs - 结构体查询

**Struct** - 查询单行数据并存入struct

```go
var user User
_, err := o.Select(`SELECT * FROM users WHERE id = #{Id}`, 1).Struct(&user)
```

**Structs** - 查询多行数据并存入struct切片

```go
var users []User
_, err := o.Select(`SELECT * FROM users`).Structs(&users)
```

#### 2. Value 和 Values - 多列值查询

**Value** - 查询单行多列的值

```go
var id int64
var email string
_, err := o.Select(`SELECT id, email FROM users WHERE id = #{Id}`, 1).Value(&id, &email)
```

**Values** - 查询多行多列的值

```go
var ids []int64
var emails []string
_, err := o.Select(`SELECT id, email FROM users`).Values(&ids, &emails)
```

#### 3. Kvs - 键值对查询

查询多行两列数据并存入map，第一列作为key，第二列作为value

```go
var idEmailMap = map[int64]string{}
_, err := o.Select(`SELECT id, email FROM users`).Kvs(&idEmailMap)
```

#### 4. ColumnsAndData - 列名和数据查询

查询多行数据，返回列名和数据（常用于数据交换）

```go
columns, datas, err := o.Select(`SELECT id, email FROM users`).ColumnsAndData()
// columns 为 []string
// datas 为 [][]string
```

#### 5. String 和 Strings - 字符串查询

**String** - 查询单个字符串值

```go
email, err := o.Select(`SELECT email FROM users WHERE id = #{Id}`, 1).String()
```

**Strings** - 查询多个字符串值

```go
emails, err := o.Select(`SELECT email FROM users`).Strings()
```

#### 6. Int 和 Ints - 整数查询

**Int** - 查询单个int值

```go
count, err := o.Select(`SELECT COUNT(*) FROM users`).Int()
```

**Ints** - 查询多个int值

```go
ages, err := o.Select(`SELECT age FROM users`).Ints()
```

#### 7. Int32 和 Int32s - 32位整数查询

**Int32** - 查询单个int32值

```go
count, err := o.Select(`SELECT count FROM table WHERE id = #{Id}`, 1).Int32()
```

**Int32s** - 查询多个int32值

```go
counts, err := o.Select(`SELECT count FROM table`).Int32s()
```

#### 8. Int64 和 Int64s - 64位整数查询

**Int64** - 查询单个int64值

```go
id, err := o.Select(`SELECT id FROM users WHERE email = #{Email}`, "test@example.com").Int64()
```

**Int64s** - 查询多个int64值

```go
ids, err := o.Select(`SELECT id FROM users`).Int64s()
```

#### 9. Uint 和 Uints - 无符号整数查询

**Uint** - 查询单个uint值

```go
count, err := o.Select(`SELECT COUNT(*) FROM users`).Uint()
```

**Uints** - 查询多个uint值

```go
counts, err := o.Select(`SELECT count FROM table`).Uints()
```

#### 10. Uint64 和 Uint64s - 64位无符号整数查询

**Uint64** - 查询单个uint64值

```go
id, err := o.Select(`SELECT id FROM users WHERE email = #{Email}`, "test@example.com").Uint64()
```

**Uint64s** - 查询多个uint64值

```go
ids, err := o.Select(`SELECT id FROM users`).Uint64s()
```

#### 11. Float32 和 Float32s - 32位浮点数查询

**Float32** - 查询单个float32值

```go
price, err := o.Select(`SELECT price FROM products WHERE id = #{Id}`, 1).Float32()
```

**Float32s** - 查询多个float32值

```go
prices, err := o.Select(`SELECT price FROM products`).Float32s()
```

#### 12. Float64 和 Float64s - 64位浮点数查询

**Float64** - 查询单个float64值

```go
avg, err := o.Select(`SELECT AVG(score) FROM users`).Float64()
```

**Float64s** - 查询多个float64值

```go
scores, err := o.Select(`SELECT score FROM users`).Float64s()
```

#### 13. Bool 和 Bools - 布尔值查询

**Bool** - 查询单个布尔值

```go
isActive, err := o.Select(`SELECT is_active FROM users WHERE id = #{Id}`, 1).Bool()
```

**Bools** - 查询多个布尔值

```go
statuses, err := o.Select(`SELECT is_active FROM users`).Bools()
```

### 📊 方法分类总结

| 数据类型 | 单值方法 | 多值方法 | 典型用途 |
|---------|---------|---------|---------|
| **通用多列** | `Value()` | `Values()` | 查询多列不同类型的值 |
| 字符串 | `String()` | `Strings()` | 名称、邮箱等文本字段 |
| 整数 | `Int()` | `Ints()` | 计数、年龄等整数 |
| 32位整数 | `Int32()` | `Int32s()` | 小范围整数 |
| 64位整数 | `Int64()` | `Int64s()` | ID、大整数 |
| 无符号整数 | `Uint()` | `Uints()` | 正整数 |
| 64位无符号 | `Uint64()` | `Uint64s()` | 大范围正整数 |
| 32位浮点 | `Float32()` | `Float32s()` | 价格、比率等小精度 |
| 64位浮点 | `Float64()` | `Float64s()` | 科学计算、高精度数值 |
| 布尔值 | `Bool()` | `Bools()` | 状态标识、开关 |
| 结构体 | `Struct()` | `Structs()` | 完整对象映射 |
| 键值对 | - | `Kvs()` | 双列数据 → Map |
| 通用数据 | - | `ColumnsAndData()` | 数据导出、交换 |

### ⚠️ 重要说明

- **多列查询**: `Value()` 和 `Values()` 方法支持查询多列不同类型的值，适用于查询不同数据类型的多个字段
- **零值处理**: 单值方法在无结果时返回类型零值（`0`, `""`, `false`）
- **空切片**: 多值方法在无结果时返回空切片 `[]`
- **数据交换**: `ColumnsAndData()` 返回的数据全部为字符串类型，适合跨语言数据交换
- **键值对**: `Kvs()` 要求查询结果必须是两列（第一列为key，第二列为value）

## 💡 完整示例

### 数据库准备

```sql
CREATE DATABASE test;
USE test;

CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` varchar(255) DEFAULT NULL,
  `nickname` varchar(45) DEFAULT NULL,
  `create_time` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户表';
```

### 示例代码

**基础示例 (osm_demo.go)**

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
	Nickname   string `db:"name"`
	CreateTime time.Time
	EmailStruct // 匿名属性
}

type EmailStruct struct {
	Email string `db:"email"`
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
		SQLReplacements: map[string]string{    // SQL替换映射（可选）
			"[TablePrefix]": "data_",         // 表前缀
			"[Schema]":      "prod",          // 数据库schema
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// 插入数据
	user := User{
		EmailStruct: EmailStruct{
			Email: "test@foxmail.com",
		},
		Nickname:   "haha",
		CreateTime: time.Now(),
	}
	id, count, err := o.Insert("INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});", user)
	if err != nil {
		logger.Error("insert error", zap.Error(err))
	}
	logger.Info("test insert", zap.Int64("id", id), zap.Int64("count", count))

	// 更新数据
	user = User{
		EmailStruct: EmailStruct{
			Email: "test@foxmail.com",
		},
		Nickname: "hello",
	}
	count, err = o.Update("UPDATE user SET nickname=#{name} WHERE email=#{Email}", user)
	if err != nil {
		logger.Error("update error", zap.Error(err))
	}
	logger.Info("test update", zap.Int64("count", count))

	// 查询数据
	user = User{
		EmailStruct: EmailStruct{
			Email: "test@foxmail.com",
		},
	}
	var results []User
	count, err = o.SelectStructs("SELECT id,email,nickname,create_time FROM user WHERE email=#{Email} or email=#{email};", user)(&results)
	if err != nil {
		logger.Error("test select", zap.Error(err))
	}
	resultBytes, _ := json.Marshal(results)
	logger.Info("test select", zap.Int64("count", count), zap.ByteString("result", resultBytes))

	// 删除数据
	count, err = o.Delete("DELETE FROM user WHERE email=#{Email}", user)
	if err != nil {
		logger.Error("test delete", zap.Error(err))
	}
	logger.Info("test delete", zap.Int64("count", count))

	// 关闭连接
	err = o.Close()
	if err != nil {
		logger.Error("close", zap.Error(err))
	}
}
```

**运行结果:**

```log
2025-01-13T16:16:41.301+0800    INFO    osmtt/main.go:47        main.go:95, readSQLParamsBySQL showSql  {"dbParams": "[\"test@foxmail.com\",\"haha\",\"2025-01-13 16:16:41\"]", "sql": "INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});", "params": "{\"ID\":0,\"Nickname\":\"haha\",\"CreateTime\":\"2025-01-13T16:16:41.301032455+08:00\",\"Email\":\"test@foxmail.com\"}", "dbSql": "INSERT INTO user (email,nickname,create_time) VALUES (?,?,?);"}
2025-01-13T16:16:41.305+0800    INFO    osmtt/main.go:99        test insert     {"id": 11, "count": 1}
2025-01-13T16:16:41.305+0800    INFO    osmtt/main.go:47        main.go:108, readSQLParamsBySQL showSql {"dbParams": "[\"hello\",\"test@foxmail.com\"]", "sql": "UPDATE user SET nickname=#{name} WHERE email=#{Email}", "params": "{\"ID\":0,\"Nickname\":\"hello\",\"CreateTime\":\"0001-01-01T00:00:00Z\",\"Email\":\"test@foxmail.com\"}", "dbSql": "UPDATE user SET nickname=? WHERE email=?"}
2025-01-13T16:16:41.310+0800    INFO    osmtt/main.go:112       test update     {"count": 1}
2025-01-13T16:16:41.310+0800    INFO    osmtt/main.go:47        main.go:121, readSQLParamsBySQL showSql {"params": "{\"ID\":0,\"Nickname\":\"\",\"CreateTime\":\"0001-01-01T00:00:00Z\",\"Email\":\"test@foxmail.com\"}", "dbSql": "SELECT id,email,nickname,create_time FROM user WHERE email=? or email=?;", "dbParams": "[\"test@foxmail.com\",\"test@foxmail.com\"]", "sql": "SELECT id,email,nickname,create_time FROM user WHERE email=#{Email} or email=#{email};"}
2025-01-13T16:16:41.310+0800    INFO    osmtt/main.go:126       test select     {"count": 1, "result": "[{\"ID\":11,\"Nickname\":\"hello\",\"CreateTime\":\"2025-01-13T16:16:41+08:00\",\"Email\":\"test@foxmail.com\"}]"}
2025-01-13T16:16:41.310+0800    INFO    osmtt/main.go:47        main.go:129readSQLParamsBySQL showSql   {"dbParams": "[\"test@foxmail.com\"]", "sql": "DELETE FROM user WHERE email=#{Email}", "params": "{\"ID\":0,\"Nickname\":\"\",\"CreateTime\":\"0001-01-01T00:00:00Z\",\"Email\":\"test@foxmail.com\"}", "dbSql": "DELETE FROM user WHERE email=?"}
2025-01-13T16:16:41.313+0800    INFO    osmtt/main.go:133       test delete     {"count": 1}
```

### 指针类型示例

**指针类型支持 NULL (osm_demo2.go)**

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
		SQLReplacements: map[string]string{    // SQL替换映射（可选）
			"[TablePrefix]": "data_",
			"[Schema]":      "prod",
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// 插入数据（Nickname 为 nil，表示 NULL）
	{
		user := User{
			Email:      stringPoint("test@foxmail.com"),
			Nickname:   nil, // NULL 值
			CreateTime: timePoint(time.Now()),
		}
		id, count, err := o.Insert("INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});", user)
		if err != nil {
			logger.Error("insert error", zap.Error(err))
		}
		logger.Info("test insert", zap.Int64("id", id), zap.Int64("count", count))
	}

	// 查询数据
	{
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

	// 更新数据
	{
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

	// 再次查询验证
	{
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

	// 删除数据
	{
		user := User{
			Email: stringPoint("test@foxmail.com"),
		}
		count, err := o.Delete("DELETE FROM user WHERE email=#{Email}", user)
		if err != nil {
			logger.Error("test delete", zap.Error(err))
		}
		logger.Info("test delete", zap.Int64("count", count))
	}

	// 关闭连接
	{
		err = o.Close()
		if err != nil {
			logger.Error("close", zap.Error(err))
		}
	}
}

```

**运行结果:**

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

## <a id="field_column_mapping"></a>🔤 Struct 字段映射规则

### 自动转换规则

SQL 列名会自动转换为 Go 结构体字段名，转换过程如下：

1. **分隔**: 用 `_` 分隔列名 
   - 例: `user_email` → `user`, `email`

2. **首字母大写**: 每个部分转为首字母大写，其余小写
   - 例: `user`, `email` → `User`, `Email`

3. **拼接**: 拼接所有部分
   - 例: `User`, `Email` → `UserEmail`

**示例:**
```
user_name     → UserName
create_time   → CreateTime
user_id       → UserId 或 UserID
```

### 常见缩写词支持

以下缩写词支持两种形式（大小写不敏感），可在结构体中任选一种：

**示例:** `user_id` 列可映射到 `UserId` 或 `UserID` 字段

> ⚠️ **注意**: 同一结构体中不能同时包含两种形式（如同时有 `UserId` 和 `UserID`），否则只有一个会被赋值。

**支持的缩写词列表:**
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

### 使用 db 标签

可以使用 `db` 标签显式指定字段与列的映射关系，标签优先级最高：

```go
type User struct {
    ID       int64  `db:"user_id"`      // 显式映射到 user_id 列
    Name     string `db:"user_name"`    // 显式映射到 user_name 列
    Email    string                      // 自动映射到 email 列
    IsActive bool   `db:"is_active"`    // 显式映射到 is_active 列
}
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

**如果这个项目对你有帮助，请给个 ⭐️ Star 支持一下！**
