# OSM - Object SQL Mapping

English | [简体中文](./README.md)

osm (Object SQL Mapping) is a lightweight SQL toolkit written in Go, widely used in production environments.

**Supported Databases:** MySQL, PostgreSQL, SQL Server

## ✨ Core Philosophy

Provide a minimalist and elegant SQL operation interface to make database operations simpler and more intuitive:

```go
// Chain call style
users, err := o.Select("SELECT * FROM users WHERE age > #{Age}", 18).Structs(&users)

// Traditional style
count, err := o.SelectStructs("SELECT * FROM users WHERE age > #{Age}", 18)(&users)
```

## 🚀 Key Features

### Zero Dependencies
- Only depends on Go standard library, no third-party dependencies
- Lightweight design, easy to integrate and maintain

### Flexible Parameter Binding

Use `#{ParamName}` syntax for parameter binding, supporting multiple parameter types:

- **Sequential Parameters**: Automatically match by parameter order
- **Map Parameters**: Support `map[string]interface{}`
- **Struct Parameters**: Use struct directly as parameters
- **IN Queries**: Native support for SQL IN statements

### Rich Result Handling

Support various data receiving methods to meet different scenario requirements:

| Method Type | Description | Use Case |
|------------|-------------|----------|
| `Value` / `Values` | Single/Multiple rows with multiple columns | Query multiple columns with different types |
| `Struct` / `Structs` | Single/Multiple rows to struct | Object mapping |
| `String` / `Strings` | Single/Multiple strings | Simple field queries |
| `Int` / `Ints` | Single/Multiple integers | Statistical queries |
| `Float64` / `Float64s` | Single/Multiple floats | Numerical calculations |
| `Bool` / `Bools` | Single/Multiple booleans | Status flags |
| `Kvs` | Key-value pairs | Two-column data → Map |
| `ColumnsAndData` | Column names + Data rows | Data exchange/export |

### Intelligent Struct Mapping

- Prioritize reading `db` tags
- Smart field name conversion (supports common abbreviations like ID, URL, HTTP, etc.)
- Support nested structs
- Support pointer types (can represent NULL)
- [View complete field mapping rules](#field_column_mapping)

### SQL Placeholder Replacement

Support using placeholders in SQL statements that are automatically replaced with configured values at runtime. This is particularly useful for the following scenarios:

- **Table Prefix Replacement**: Add a unified prefix to table names
- **Database Schema Switching**: Dynamically switch database schema based on environment
- **Environment Identification**: Insert environment-related identifiers in SQL

**Configuration Example:**

```go
o, err := osm.New("mysql", "root:123456@/test?charset=utf8", osm.Options{
    SQLReplacements: map[string]string{
        "[TablePrefix]": "data_",   // Table prefix
        "[Schema]":      "prod",     // Database schema
        "[Env]":         "prod",     // Environment identifier
    },
})

// Placeholders in SQL will be automatically replaced
// SELECT * FROM [TablePrefix]users
// Actually executed: SELECT * FROM data_users
```

**Usage Examples:**

```go
// Single table query
o.Select("SELECT * FROM [TablePrefix]users WHERE id = #{Id}", 1)

// Multi-table JOIN
o.Select("SELECT * FROM [Schema].[TablePrefix]users u JOIN [TablePrefix]orders o ON u.id = o.user_id")

// Environment-related conditions
o.Select("SELECT * FROM [TablePrefix]config WHERE env = '[Env]'")
```

**Features:**
- ✅ High Performance: Only adds about 174ns overhead
- ✅ Zero Config Zero Cost: No performance impact when not configured
- ✅ Multiple Placeholder Support: Can replace multiple different placeholders simultaneously
- ✅ Repeated Placeholder Support: Same placeholder can be used multiple times in one SQL
- ✅ Pre-Execution Replacement: Replacement happens before parameter parsing, does not affect `#{...}` parameter binding

## 📦 Installation

```bash
go get github.com/yinshuwei/osm/v2
```

**go.mod:**
```go
require (
    github.com/yinshuwei/osm/v2 v2.0.8
)
```

## 📖 API Documentation

Complete documentation: https://pkg.go.dev/github.com/yinshuwei/osm/v2

## 🔗 Chain Call API

osm supports elegant chain calls. The `Select()` method returns a `SelectResult` object, allowing flexible result processing.

### Quick Start

```go
// Query struct list
var users []User
_, err := o.Select("SELECT * FROM users WHERE id > #{Id}", 1).Structs(&users)

// Query single value
count, err := o.Select("SELECT COUNT(*) FROM users").Int()

// Query string
email, err := o.Select("SELECT email FROM users WHERE id = #{Id}", 1).String()

// Query multiple columns with different types
var id int64
var username string
_, err := o.Select("SELECT id, username FROM users WHERE id = #{Id}", 1).Value(&id, &username)
```

### Complete Method List

#### 1. Struct and Structs - Struct Queries

**Struct** - Query single row and store in struct

```go
var user User
_, err := o.Select(`SELECT * FROM users WHERE id = #{Id}`, 1).Struct(&user)
```

**Structs** - Query multiple rows and store in struct slice

```go
var users []User
_, err := o.Select(`SELECT * FROM users`).Structs(&users)
```

#### 2. Value and Values - Multi-Column Value Queries

**Value** - Query single row with multiple columns

```go
var id int64
var email string
_, err := o.Select(`SELECT id, email FROM users WHERE id = #{Id}`, 1).Value(&id, &email)
```

**Values** - Query multiple rows with multiple columns

```go
var ids []int64
var emails []string
_, err := o.Select(`SELECT id, email FROM users`).Values(&ids, &emails)
```

#### 3. Kvs - Key-Value Pair Queries

Query multiple rows with two columns and store in map, first column as key, second as value

```go
var idEmailMap = map[int64]string{}
_, err := o.Select(`SELECT id, email FROM users`).Kvs(&idEmailMap)
```

#### 4. ColumnsAndData - Column Names and Data Queries

Query multiple rows and return column names and data (commonly used for data exchange)

```go
columns, datas, err := o.Select(`SELECT id, email FROM users`).ColumnsAndData()
// columns is []string
// datas is [][]string
```

#### 5. String and Strings - String Queries

**String** - Query single string value

```go
email, err := o.Select(`SELECT email FROM users WHERE id = #{Id}`, 1).String()
```

**Strings** - Query multiple string values

```go
emails, err := o.Select(`SELECT email FROM users`).Strings()
```

#### 6. Int and Ints - Integer Queries

**Int** - Query single int value

```go
count, err := o.Select(`SELECT COUNT(*) FROM users`).Int()
```

**Ints** - Query multiple int values

```go
ages, err := o.Select(`SELECT age FROM users`).Ints()
```

#### 7. Int32 and Int32s - 32-bit Integer Queries

**Int32** - Query single int32 value

```go
count, err := o.Select(`SELECT count FROM table WHERE id = #{Id}`, 1).Int32()
```

**Int32s** - Query multiple int32 values

```go
counts, err := o.Select(`SELECT count FROM table`).Int32s()
```

#### 8. Int64 and Int64s - 64-bit Integer Queries

**Int64** - Query single int64 value

```go
id, err := o.Select(`SELECT id FROM users WHERE email = #{Email}`, "test@example.com").Int64()
```

**Int64s** - Query multiple int64 values

```go
ids, err := o.Select(`SELECT id FROM users`).Int64s()
```

#### 9. Uint and Uints - Unsigned Integer Queries

**Uint** - Query single uint value

```go
count, err := o.Select(`SELECT COUNT(*) FROM users`).Uint()
```

**Uints** - Query multiple uint values

```go
counts, err := o.Select(`SELECT count FROM table`).Uints()
```

#### 10. Uint64 and Uint64s - 64-bit Unsigned Integer Queries

**Uint64** - Query single uint64 value

```go
id, err := o.Select(`SELECT id FROM users WHERE email = #{Email}`, "test@example.com").Uint64()
```

**Uint64s** - Query multiple uint64 values

```go
ids, err := o.Select(`SELECT id FROM users`).Uint64s()
```

#### 11. Float32 and Float32s - 32-bit Float Queries

**Float32** - Query single float32 value

```go
price, err := o.Select(`SELECT price FROM products WHERE id = #{Id}`, 1).Float32()
```

**Float32s** - Query multiple float32 values

```go
prices, err := o.Select(`SELECT price FROM products`).Float32s()
```

#### 12. Float64 and Float64s - 64-bit Float Queries

**Float64** - Query single float64 value

```go
avg, err := o.Select(`SELECT AVG(score) FROM users`).Float64()
```

**Float64s** - Query multiple float64 values

```go
scores, err := o.Select(`SELECT score FROM users`).Float64s()
```

#### 13. Bool and Bools - Boolean Queries

**Bool** - Query single boolean value

```go
isActive, err := o.Select(`SELECT is_active FROM users WHERE id = #{Id}`, 1).Bool()
```

**Bools** - Query multiple boolean values

```go
statuses, err := o.Select(`SELECT is_active FROM users`).Bools()
```

### 📊 Method Classification Summary

| Data Type | Single Value Method | Multiple Values Method | Typical Use |
|-----------|-------------------|---------------------|------------|
| **Generic Multi-Column** | `Value()` | `Values()` | Query multiple columns with different types |
| String | `String()` | `Strings()` | Names, emails, text fields |
| Integer | `Int()` | `Ints()` | Counts, ages, integers |
| 32-bit Integer | `Int32()` | `Int32s()` | Small range integers |
| 64-bit Integer | `Int64()` | `Int64s()` | IDs, large integers |
| Unsigned Integer | `Uint()` | `Uints()` | Positive integers |
| 64-bit Unsigned | `Uint64()` | `Uint64s()` | Large range positive integers |
| 32-bit Float | `Float32()` | `Float32s()` | Prices, ratios, low precision |
| 64-bit Float | `Float64()` | `Float64s()` | Scientific calculations, high precision |
| Boolean | `Bool()` | `Bools()` | Status flags, switches |
| Struct | `Struct()` | `Structs()` | Complete object mapping |
| Key-Value | - | `Kvs()` | Two-column data → Map |
| Generic Data | - | `ColumnsAndData()` | Data export, exchange |

### ⚠️ Important Notes

- **Multi-Column Query**: `Value()` and `Values()` methods support querying multiple columns with different types, suitable for querying multiple fields with different data types
- **Zero Value Handling**: Single value methods return type zero value (`0`, `""`, `false`) when no result
- **Empty Slice**: Multiple value methods return empty slice `[]` when no result
- **Data Exchange**: `ColumnsAndData()` returns all data as strings, suitable for cross-language data exchange
- **Key-Value**: `Kvs()` requires query result to have exactly two columns (first as key, second as value)

## 💡 Complete Examples

### Database Preparation

```sql
CREATE DATABASE test;
USE test;

CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` varchar(255) DEFAULT NULL,
  `nickname` varchar(45) DEFAULT NULL,
  `create_time` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='User table';
```

### Example Code

**Basic Example (osm_demo.go)**

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

// InfoLogger adapter for zap logger
type InfoLogger struct {
	zapLogger *zap.Logger
}

// WarnLogger adapter for zap logger
type WarnLogger struct {
	zapLogger *zap.Logger
}

// ErrorLogger adapter for zap logger
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

func (l *WarnLogger) Log(msg string, data map[string]string) {
	if l == nil || l.zapLogger == nil {
		return
	}
	l.zapLogger.Warn(msg, loggerFields(data)...)
}

// User model
type User struct {
	ID         int64
	Nickname   string `db:"name"`
	CreateTime time.Time
	EmailStruct // Anonymous property
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
		WarnLogger:      &WarnLogger{logger},  // Logger
		ErrorLogger:     &ErrorLogger{logger}, // Logger
		InfoLogger:      &InfoLogger{logger},  // Logger
		ShowSQL:         true,                 // bool
		SlowLogDuration: 0,                    // time.Duration
		SQLReplacements: map[string]string{    // SQL replacement map (optional)
			"[TablePrefix]": "data_",         // Table prefix
			"[Schema]":      "prod",          // Database schema
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// Insert data
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

	// Update data
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

	// Query data
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

	// Delete data
	count, err = o.Delete("DELETE FROM user WHERE email=#{Email}", user)
	if err != nil {
		logger.Error("test delete", zap.Error(err))
	}
	logger.Info("test delete", zap.Int64("count", count))

	// Close connection
	err = o.Close()
	if err != nil {
		logger.Error("close", zap.Error(err))
	}
}
```

**Execution Result:**

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

### Pointer Type Example

**Pointer Type Supporting NULL (osm_demo2.go)**

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

// InfoLogger adapter for zap logger
type InfoLogger struct {
	zapLogger *zap.Logger
}

// WarnLogger adapter for zap logger
type WarnLogger struct {
	zapLogger *zap.Logger
}

// ErrorLogger adapter for zap logger
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

func (l *WarnLogger) Log(msg string, data map[string]string) {
	if l == nil || l.zapLogger == nil {
		return
	}
	l.zapLogger.Warn(msg, loggerFields(data)...)
}

// User model
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
		WarnLogger:      &WarnLogger{logger},  // Logger
		ErrorLogger:     &ErrorLogger{logger}, // Logger
		InfoLogger:      &InfoLogger{logger},  // Logger
		ShowSQL:         true,                 // bool
		SlowLogDuration: 0,                    // time.Duration
		SQLReplacements: map[string]string{    // SQL replacement map (optional)
			"[TablePrefix]": "data_",
			"[Schema]":      "prod",
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// Insert data (Nickname is nil, representing NULL)
	{
		user := User{
			Email:      stringPoint("test@foxmail.com"),
			Nickname:   nil, // NULL value
			CreateTime: timePoint(time.Now()),
		}
		id, count, err := o.Insert("INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});", user)
		if err != nil {
			logger.Error("insert error", zap.Error(err))
		}
		logger.Info("test insert", zap.Int64("id", id), zap.Int64("count", count))
	}

	// Query data
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

	// Update data
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

	// Query again to verify
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

	// Delete data
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

	// Close connection
	{
		err = o.Close()
		if err != nil {
			logger.Error("close", zap.Error(err))
		}
	}
}

```

**Execution Result:**

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

## <a id="field_column_mapping"></a>🔤 Struct Field Mapping Rules

### Automatic Conversion Rules

SQL column names are automatically converted to Go struct field names through the following process:

1. **Split**: Split column name by `_`
   - Example: `user_email` → `user`, `email`

2. **Capitalize**: Capitalize the first letter of each part, lowercase the rest
   - Example: `user`, `email` → `User`, `Email`

3. **Concatenate**: Concatenate all parts
   - Example: `User`, `Email` → `UserEmail`

**Examples:**
```
user_name     → UserName
create_time   → CreateTime
user_id       → UserId or UserID
```

### Common Abbreviation Support

The following abbreviations support both forms (case insensitive), you can choose either in your struct:

**Example:** The `user_id` column can map to either `UserId` or `UserID` field

> ⚠️ **Note**: The same struct cannot contain both forms (e.g., both `UserId` and `UserID`), otherwise only one will be assigned.

**Supported Abbreviation List:**
```
  Acl  or   ACL
  Api  or   API
  Ascii  or ASCII
  Cpu  or   CPU
  Css  or   CSS
  Dns  or   DNS
  Eof  or   EOF
  Guid  or  GUID
  Html  or  HTML
  Http  or  HTTP
  Https  or HTTPS
  Id  or    ID
  Ip  or    IP
  Json  or  JSON
  Lhs  or   LHS
  Qps  or   QPS
  Ram  or   RAM
  Rhs  or   RHS
  Rpc  or   RPC
  Sla  or   SLA
  Smtp  or  SMTP
  Sql  or   SQL
  Ssh  or   SSH
  Tcp  or   TCP
  Tls  or   TLS
  Ttl  or   TTL
  Udp  or   UDP
  Ui  or    UI
  Uid  or   UID
  Uuid  or  UUID
  Uri  or   URI
  Url  or   URL
  Utf8  or  UTF8
  Vm  or    VM
  Xml  or   XML
  Xmpp  or  XMPP
  Xsrf  or  XSRF
  Xss  or   XSS
```

### Using db Tags

You can use the `db` tag to explicitly specify field-to-column mapping, tags have the highest priority:

```go
type User struct {
    ID       int64  `db:"user_id"`      // Explicitly map to user_id column
    Name     string `db:"user_name"`    // Explicitly map to user_name column
    Email    string                      // Automatically map to email column
    IsActive bool   `db:"is_active"`    // Explicitly map to is_active column
}
```

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

---

**If this project helps you, please give it a ⭐️ Star to support us!**

