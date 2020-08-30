package gorm_test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"testing"
)

// run test:
// DEBUG=false GORM_DIALECT="mysql" GORM_DSN="root:MySQL9.9@tcp(mysql57:3306)/gorm?charset=utf8&parseTime=True" go test
// goland Environment in configuration:
// GORM_DSN=root:MySQL9.9@tcp(localhost:3357)/gorm?charset=utf8&parseTime=True;GORM_DIALECT=mysql;DEBUG=false

// TestLouvPureQueryFlow show pure query flow
func TestLouvPureQueryFlow(t *testing.T) {

	var user User

	err := DB.
		Debug().
		Where(map[string]interface{}{
			`id`: 79,
		}).
		Select([]string{`id`, `age`}).
		First(&user).
		Error

	// SELECT id, age FROM `users`  WHERE (`users`.`id` = 79) ORDER BY `users`.`id` ASC LIMIT 1

	if err != nil {
		t.Fatalf(`query err: %s`, err)
	}

	t.Logf(`get user, id: %d, age: %d`, user.Id, user.Age)
}

func TestLouvQuery(t *testing.T) {

	var user1 User

	DB.Callback().Query().Before(`gorm:after_query`).
		Register(`louv:test_query_callback`, func(scope *gorm.Scope) {
			scope.Log(fmt.Sprintf(`scope is %+v`, scope))
		})

	result := DB.
		Debug().
		Where(map[string]interface{}{
			`id`: 79,
		}).
		First(&user1)

	t.Logf("\n%+v\n", user1)
	t.Log(result.Error, result.RowsAffected)
}

func TestLouvUpdate(t *testing.T) {

	var user2 User

	result := DB.Model(user2).Debug().
		Where(map[string]interface{}{
			`id`: 79,
		}).
		Updates(map[string]interface{}{
			`age`: gorm.Expr(`age + 1`),
		})

	t.Log(result.Error, result.RowsAffected)
}

// TestLouvExecRows for question 2
func TestLouvExecRows(t *testing.T) {

	const sql = `SELECT id FROM users WHERE id = 79`

	rows, err := DB.Table(`users`).Exec(sql).Rows()

	if err != nil {
		t.Error(`Rows err: `, err)
		return
	}

	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	var id uint64

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			t.Errorf(`Scan err: [%s]`, err)
			continue
		}
		t.Logf(`Scan ok: [%d]`, id)
	}
}

// TestLouvTxRows for question 3
func TestLouvTxRows(t *testing.T) {

	tx := DB.Begin()
	if tx.Error != nil {
		t.Fatalf(`Begin err: [%v]`, tx.GetErrors())
	}

	// SELECT id FROM `users`  WHERE (id < 80) LIMIT 1
	rows, err := tx.Table(`users`).
		Where(`id < 80`).Select([]string{`id`}).Limit(1).
		Rows()

	if err != nil {
		t.Fatalf(`Rows err: [%s]`, err)
		return
	}

	defer rows.Close()

	var id uint64

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			t.Errorf(`Scan err: [%s]`, err)
			continue
		}
		t.Logf(`Scan ok: [%d]`, id)
		break
	}

	// 省略事务中的其他操作，例如 INSET / UPDATE

	err = tx.Commit().Error
	if err != nil {
		t.Fatalf(`Commit err: [%s]`, err)
	}
}
