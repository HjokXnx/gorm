package gorm_test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"testing"
)

// run test:
// DEBUG=false GORM_DIALECT="mysql" GORM_DSN="root:MySQL9.9@tcp(mysql57:3306)/gorm?charset=utf8&parseTime=True" go test

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

	t.Log("\n%+v\n", user1)
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

func TestLouvExecFind(t *testing.T) {

	const sql = `SELECT id FROM users WHERE id = 79`

	var user2 User

	rows, err := DB.Model(user2).Table(`users`).
		//Debug().
		Exec(sql).
		//Raw(sql).
		Rows()

	if err != nil {
		t.Error(`Rows err: `, err)
		return
	}

	// noinspection GoUnhandledErrorResult
	defer rows.Close()

	var scanErrAmt uint32

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			scanErrAmt++
			t.Errorf(`Scan err amt: %d, is: %s`, scanErrAmt, err)
			continue
		}
		t.Log(`Scan ok. id: `, id)
	}

	t.Log(`ok`)
}

func TestLouvTxRows(t *testing.T) {

	tx := DB.Begin()
	//tx := DB
	if tx.Error != nil {
		t.Fatalf(`tx begin err: %v`, tx.GetErrors())
	}

	rows, err := tx.Table(`users`).
		//Debug().
		Where(`id < 80`).
		Select([]string{`id`}).
		Limit(1).
		Rows()

	if err != nil {
		t.Error(`Rows err: `, err)
		return
	}

	// noinspection GoUnhandledErrorResult
	defer rows.Close()

	var scanErrAmt uint32

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			scanErrAmt++
			t.Errorf(`Scan err amt: %d, is: %s`, scanErrAmt, err)
			continue
		}

		t.Log(`Scan ok. id: `, id)
		break
	}

	//_ = rows.Close()

	err = tx.Commit().Error
	if err != nil {
		t.Error(`Commit err: `, err)
		return
	}

	t.Log(`ok`)
}
