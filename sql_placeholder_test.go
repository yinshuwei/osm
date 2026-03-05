package osm

import (
	"testing"
)

// TestDatabasePlaceholderFormats 测试不同数据库的占位符格式
func TestDatabasePlaceholderFormats(t *testing.T) {
	testCases := []struct {
		name          string
		dbType        int
		sqlTemplate   string
		params        interface{}
		expectedSQL   string
		expectedCount int
	}{
		// MySQL 占位符
		{
			name:        "MySQL简单查询",
			dbType:      dbTypeMysql,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id}",
			params:      1,
			expectedSQL: "SELECT * FROM users WHERE id = ?",
		},
		{
			name:        "MySQL多参数",
			dbType:      dbTypeMysql,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id} AND name = #{Name}",
			params:      map[string]interface{}{"Id": 1, "Name": "test"},
			expectedSQL: "SELECT * FROM users WHERE id = ? AND name = ?",
		},
		{
			name:        "MySQL IN查询",
			dbType:      dbTypeMysql,
			sqlTemplate: "SELECT * FROM users WHERE id IN #{Ids}",
			params:      map[string]interface{}{"Ids": []int{1, 2, 3}},
			expectedSQL: "SELECT * FROM users WHERE id IN (?,?,?)",
		},

		// SQLite 占位符（与 MySQL 相同）
		{
			name:        "SQLite简单查询",
			dbType:      dbTypeSqlite,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id}",
			params:      1,
			expectedSQL: "SELECT * FROM users WHERE id = ?",
		},

		// PostgreSQL 占位符
		{
			name:        "PostgreSQL简单查询",
			dbType:      dbTypePostgres,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id}",
			params:      1,
			expectedSQL: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:        "PostgreSQL多参数",
			dbType:      dbTypePostgres,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id} AND name = #{Name}",
			params:      map[string]interface{}{"Id": 1, "Name": "test"},
			expectedSQL: "SELECT * FROM users WHERE id = $1 AND name = $2",
		},
		{
			name:        "PostgreSQL IN查询",
			dbType:      dbTypePostgres,
			sqlTemplate: "SELECT * FROM users WHERE id IN #{Ids}",
			params:      map[string]interface{}{"Ids": []int{1, 2, 3}},
			expectedSQL: "SELECT * FROM users WHERE id IN ($1,$2,$3)",
		},

		// SQL Server 占位符（与 PostgreSQL 相同）
		{
			name:        "MSSQL简单查询",
			dbType:      dbTypeMssql,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id}",
			params:      1,
			expectedSQL: "SELECT * FROM users WHERE id = $1",
		},

		// Oracle 占位符
		{
			name:        "Oracle简单查询",
			dbType:      dbTypeOracle,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id}",
			params:      1,
			expectedSQL: "SELECT * FROM users WHERE id = :1",
		},
		{
			name:        "Oracle多参数",
			dbType:      dbTypeOracle,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id} AND name = #{Name}",
			params:      map[string]interface{}{"Id": 1, "Name": "test"},
			expectedSQL: "SELECT * FROM users WHERE id = :1 AND name = :2",
		},
		{
			name:        "Oracle IN查询",
			dbType:      dbTypeOracle,
			sqlTemplate: "SELECT * FROM users WHERE id IN #{Ids}",
			params:      map[string]interface{}{"Ids": []int{1, 2, 3}},
			expectedSQL: "SELECT * FROM users WHERE id IN (:1,:2,:3)",
		},

		// TiDB 占位符（与 MySQL 相同）
		{
			name:        "TiDB简单查询",
			dbType:      dbTypeTiDB,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id}",
			params:      1,
			expectedSQL: "SELECT * FROM users WHERE id = ?",
		},

		// CockroachDB 占位符（与 PostgreSQL 相同）
		{
			name:        "CockroachDB简单查询",
			dbType:      dbTypeCockroach,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id}",
			params:      1,
			expectedSQL: "SELECT * FROM users WHERE id = $1",
		},

		// ClickHouse 占位符（与 PostgreSQL 相同）
		{
			name:        "ClickHouse简单查询",
			dbType:      dbTypeClickHouse,
			sqlTemplate: "SELECT * FROM users WHERE id = #{Id}",
			params:      1,
			expectedSQL: "SELECT * FROM users WHERE id = $1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o := &osmBase{
				dbType:  tc.dbType,
				options: &Options{},
			}

			sql, sqlParams, err := o.readSQLParamsBySQL("test", tc.sqlTemplate, tc.params)
			if err != nil {
				t.Errorf("解析失败: %v", err)
				return
			}

			if sql != tc.expectedSQL {
				t.Errorf("SQL 不匹配\n期望: %s\n实际: %s", tc.expectedSQL, sql)
			}

			if tc.expectedCount > 0 && len(sqlParams) != tc.expectedCount {
				t.Errorf("参数数量不匹配\n期望: %d\n实际: %d", tc.expectedCount, len(sqlParams))
			}
		})
	}
}

// BenchmarkDatabasePlaceholderFormats 性能测试：不同数据库占位符生成
func BenchmarkDatabasePlaceholderFormats(b *testing.B) {
	o := &osmBase{
		options: &Options{},
	}

	// MySQL
	b.Run("MySQL", func(b *testing.B) {
		o.dbType = dbTypeMysql
		sql := "SELECT * FROM users WHERE id = #{Id} AND name = #{Name} AND status = #{Status}"
		params := map[string]interface{}{"Id": 1, "Name": "test", "Status": "active"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL("test", sql, params)
		}
	})

	// PostgreSQL
	b.Run("PostgreSQL", func(b *testing.B) {
		o.dbType = dbTypePostgres
		sql := "SELECT * FROM users WHERE id = #{Id} AND name = #{Name} AND status = #{Status}"
		params := map[string]interface{}{"Id": 1, "Name": "test", "Status": "active"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL("test", sql, params)
		}
	})

	// Oracle
	b.Run("Oracle", func(b *testing.B) {
		o.dbType = dbTypeOracle
		sql := "SELECT * FROM users WHERE id = #{Id} AND name = #{Name} AND status = #{Status}"
		params := map[string]interface{}{"Id": 1, "Name": "test", "Status": "active"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL("test", sql, params)
		}
	})

	// SQLite
	b.Run("SQLite", func(b *testing.B) {
		o.dbType = dbTypeSqlite
		sql := "SELECT * FROM users WHERE id = #{Id} AND name = #{Name} AND status = #{Status}"
		params := map[string]interface{}{"Id": 1, "Name": "test", "Status": "active"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL("test", sql, params)
		}
	})
}
