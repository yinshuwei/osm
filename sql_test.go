package osm

import (
	"testing"
)

func TestReadSQLParamsBySQL(t *testing.T) {
	// 初始化 osmBase 实例
	osmBase := &osmBase{options: &Options{}}

	// 测试用例 1：正常情况，单个参数
	sql1, params1, err1 := osmBase.readSQLParamsBySQL("TestPrefix1", "SELECT * FROM table WHERE id = #{id}", 1)
	if err1 != nil {
		t.Errorf("Expected no error, got %v", err1)
	}
	expectedSQL1 := "SELECT * FROM table WHERE id = ?"
	if sql1 != expectedSQL1 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL1, sql1)
	}
	if len(params1) != 1 || params1[0] != 1 {
		t.Errorf("Expected params [1], got %v", params1)
	}

	// 测试用例 2：正常情况，多个参数
	sql2, params2, err2 := osmBase.readSQLParamsBySQL("TestPrefix2", "SELECT * FROM table WHERE name = #{name} AND age = #{age}", "John", 25)
	if err2 != nil {
		t.Errorf("Expected no error, got %v", err2)
	}
	expectedSQL2 := "SELECT * FROM table WHERE name = ? AND age = ?"
	if sql2 != expectedSQL2 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL2, sql2)
	}
	if len(params2) != 2 || params2[0] != "John" || params2[1] != 25 {
		t.Errorf("Expected params ['John', 25], got %v", params2)
	}

	// 测试用例 3：参数为数组
	sql3, params3, err3 := osmBase.readSQLParamsBySQL("TestPrefix3", "SELECT * FROM table WHERE ids IN #{ids}", []int{1, 2, 3})
	if err3 != nil {
		t.Errorf("Expected no error, got %v", err3)
	}
	expectedSQL3 := "SELECT * FROM table WHERE ids IN (?,?,?)"
	if sql3 != expectedSQL3 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL3, sql3)
	}
	if len(params3) != 3 || params3[0] != 1 || params3[1] != 2 || params3[2] != 3 {
		t.Errorf("Expected params [1, 2, 3], got %v", params3)
	}

	// 测试用例 4：参数为结构体
	type Person struct {
		Name string
		Age  int
	}
	person := Person{Name: "Alice", Age: 30}
	sql4, params4, err4 := osmBase.readSQLParamsBySQL("TestPrefix4", "SELECT * FROM table WHERE name = #{Name} AND age = #{Age}", person)
	if err4 != nil {
		t.Errorf("Expected no error, got %v", err4)
	}
	expectedSQL4 := "SELECT * FROM table WHERE name = ? AND age = ?"
	if sql4 != expectedSQL4 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL4, sql4)
	}
	if len(params4) != 2 || params4[0] != "Alice" || params4[1] != 30 {
		t.Errorf("Expected params ['Alice', 30], got %v", params4)
	}

	// 测试用例 5：参数为映射
	paramMap := make(map[string]interface{})
	paramMap["name"] = "Bob"
	paramMap["age"] = 35
	sql5, params5, err5 := osmBase.readSQLParamsBySQL("TestPrefix5", "SELECT * FROM table WHERE name = #{name} AND age = #{age}", paramMap)
	if err5 != nil {
		t.Errorf("Expected no error, got %v", err5)
	}
	expectedSQL5 := "SELECT * FROM table WHERE name = ? AND age = ?"
	if sql5 != expectedSQL5 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL5, sql5)
	}
	if len(params5) != 2 || params5[0] != "Bob" || params5[1] != 35 {
		t.Errorf("Expected params ['Bob', 35], got %v", params5)
	}

	// 测试用例 6：没有参数
	sql6, params6, err6 := osmBase.readSQLParamsBySQL("TestPrefix6", "SELECT * FROM table", nil)
	if err6 != nil {
		t.Errorf("Expected no error, got %v", err6)
	}
	expectedSQL6 := "SELECT * FROM table"
	if sql6 != expectedSQL6 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL6, sql6)
	}
	if len(params6) != 0 {
		t.Errorf("Expected no params, got %v", params6)
	}

	// 测试用例 7：in和其他参数混合
	sql7, params7, err7 := osmBase.readSQLParamsBySQL("TestPrefix7", "SELECT * FROM table WHERE id IN #{ids} AND name = #{name}", []int{1, 2, 3}, "John")
	if err7 != nil {
		t.Errorf("Expected no error, got %v", err7)
	}
	expectedSQL7 := "SELECT * FROM table WHERE id IN (?,?,?) AND name = ?"
	if sql7 != expectedSQL7 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL7, sql7)
	}
	if len(params7) != 4 || params7[0] != 1 || params7[1] != 2 || params7[2] != 3 || params7[3] != "John" {
		t.Errorf("Expected params [1, 2, 3, 'John'], got %v", params7)
	}

	// 测试用例 8：将in参数放入结构体中
	type Person2 struct {
		Ids  []int
		Name string
	}
	person2 := Person2{Ids: []int{1, 2, 3}, Name: "John"}
	sql8, params8, err8 := osmBase.readSQLParamsBySQL("TestPrefix8", "SELECT * FROM table WHERE id IN #{Ids} AND name = #{Name}", person2)
	if err8 != nil {
		t.Errorf("Expected no error, got %v", err8)
	}
	expectedSQL8 := "SELECT * FROM table WHERE id IN (?,?,?) AND name = ?"
	if sql8 != expectedSQL8 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL8, sql8)
	}
	if len(params8) != 4 || params8[0] != 1 || params8[1] != 2 || params8[2] != 3 || params8[3] != "John" {
		t.Errorf("Expected params [1, 2, 3, 'John'], got %v", params8)
	}

	// 测试用例 9：将in参数放入映射中
	paramMap2 := make(map[string]interface{})
	paramMap2["Ids"] = []int{1, 2, 3}
	paramMap2["Name"] = "John"
	paramMap2["Abc"] = "test"
	sql9, params9, err9 := osmBase.readSQLParamsBySQL("TestPrefix9", "SELECT *,'{' as a, '}' as b FROM table WHERE id IN #{Ids} AND name = #{Name}", paramMap2)
	if err9 != nil {
		t.Errorf("Expected no error, got %v", err9)
	}
	expectedSQL9 := "SELECT *,'{' as a, '}' as b FROM table WHERE id IN (?,?,?) AND name = ?"
	if sql9 != expectedSQL9 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL9, sql9)
	}
	if len(params9) != 4 || params9[0] != 1 || params9[1] != 2 || params9[2] != 3 || params9[3] != "John" {
		t.Errorf("Expected params [1, 2, 3, 'John'], got %v", params9)
	}
}

// BenchmarkReadSQLParamsBySQL 对 readSQLParamsBySQL 函数进行压力测试
func BenchmarkReadSQLParamsBySQL(b *testing.B) {
	// 初始化 osmBase 实例
	o := &osmBase{options: &Options{}}

	// 定义一个示例 SQL 语句
	sqlOrg := `SELECT id, user, name, age, status, address, other, field1, field2, field3, field4, field5, '#{' as a, '}' as b
	 FROM users WHERE id IN (#{ids}) AND name = #{name} AND age = #{age} AND status = #{status} AND address = #{address};`

	// 定义示例参数，使用结构体
	type Person struct {
		Ids     []int  `db:"ids"`
		Name    string `db:"name"`
		Age     int    `db:"age"`
		Status  string `db:"status"`
		Address string `db:"address"`
	}

	// 记录日志前缀
	logPrefix := "Benchmark"

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		_, _, _ = o.readSQLParamsBySQL(logPrefix, sqlOrg, Person{
			Ids:     []int{1, 2, 3},
			Name:    "John",
			Age:     30,
			Status:  "active",
			Address: "123 Main St",
		})
	}
}

// BenchmarkReadSQLParamsVariants 针对不同参数形态/IN长度/容器类型的性能对比
func BenchmarkReadSQLParamsVariants(b *testing.B) {
	b.ReportAllocs()

	// 公共对象
	o := &osmBase{options: &Options{}}
	logPrefix := "BenchVariants"

	// 场景1：struct + 多字段 + IN(3)
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

	// 场景2：map + IN(10)
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

	// 场景3：slice 参数列表（非IN），多占位符
	b.Run("args_multi_values", func(b *testing.B) {
		sql := `SELECT * FROM t WHERE a=#{a} AND b=#{b} AND c=#{c} AND d=#{d} AND e=#{e}`
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = o.readSQLParamsBySQL(logPrefix, sql, "A", 2, 3.14, true, "Z")
		}
	})

	// 场景4：IN(100) 压力
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

	// 场景5：并行（模拟多协程）
	b.RunParallel(func(pb *testing.PB) {
		sql := `SELECT * FROM t WHERE id IN #{ids} AND name=#{name}`
		m := map[string]interface{}{"ids": []int{1, 2, 3, 4, 5}, "name": "Bob"}
		for pb.Next() {
			_, _, _ = o.readSQLParamsBySQL(logPrefix, sql, m)
		}
	})
}

// TestSQLReplacements 测试SQL占位符替换功能
func TestSQLReplacements(t *testing.T) {
	// 测试用例 1：配置了表前缀替换
	o := &osmBase{
		options: &Options{
			SQLReplacements: map[string]string{
				"[TablePrefix]": "data_",
				"[Env]":         "prod",
			},
		},
	}

	o.options.tidy()

	sql1 := "SELECT * FROM [TablePrefix]user WHERE id = #{id}"
	sql1, _, err1 := o.readSQLParamsBySQL("TestPrefix1", sql1, 1)
	if err1 != nil {
		t.Errorf("Expected no error, got %v", err1)
	}
	expectedSQL1 := "SELECT * FROM data_user WHERE id = ?"
	if sql1 != expectedSQL1 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL1, sql1)
	}

	// 测试用例 2：多个替换同时进行
	sql2 := "SELECT * FROM [TablePrefix]user_[Env] WHERE id = #{id}"
	sql2, _, err2 := o.readSQLParamsBySQL("TestPrefix2", sql2, 1)
	if err2 != nil {
		t.Errorf("Expected no error, got %v", err2)
	}
	expectedSQL2 := "SELECT * FROM data_user_prod WHERE id = ?"
	if sql2 != expectedSQL2 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL2, sql2)
	}

	// 测试用例 3：没有配置替换时，SQL保持不变
	oEmpty := &osmBase{options: &Options{}}
	sql3 := "SELECT * FROM [TablePrefix]user WHERE id = #{id}"
	sql3, _, err3 := oEmpty.readSQLParamsBySQL("TestPrefix3", sql3, 1)
	if err3 != nil {
		t.Errorf("Expected no error, got %v", err3)
	}
	expectedSQL3 := "SELECT * FROM [TablePrefix]user WHERE id = ?"
	if sql3 != expectedSQL3 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL3, sql3)
	}

	// 测试用例 4：替换多个相同的占位符
	sql4 := "SELECT * FROM [TablePrefix]user1, [TablePrefix]user2 WHERE id = #{id}"
	sql4, _, err4 := o.readSQLParamsBySQL("TestPrefix4", sql4, 1)
	if err4 != nil {
		t.Errorf("Expected no error, got %v", err4)
	}
	expectedSQL4 := "SELECT * FROM data_user1, data_user2 WHERE id = ?"
	if sql4 != expectedSQL4 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL4, sql4)
	}

	// 测试用例 5：替换在参数解析之前执行
	sql5 := "INSERT INTO [TablePrefix]user (name) VALUES (#{name})"
	sql5, params5, err5 := o.readSQLParamsBySQL("TestPrefix5", sql5, "test")
	if err5 != nil {
		t.Errorf("Expected no error, got %v", err5)
	}
	expectedSQL5 := "INSERT INTO data_user (name) VALUES (?)"
	if sql5 != expectedSQL5 {
		t.Errorf("Expected SQL '%s', got '%s'", expectedSQL5, sql5)
	}
	if len(params5) != 1 || params5[0] != "test" {
		t.Errorf("Expected params ['test'], got %v", params5)
	}
}

// BenchmarkSQLReplacements 对SQL替换功能进行性能测试
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

	sql := "SELECT * FROM [Schema].[TablePrefix]user_[Env] WHERE id = #{id}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = o.readSQLParamsBySQL("Benchmark", sql, 1)
	}
}

// BenchmarkSQLWithoutReplacements 测试不使用替换时的性能
func BenchmarkSQLWithoutReplacements(b *testing.B) {
	o := &osmBase{options: &Options{}}

	sql := "SELECT * FROM user WHERE id = #{id}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = o.readSQLParamsBySQL("Benchmark", sql, 1)
	}
}
