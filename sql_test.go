package osm

import (
	"testing"
)

func TestReadSQLParamsBySQL(t *testing.T) {
	o := &osmBase{options: &Options{}}

	tests := []struct {
		name       string
		sql        string
		params     []interface{}
		wantSQL    string
		wantParams []interface{}
	}{
		{
			name:       "single param",
			sql:        "SELECT * FROM table WHERE id = #{id}",
			params:     []interface{}{1},
			wantSQL:    "SELECT * FROM table WHERE id = ?",
			wantParams: []interface{}{1},
		},
		{
			name:       "multiple params",
			sql:        "SELECT * FROM table WHERE name = #{name} AND age = #{age}",
			params:     []interface{}{"John", 25},
			wantSQL:    "SELECT * FROM table WHERE name = ? AND age = ?",
			wantParams: []interface{}{"John", 25},
		},
		{
			name:       "array param with IN",
			sql:        "SELECT * FROM table WHERE ids IN #{ids}",
			params:     []interface{}{[]int{1, 2, 3}},
			wantSQL:    "SELECT * FROM table WHERE ids IN (?,?,?)",
			wantParams: []interface{}{1, 2, 3},
		},
		{
			name:       "struct param",
			sql:        "SELECT * FROM table WHERE name = #{Name} AND age = #{Age}",
			params:     []interface{}{struct{ Name string; Age int }{Name: "Alice", Age: 30}},
			wantSQL:    "SELECT * FROM table WHERE name = ? AND age = ?",
			wantParams: []interface{}{"Alice", 30},
		},
		{
			name:       "map param",
			sql:        "SELECT * FROM table WHERE name = #{name} AND age = #{age}",
			params:     []interface{}{map[string]interface{}{"name": "Bob", "age": 35}},
			wantSQL:    "SELECT * FROM table WHERE name = ? AND age = ?",
			wantParams: []interface{}{"Bob", 35},
		},
		{
			name:       "no params",
			sql:        "SELECT * FROM table",
			params:     nil,
			wantSQL:    "SELECT * FROM table",
			wantParams: nil,
		},
		{
			name:       "IN with args list",
			sql:        "SELECT * FROM table WHERE id IN #{ids} AND name = #{name}",
			params:     []interface{}{[]int{1, 2, 3}, "John"},
			wantSQL:    "SELECT * FROM table WHERE id IN (?,?,?) AND name = ?",
			wantParams: []interface{}{1, 2, 3, "John"},
		},
		{
			name:       "IN with struct field",
			sql:        "SELECT * FROM table WHERE id IN #{Ids} AND name = #{Name}",
			params:     []interface{}{struct{ Ids []int; Name string }{Ids: []int{1, 2, 3}, Name: "John"}},
			wantSQL:    "SELECT * FROM table WHERE id IN (?,?,?) AND name = ?",
			wantParams: []interface{}{1, 2, 3, "John"},
		},
		{
			name:       "IN with map, special chars in SQL",
			sql:        "SELECT *,'{' as a, '}' as b FROM table WHERE id IN #{Ids} AND name = #{Name}",
			params:     []interface{}{map[string]interface{}{"Ids": []int{1, 2, 3}, "Name": "John", "Abc": "test"}},
			wantSQL:    "SELECT *,'{' as a, '}' as b FROM table WHERE id IN (?,?,?) AND name = ?",
			wantParams: []interface{}{1, 2, 3, "John"},
		},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			gotSQL, gotParams, err := o.readSQLParamsBySQL("test", tc.sql, tc.params...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotSQL != tc.wantSQL {
				t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", gotSQL, tc.wantSQL)
			}
			if len(gotParams) != len(tc.wantParams) {
				t.Fatalf("params length mismatch: got %d, want %d", len(gotParams), len(tc.wantParams))
			}
			for i := range gotParams {
				if gotParams[i] != tc.wantParams[i] {
					t.Errorf("param[%d] mismatch: got %v, want %v", i, gotParams[i], tc.wantParams[i])
				}
			}
		})
	}
}

func TestSQLReplacements(t *testing.T) {
	o := &osmBase{
		options: &Options{
			SQLReplacements: map[string]string{
				"[TablePrefix]": "data_",
				"[Env]":         "prod",
			},
		},
	}
	o.options.tidy()

	oEmpty := &osmBase{options: &Options{}}

	tests := []struct {
		name    string
		osm     *osmBase
		sql     string
		params  []interface{}
		wantSQL string
	}{
		{
			name:    "table prefix replacement",
			osm:     o,
			sql:     "SELECT * FROM [TablePrefix]user WHERE id = #{id}",
			params:  []interface{}{1},
			wantSQL: "SELECT * FROM data_user WHERE id = ?",
		},
		{
			name:    "multiple replacements",
			osm:     o,
			sql:     "SELECT * FROM [TablePrefix]user_[Env] WHERE id = #{id}",
			params:  []interface{}{1},
			wantSQL: "SELECT * FROM data_user_prod WHERE id = ?",
		},
		{
			name:    "no replacement configured",
			osm:     oEmpty,
			sql:     "SELECT * FROM [TablePrefix]user WHERE id = #{id}",
			params:  []interface{}{1},
			wantSQL: "SELECT * FROM [TablePrefix]user WHERE id = ?",
		},
		{
			name:    "duplicate placeholders",
			osm:     o,
			sql:     "SELECT * FROM [TablePrefix]user1, [TablePrefix]user2 WHERE id = #{id}",
			params:  []interface{}{1},
			wantSQL: "SELECT * FROM data_user1, data_user2 WHERE id = ?",
		},
		{
			name:    "replacement before param parsing",
			osm:     o,
			sql:     "INSERT INTO [TablePrefix]user (name) VALUES (#{name})",
			params:  []interface{}{"test"},
			wantSQL: "INSERT INTO data_user (name) VALUES (?)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSQL, _, err := tc.osm.readSQLParamsBySQL("test", tc.sql, tc.params...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotSQL != tc.wantSQL {
				t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", gotSQL, tc.wantSQL)
			}
		})
	}
}

func TestNativeSQLPlaceholders(t *testing.T) {
	o := &osmBase{options: &Options{}}

	oWithReplacer := &osmBase{
		options: &Options{
			SQLReplacements: map[string]string{
				"[TablePrefix]": "data_",
			},
		},
	}
	oWithReplacer.options.tidy()

	tests := []struct {
		name       string
		osm        *osmBase
		sql        string
		params     []interface{}
		wantSQL    string
		wantParams []interface{}
	}{
		{
			name:       "MySQL ? placeholder",
			osm:        o,
			sql:        "SELECT * FROM table WHERE id = ? AND name = ?",
			params:     []interface{}{1, "John"},
			wantSQL:    "SELECT * FROM table WHERE id = ? AND name = ?",
			wantParams: []interface{}{1, "John"},
		},
		{
			name:       "PostgreSQL $1,$2 placeholder",
			osm:        o,
			sql:        "SELECT * FROM table WHERE id = $1 AND name = $2",
			params:     []interface{}{1, "John"},
			wantSQL:    "SELECT * FROM table WHERE id = $1 AND name = $2",
			wantParams: []interface{}{1, "John"},
		},
		{
			name:       "mixed ? and $ placeholder",
			osm:        o,
			sql:        "SELECT * FROM table WHERE id = $1 AND age > ?",
			params:     []interface{}{1, 25},
			wantSQL:    "SELECT * FROM table WHERE id = $1 AND age > ?",
			wantParams: []interface{}{1, 25},
		},
		{
			name:       "single param",
			osm:        o,
			sql:        "SELECT * FROM table WHERE id = ?",
			params:     []interface{}{1},
			wantSQL:    "SELECT * FROM table WHERE id = ?",
			wantParams: []interface{}{1},
		},
		{
			name:       "multiple params",
			osm:        o,
			sql:        "SELECT * FROM table WHERE id = ? AND name = ? AND age = ?",
			params:     []interface{}{1, "John", 25},
			wantSQL:    "SELECT * FROM table WHERE id = ? AND name = ? AND age = ?",
			wantParams: []interface{}{1, "John", 25},
		},
		{
			name:       "IN query with ?",
			osm:        o,
			sql:        "SELECT * FROM table WHERE id IN (?,?,?)",
			params:     []interface{}{1, 2, 3},
			wantSQL:    "SELECT * FROM table WHERE id IN (?,?,?)",
			wantParams: []interface{}{1, 2, 3},
		},
		{
			name:       "native with SQLReplacements",
			osm:        oWithReplacer,
			sql:        "SELECT * FROM [TablePrefix]user WHERE id = ?",
			params:     []interface{}{1},
			wantSQL:    "SELECT * FROM data_user WHERE id = ?",
			wantParams: []interface{}{1},
		},
		{
			name:       "named param still works",
			osm:        o,
			sql:        "SELECT * FROM table WHERE id = #{id}",
			params:     []interface{}{1},
			wantSQL:    "SELECT * FROM table WHERE id = ?",
			wantParams: []interface{}{1},
		},
		{
			name:       "no params",
			osm:        o,
			sql:        "SELECT * FROM table",
			params:     nil,
			wantSQL:    "SELECT * FROM table",
			wantParams: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSQL, gotParams, err := tc.osm.readSQLParamsBySQL("test", tc.sql, tc.params...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotSQL != tc.wantSQL {
				t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", gotSQL, tc.wantSQL)
			}
			if len(gotParams) != len(tc.wantParams) {
				t.Fatalf("params length mismatch: got %d, want %d", len(gotParams), len(tc.wantParams))
			}
			for i := range gotParams {
				if gotParams[i] != tc.wantParams[i] {
					t.Errorf("param[%d] mismatch: got %v, want %v", i, gotParams[i], tc.wantParams[i])
				}
			}
		})
	}
}

// BenchmarkReadSQLParamsBySQL benchmark with a complex query
func BenchmarkReadSQLParamsBySQL(b *testing.B) {
	o := &osmBase{options: &Options{}}

	sqlOrg := `SELECT id, user, name, age, status, address, other, field1, field2, field3, field4, field5, '#{' as a, '}' as b
	 FROM users WHERE id IN (#{ids}) AND name = #{name} AND age = #{age} AND status = #{status} AND address = #{address};`

	type Person struct {
		Ids     []int  `db:"ids"`
		Name    string `db:"name"`
		Age     int    `db:"age"`
		Status  string `db:"status"`
		Address string `db:"address"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = o.readSQLParamsBySQL("Bench", sqlOrg, Person{
			Ids: []int{1, 2, 3}, Name: "John", Age: 30, Status: "active", Address: "123 Main St",
		})
	}
}

func BenchmarkReadSQLParamsVariants(b *testing.B) {
	b.ReportAllocs()

	o := &osmBase{options: &Options{}}
	logPrefix := "BenchVariants"

	b.Run("struct_IN_3", func(b *testing.B) {
		type P struct {
			Ids     []int  `db:"ids"`
			Name    string `db:"name"`
			Age     int    `db:"age"`
			Status  string `db:"status"`
			Address string `db:"address"`
		}
		sql := `SELECT id, name FROM users WHERE id IN #{ids} AND name=#{name} AND age=#{age} AND status=#{status} AND address=#{address}`
		p := P{Ids: []int{1, 2, 3}, Name: "John", Age: 30, Status: "active", Address: "Main"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL(logPrefix, sql, p)
		}
	})

	b.Run("map_IN_10", func(b *testing.B) {
		sql := `SELECT * FROM t WHERE id IN #{ids} AND name=#{name}`
		m := map[string]interface{}{
			"ids":  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			"name": "Alice",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL(logPrefix, sql, m)
		}
	})

	b.Run("args_multi_values", func(b *testing.B) {
		sql := `SELECT * FROM t WHERE a=#{a} AND b=#{b} AND c=#{c} AND d=#{d} AND e=#{e}`
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL(logPrefix, sql, "A", 2, 3.14, true, "Z")
		}
	})

	b.Run("IN_100_pressure", func(b *testing.B) {
		ids := make([]int, 100)
		for i := range ids {
			ids[i] = i + 1
		}
		sql := `SELECT * FROM t WHERE id IN #{ids}`
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL(logPrefix, sql, map[string]interface{}{"ids": ids})
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		sql := `SELECT * FROM t WHERE id IN #{ids} AND name=#{name}`
		m := map[string]interface{}{"ids": []int{1, 2, 3, 4, 5}, "name": "Bob"}
		for pb.Next() {
			_, _, _ = o.readSQLParamsBySQL(logPrefix, sql, m)
		}
	})
}

func BenchmarkSQLReplacements(b *testing.B) {
	o := &osmBase{
		options: &Options{
			SQLReplacements: map[string]string{
				"[TablePrefix]": "data_",
				"[Env]":         "prod",
				"[Schema]":      "public",
			},
		},
	}
	o.options.tidy()

	sql := "SELECT * FROM [Schema].[TablePrefix]user_[Env] WHERE id = #{id}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = o.readSQLParamsBySQL("Bench", sql, 1)
	}
}

func BenchmarkSQLWithoutReplacements(b *testing.B) {
	o := &osmBase{options: &Options{}}
	sql := "SELECT * FROM user WHERE id = #{id}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = o.readSQLParamsBySQL("Bench", sql, 1)
	}
}

func BenchmarkNativeSQLPlaceholders(b *testing.B) {
	o := &osmBase{options: &Options{}}

	b.Run("MySQL_style", func(b *testing.B) {
		sql := "SELECT * FROM table WHERE id = ? AND name = ? AND age = ?"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL("Bench", sql, 1, "John", 25)
		}
	})

	b.Run("PostgreSQL_style", func(b *testing.B) {
		sql := "SELECT * FROM table WHERE id = $1 AND name = $2 AND age = $3"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL("Bench", sql, 1, "John", 25)
		}
	})

	b.Run("Comparison_Native_vs_Named", func(b *testing.B) {
		sqlNative := "SELECT * FROM table WHERE id = ? AND name = ? AND age = ?"
		sqlNamed := "SELECT * FROM table WHERE id = #{id} AND name = #{name} AND age = #{age}"
		params := map[string]interface{}{"id": 1, "name": "John", "age": 25}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL("Bench", sqlNative, 1, "John", 25)
			_, _, _ = o.readSQLParamsBySQL("Bench", sqlNamed, params)
		}
	})
}
