package osm

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockOsm(t *testing.T) (*osmBase, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	opts := &Options{}
	opts.tidy()
	o := &osmBase{
		db:      db,
		dbType:  dbTypeMysql,
		options: opts,
	}
	return o, mock
}

func TestInsert(t *testing.T) {
	t.Run("mysql returns lastInsertId", func(t *testing.T) {
		o, mock := newMockOsm(t)
		mock.ExpectPrepare("INSERT INTO user").
			ExpectExec().
			WithArgs("test@example.com").
			WillReturnResult(sqlmock.NewResult(42, 1))

		insertID, count, err := o.Insert("INSERT INTO user (email) VALUES (#{Email})", map[string]interface{}{"Email": "test@example.com"})
		if err != nil {
			t.Fatal(err)
		}
		if insertID != 42 {
			t.Errorf("insertID: got %d, want 42", insertID)
		}
		if count != 1 {
			t.Errorf("count: got %d, want 1", count)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("postgres skips lastInsertId", func(t *testing.T) {
		o, mock := newMockOsm(t)
		o.dbType = dbTypePostgres
		mock.ExpectPrepare("INSERT INTO user").
			ExpectExec().
			WithArgs("test@example.com").
			WillReturnResult(sqlmock.NewResult(0, 1))

		insertID, count, err := o.Insert("INSERT INTO user (email) VALUES (#{Email})", map[string]interface{}{"Email": "test@example.com"})
		if err != nil {
			t.Fatal(err)
		}
		if insertID != 0 {
			t.Errorf("insertID: expected 0 for postgres, got %d", insertID)
		}
		if count != 1 {
			t.Errorf("count: got %d, want 1", count)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("returns error on exec failure", func(t *testing.T) {
		o, mock := newMockOsm(t)
		mock.ExpectPrepare("INSERT INTO user").
			ExpectExec().
			WithArgs("test").
			WillReturnError(sql.ErrNoRows)

		_, _, err := o.Insert("INSERT INTO user (name) VALUES (#{Name})", map[string]interface{}{"Name": "test"})
		if err == nil {
			t.Fatal("expected error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestUpdate(t *testing.T) {
	o, mock := newMockOsm(t)
	mock.ExpectPrepare("UPDATE user").
		ExpectExec().
		WithArgs("new@example.com", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	count, err := o.Update("UPDATE user SET email = #{Email} WHERE id = #{ID}", map[string]interface{}{"Email": "new@example.com", "ID": 1})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count: got %d, want 1", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestDelete(t *testing.T) {
	o, mock := newMockOsm(t)
	mock.ExpectPrepare("DELETE FROM user").
		ExpectExec().
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 2))

	count, err := o.Delete("DELETE FROM user WHERE id IN #{Ids}", []int64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count: got %d, want 2", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUpdateMulti(t *testing.T) {
	o, mock := newMockOsm(t)
	// #{Email} appears twice, #{Id}, #{Id2} — 4 params total
	mock.ExpectExec("UPDATE user").
		WithArgs("test@foxmail.com", 3, "test@foxmail.com", 4).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := o.UpdateMulti("UPDATE user SET email='#{Email}' WHERE id = #{Id}; UPDATE user SET email='#{Email}' WHERE id = #{Id2};",
		map[string]interface{}{"Email": "test@foxmail.com", "Id": 3, "Id2": 4})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// SQL parsing error returns early - no DB interaction
func TestResultParseErrorReturnsEarly(t *testing.T) {
	o, _ := newMockOsm(t)

	t.Run("Select with parse error", func(t *testing.T) {
		sr := o.Select("SELECT * FROM t WHERE id = #{missing", 1)
		_, err := sr.Struct(nil)
		if err == nil {
			t.Fatal("expected error for unclosed #{")
		}
	})

	t.Run("Delete with parse error", func(t *testing.T) {
		_, err := o.Delete("DELETE FROM t WHERE id = #{missing", 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Insert with parse error", func(t *testing.T) {
		_, _, err := o.Insert("INSERT INTO t (id) VALUES (#{missing", 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Update with parse error", func(t *testing.T) {
		_, err := o.Update("UPDATE t SET x = #{missing", 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("UpdateMulti with parse error", func(t *testing.T) {
		err := o.UpdateMulti("UPDATE t SET x = #{missing", 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

type testUser struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func TestResultStruct(t *testing.T) {
	t.Run("single row found", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "name", "email"}).
			AddRow(1, "Alice", "alice@example.com")
		mock.ExpectQuery("SELECT id, name, email FROM user").
			WithArgs(1).
			WillReturnRows(rows)

		var user testUser
		sr := o.Select("SELECT id, name, email FROM user WHERE id = #{id}", 1)
		count, err := sr.Struct(&user)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("count: got %d, want 1", count)
		}
		if user.ID != 1 || user.Name != "Alice" || user.Email != "alice@example.com" {
			t.Errorf("got %+v", user)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("no rows returns 0 count", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "name", "email"})
		mock.ExpectQuery("SELECT id, name, email FROM user").
			WithArgs(999).
			WillReturnRows(rows)

		var user testUser
		sr := o.Select("SELECT id, name, email FROM user WHERE id = #{id}", 999)
		count, err := sr.Struct(&user)
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("count: got %d, want 0", count)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("pointer struct", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Bob")
		mock.ExpectQuery("SELECT id, name FROM user").
			WithArgs(1).
			WillReturnRows(rows)

		var user *testUser
		sr := o.Select("SELECT id, name FROM user WHERE id = #{id}", 1)
		_, err := sr.Struct(&user)
		if err != nil {
			t.Fatal(err)
		}
		if user == nil || user.ID != 1 || user.Name != "Bob" {
			t.Errorf("got %+v", user)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("non-pointer container returns error", func(t *testing.T) {
		o, mock := newMockOsm(t)
		sr := o.Select("SELECT id FROM user WHERE id = #{id}", 1)
		var user testUser
		_, err := sr.Struct(user)
		if err == nil {
			t.Fatal("expected error for non-pointer")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestResultStructs(t *testing.T) {
	t.Run("multiple rows", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob").
			AddRow(3, "Charlie")
		mock.ExpectQuery("SELECT id, name FROM user").
			WillReturnRows(rows)

		var users []testUser
		sr := o.Select("SELECT id, name FROM user")
		count, err := sr.Structs(&users)
		if err != nil {
			t.Fatal(err)
		}
		if count != 3 {
			t.Errorf("count: got %d, want 3", count)
		}
		if len(users) != 3 {
			t.Fatalf("len: got %d, want 3", len(users))
		}
		if users[0].ID != 1 || users[0].Name != "Alice" {
			t.Errorf("users[0]: got %+v", users[0])
		}
		if users[2].Name != "Charlie" {
			t.Errorf("users[2]: got %+v", users[2])
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("no rows returns empty slice", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "name"})
		mock.ExpectQuery("SELECT id, name FROM user").
			WillReturnRows(rows)

		var users []testUser
		sr := o.Select("SELECT id, name FROM user")
		count, err := sr.Structs(&users)
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("count: got %d, want 0", count)
		}
		if len(users) != 0 {
			t.Errorf("len: got %d, want 0", len(users))
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("pointer slice elements", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice")
		mock.ExpectQuery("SELECT id, name FROM user").
			WillReturnRows(rows)

		var users []*testUser
		sr := o.Select("SELECT id, name FROM user")
		count, err := sr.Structs(&users)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("count: got %d, want 1", count)
		}
		if users[0] == nil || users[0].ID != 1 || users[0].Name != "Alice" {
			t.Errorf("got %+v", users[0])
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("non-pointer container returns error", func(t *testing.T) {
		o, mock := newMockOsm(t)
		sr := o.Select("SELECT id FROM user")
		var users []testUser
		_, err := sr.Structs(users)
		if err == nil {
			t.Fatal("expected error for non-pointer")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestResultValue(t *testing.T) {
	t.Run("single column", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"email"}).
			AddRow("alice@example.com")
		mock.ExpectQuery("SELECT email FROM user").
			WithArgs(1).
			WillReturnRows(rows)

		var email string
		sr := o.Select("SELECT email FROM user WHERE id = #{id}", 1)
		count, err := sr.Value(&email)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("count: got %d, want 1", count)
		}
		if email != "alice@example.com" {
			t.Errorf("got %q, want alice@example.com", email)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("multiple columns", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice")
		mock.ExpectQuery("SELECT id, name FROM user").
			WithArgs(1).
			WillReturnRows(rows)

		var id int
		var name string
		sr := o.Select("SELECT id, name FROM user WHERE id = #{id}", 1)
		_, err := sr.Value(&id, &name)
		if err != nil {
			t.Fatal(err)
		}
		if id != 1 || name != "Alice" {
			t.Errorf("got id=%d, name=%s", id, name)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("non-pointer container returns error", func(t *testing.T) {
		o, mock := newMockOsm(t)
		var email string
		sr := o.Select("SELECT email FROM user")
		_, err := sr.Value(email)
		if err == nil {
			t.Fatal("expected error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestResultValues(t *testing.T) {
	t.Run("multiple rows single column", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"email"}).
			AddRow("alice@example.com").
			AddRow("bob@example.com")
		mock.ExpectQuery("SELECT email FROM user").
			WillReturnRows(rows)

		var emails []string
		sr := o.Select("SELECT email FROM user")
		count, err := sr.Values(&emails)
		if err != nil {
			t.Fatal(err)
		}
		if count != 2 {
			t.Errorf("count: got %d, want 2", count)
		}
		if len(emails) != 2 || emails[0] != "alice@example.com" || emails[1] != "bob@example.com" {
			t.Errorf("got %v", emails)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("no rows returns empty slice", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"email"})
		mock.ExpectQuery("SELECT email FROM user").
			WillReturnRows(rows)

		var emails []string
		sr := o.Select("SELECT email FROM user")
		count, err := sr.Values(&emails)
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("count: got %d, want 0", count)
		}
		if len(emails) != 0 {
			t.Errorf("expected empty slice, got %v", emails)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestSelectInt(t *testing.T) {
	o, mock := newMockOsm(t)
	rows := sqlmock.NewRows([]string{"count"}).
		AddRow(42)
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(rows)

	result, err := o.Select("SELECT COUNT(*) FROM user").Int()
	if err != nil {
		t.Fatal(err)
	}
	if result != 42 {
		t.Errorf("got %d, want 42", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSelectInt64(t *testing.T) {
	o, mock := newMockOsm(t)
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(int64(100))
	mock.ExpectQuery("SELECT id FROM user").
		WithArgs(1).
		WillReturnRows(rows)

	result, err := o.Select("SELECT id FROM user WHERE id = #{id}", 1).Int64()
	if err != nil {
		t.Fatal(err)
	}
	if result != 100 {
		t.Errorf("got %d, want 100", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSelectString(t *testing.T) {
	o, mock := newMockOsm(t)
	rows := sqlmock.NewRows([]string{"email"}).
		AddRow("alice@example.com")
	mock.ExpectQuery("SELECT email FROM user").
		WithArgs(1).
		WillReturnRows(rows)

	result, err := o.Select("SELECT email FROM user WHERE id = #{id}", 1).String()
	if err != nil {
		t.Fatal(err)
	}
	if result != "alice@example.com" {
		t.Errorf("got %q, want alice@example.com", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSelectStrings(t *testing.T) {
	o, mock := newMockOsm(t)
	rows := sqlmock.NewRows([]string{"email"}).
		AddRow("alice@example.com").
		AddRow("bob@example.com")
	mock.ExpectQuery("SELECT email FROM user").
		WillReturnRows(rows)

	result, err := o.Select("SELECT email FROM user").Strings()
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 || result[0] != "alice@example.com" || result[1] != "bob@example.com" {
		t.Errorf("got %v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSelectInts(t *testing.T) {
	o, mock := newMockOsm(t)
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1).
		AddRow(2).
		AddRow(3)
	mock.ExpectQuery("SELECT id FROM user").
		WillReturnRows(rows)

	result, err := o.Select("SELECT id FROM user").Ints()
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 || result[0] != 1 || result[2] != 3 {
		t.Errorf("got %v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSelectBool(t *testing.T) {
	o, mock := newMockOsm(t)
	rows := sqlmock.NewRows([]string{"exists"}).
		AddRow(true)
	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(rows)

	result, err := o.Select("SELECT EXISTS(SELECT 1 FROM user WHERE id = #{id})", 1).Bool()
	if err != nil {
		t.Fatal(err)
	}
	if result != true {
		t.Errorf("got %v, want true", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSelectUint64(t *testing.T) {
	o, mock := newMockOsm(t)
	rows := sqlmock.NewRows([]string{"val"}).
		AddRow(uint64(999))
	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	result, err := o.Select("SELECT 999").Uint64()
	if err != nil {
		t.Fatal(err)
	}
	if result != 999 {
		t.Errorf("got %d, want 999", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSelectFloat64(t *testing.T) {
	o, mock := newMockOsm(t)
	rows := sqlmock.NewRows([]string{"avg"}).
		AddRow(3.14)
	mock.ExpectQuery("SELECT AVG").
		WillReturnRows(rows)

	result, err := o.Select("SELECT AVG(score) FROM user").Float64()
	if err != nil {
		t.Fatal(err)
	}
	if result != 3.14 {
		t.Errorf("got %f, want 3.14", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSelectResultKvs(t *testing.T) {
	t.Run("id to email map", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "email"}).
			AddRow(1, "alice@example.com").
			AddRow(2, "bob@example.com")
		mock.ExpectQuery("SELECT id, email FROM user").
			WillReturnRows(rows)

		result := map[int64]string{}
		sr := o.Select("SELECT id, email FROM user")
		count, err := sr.Kvs(&result)
		if err != nil {
			t.Fatal(err)
		}
		if count != 2 {
			t.Errorf("count: got %d, want 2", count)
		}
		if len(result) != 2 {
			t.Fatalf("map len: got %d, want 2", len(result))
		}
		if result[1] != "alice@example.com" || result[2] != "bob@example.com" {
			t.Errorf("got %v", result)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("non-map container returns error", func(t *testing.T) {
		o, mock := newMockOsm(t)
		sr := o.Select("SELECT id, email FROM user")
		var result []string
		_, err := sr.Kvs(&result)
		if err == nil {
			t.Fatal("expected error for non-map")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestSelectStringsResult(t *testing.T) {
	t.Run("columns and datas", func(t *testing.T) {
		o, mock := newMockOsm(t)
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("1", "Alice").
			AddRow("2", "Bob")
		mock.ExpectQuery("SELECT id, name FROM user").
			WillReturnRows(rows)

		columns, datas, err := o.Select("SELECT id, name FROM user").ColumnsAndData()
		if err != nil {
			t.Fatal(err)
		}
		if len(columns) != 2 || columns[0] != "id" || columns[1] != "name" {
			t.Errorf("columns: got %v", columns)
		}
		if len(datas) != 2 || datas[0][0] != "1" || datas[1][1] != "Bob" {
			t.Errorf("datas: got %v", datas)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("query error returns early", func(t *testing.T) {
		o, mock := newMockOsm(t)
		mock.ExpectQuery("SELECT").
			WillReturnError(sql.ErrNoRows)

		_, _, err := o.Select("SELECT id FROM user").ColumnsAndData()
		if err == nil {
			t.Fatal("expected error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestSelectResultConvenienceTyped(t *testing.T) {
	o, mock := newMockOsm(t)

	t.Run("Int32", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"cnt"}).AddRow(int32(10))
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT 10").Int32()
		if err != nil {
			t.Fatal(err)
		}
		if r != 10 {
			t.Errorf("got %d", r)
		}
	})

	t.Run("Float32", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"price"}).AddRow(float32(1.5))
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT 1.5").Float32()
		if err != nil {
			t.Fatal(err)
		}
		if r != 1.5 {
			t.Errorf("got %f", r)
		}
	})

	t.Run("Uint", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"cnt"}).AddRow(uint(7))
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT 7").Uint()
		if err != nil {
			t.Fatal(err)
		}
		if r != 7 {
			t.Errorf("got %d", r)
		}
	})

	t.Run("Bools", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"active"}).AddRow(true).AddRow(false)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT active FROM t").Bools()
		if err != nil {
			t.Fatal(err)
		}
		if len(r) != 2 || r[0] != true || r[1] != false {
			t.Errorf("got %v", r)
		}
	})

	t.Run("Int64s", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1)).AddRow(int64(2))
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT id FROM t").Int64s()
		if err != nil {
			t.Fatal(err)
		}
		if len(r) != 2 || r[0] != 1 || r[1] != 2 {
			t.Errorf("got %v", r)
		}
	})

	t.Run("Float64s", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"score"}).AddRow(1.1).AddRow(2.2)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT score FROM t").Float64s()
		if err != nil {
			t.Fatal(err)
		}
		if len(r) != 2 || r[0] != 1.1 || r[1] != 2.2 {
			t.Errorf("got %v", r)
		}
	})

	t.Run("Int32s", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"v"}).AddRow(int32(1)).AddRow(int32(2))
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT v FROM t").Int32s()
		if err != nil {
			t.Fatal(err)
		}
		if len(r) != 2 || r[0] != 1 || r[1] != 2 {
			t.Errorf("got %v", r)
		}
	})

	t.Run("Float32s", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"v"}).AddRow(float32(1.5)).AddRow(float32(2.5))
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT v FROM t").Float32s()
		if err != nil {
			t.Fatal(err)
		}
		if len(r) != 2 || r[0] != 1.5 || r[1] != 2.5 {
			t.Errorf("got %v", r)
		}
	})

	t.Run("Uints", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"v"}).AddRow(uint(1)).AddRow(uint(2))
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT v FROM t").Uints()
		if err != nil {
			t.Fatal(err)
		}
		if len(r) != 2 || r[0] != 1 || r[1] != 2 {
			t.Errorf("got %v", r)
		}
	})

	t.Run("Uint64s", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"v"}).AddRow(uint64(1)).AddRow(uint64(2))
		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		r, err := o.Select("SELECT v FROM t").Uint64s()
		if err != nil {
			t.Fatal(err)
		}
		if len(r) != 2 || r[0] != 1 || r[1] != 2 {
			t.Errorf("got %v", r)
		}
	})
}

func TestSelectResultErrorPropagation(t *testing.T) {
	t.Run("parse error in Select propagates to all methods", func(t *testing.T) {
		o, mock := newMockOsm(t)
		sr := o.Select("SELECT * FROM t WHERE id = #{missing", 1)

		if _, err := sr.Struct(nil); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Structs(nil); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Kvs(nil); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Value(nil); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Values(nil); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.String(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Strings(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Int(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Ints(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Int64(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Int64s(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Float64(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Float64s(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Int32(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Int32s(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Float32(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Float32s(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Uint(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Uints(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Uint64(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Uint64s(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Bool(); err == nil {
			t.Error("expected error")
		}
		if _, err := sr.Bools(); err == nil {
			t.Error("expected error")
		}
		if _, _, err := sr.ColumnsAndData(); err == nil {
			t.Error("expected error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("wrong container for Struct returns error before query", func(t *testing.T) {
		o, _ := newMockOsm(t)
		// struct validation fails before query is made
		_, err := o.Select("SELECT id FROM t WHERE id = #{id}", 1).Struct(nil)
		if err == nil {
			t.Fatal("expected error for struct with nil (non-pointer)")
		}
	})

	t.Run("Structs with non-slice pointer", func(t *testing.T) {
		o, _ := newMockOsm(t)
		_, err := o.Select("SELECT id FROM t").Structs(&testUser{})
		if err == nil {
			t.Fatal("expected error for struct (not slice)")
		}
	})

	t.Run("Value with no containers", func(t *testing.T) {
		o, mock := newMockOsm(t)
		_, err := o.Select("SELECT id FROM t WHERE id = #{id}", 1).Value()
		if err == nil {
			t.Fatal("expected error for no containers")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("Values with no containers", func(t *testing.T) {
		o, mock := newMockOsm(t)
		_, err := o.Select("SELECT id FROM t").Values()
		if err == nil {
			t.Fatal("expected error for no containers")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("Kvs with non-pointer", func(t *testing.T) {
		o, mock := newMockOsm(t)
		_, err := o.Select("SELECT id, name FROM t").Kvs(map[int64]string{})
		if err == nil {
			t.Fatal("expected error for non-pointer")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestNativePlaceholderMySQL(t *testing.T) {
	o, mock := newMockOsm(t)

	t.Run("Insert with native ? placeholder", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO user").
			ExpectExec().
			WithArgs("test@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		insertID, count, err := o.Insert("INSERT INTO user (email) VALUES (?)", "test@example.com")
		if err != nil {
			t.Fatal(err)
		}
		if insertID != 1 {
			t.Errorf("insertID: got %d, want 1", insertID)
		}
		if count != 1 {
			t.Errorf("count: got %d, want 1", count)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("Select with native $1 placeholder on postgres", func(t *testing.T) {
		o, mock := newMockOsm(t)
		o.dbType = dbTypePostgres

		rows := sqlmock.NewRows([]string{"email"}).
			AddRow("alice@example.com")
		mock.ExpectQuery("SELECT email FROM user").
			WithArgs(1).
			WillReturnRows(rows)

		var email string
		sr := o.Select("SELECT email FROM user WHERE id = $1", 1)
		_, err := sr.Value(&email)
		if err != nil {
			t.Fatal(err)
		}
		if email != "alice@example.com" {
			t.Errorf("got %q", email)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestGetCallerInfo(t *testing.T) {
	info := getCallerInfo(0)
	if info == "" {
		t.Error("expected non-empty caller info")
	}
}

func TestSelectBySQLDefaultCase(t *testing.T) {
	o, mock := newMockOsm(t)
	// selectBySQL with unknown resultType
	cb := o.selectBySQL("test", "SELECT 1", resultType(99), nil)
	_, err := cb()
	if err == nil {
		t.Fatal("expected error for unknown result type")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
