package gorm_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// dsn = `root:MySQL9.9@tcp(mysql57:3306)/gorm?charset=utf8&parseTime=True`
	dsn = `root:MySQL9.9@tcp(localhost:3357)/gorm?charset=utf8&parseTime=True`
)

func aTestLocal(t *testing.T) {

	var user User

	DB.Where(map[string]interface{}{
		`id`: 79,
	}).
		First(&user)

	fmt.Printf("\n%+v\n", user)
}

func aTestDriver(t *testing.T) {
	db, err := sql.Open(`mysql`, dsn)
	if err != nil {
		t.Fatalf(`open failed due to [%s]`, err)
	}

	defer db.Close()

	rows, err := db.Query(`SELECT name FROM users LIMIT 10`)
	if err != nil {
		t.Errorf(`query err %s`, err)
	}

	defer rows.Close()

	var v string

	for rows.Next() {
		err := rows.Scan(&v)
		if err != nil {
			t.Errorf(`Scan err %s`, err)
		}

		t.Logf(`v is [%v]`, v)
	}




	result, err := db.Exec(`UPDATE users SET age = age + 1 WHERE id IN (?, ?)`, 80, 81)
	if err != nil {
		t.Errorf(`exec err %s`, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Errorf(`RowsAffected err %s`, err)
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		t.Errorf(`LastInsertId err %s`, err)
	}

	t.Logf(`exec LastInsertId %d RowsAffected %d`, lastInsertId, rowsAffected, )




	result2, err := db.Exec(`SELECT name FROM users LIMIT 10`)
	if err != nil {
		t.Errorf(`exec err %s`, err)
	}

	rowsAffected2, err := result2.RowsAffected()
	if err != nil {
		t.Errorf(`RowsAffected err %s`, err)
	}

	lastInsertId2, err := result2.LastInsertId()
	if err != nil {
		t.Errorf(`LastInsertId err %s`, err)
	}

	t.Logf(`exec LastInsertId %d RowsAffected %d`, lastInsertId2, rowsAffected2, )

}
