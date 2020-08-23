package gorm

import (
	"context"
	"database/sql"
)

// SQLCommon 是 GORM 依赖的底层数据库连接对象的最小化抽象接口。实质是 *sql.DB 或 *sql.Tx 。
type SQLCommon interface {
	Exec(query string, args ...interface{}) (sql.Result, error) // 不返回数据集 Rows，只返回数据结果（最后 ID 和影响行数）
	Prepare(query string) (*sql.Stmt, error)                    // 预处理语句，GORM 实际上未使用
	Query(query string, args ...interface{}) (*sql.Rows, error) // 返回可迭代的数据集 Rows
	QueryRow(query string, args ...interface{}) *sql.Row        // 仅返回一行数据，实际上是调用上面的 Query() 方法
}

type sqlDb interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type sqlTx interface {
	Commit() error
	Rollback() error
}
