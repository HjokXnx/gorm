package gorm

import (
	"errors"
	"fmt"
	"reflect"
)

// Define callbacks for querying
func init() {
	// 查询回调。执行数据库操作，并将查询结果进行映射
	DefaultCallback.Query().Register("gorm:query", queryCallback)
	// 预加载回调。关联模型的预加载，此处不表
	DefaultCallback.Query().Register("gorm:preload", preloadCallback)
	// 查询后回调。对 struct 或 struct 切片，调用钩子事件（如果已定义）
	DefaultCallback.Query().Register("gorm:after_query", afterQueryCallback)
}

// queryCallback 从数据库查询数据
func queryCallback(scope *Scope) {
	// 跳过查询。在关联模型对多对中使用
	if _, skip := scope.InstanceGet("gorm:skip_query_callback"); skip {
		return
	}

	// 只加载预关联模型，不进行实际操作。在调用 DB.Preload() 方法时触发
	if _, skip := scope.InstanceGet("gorm:only_preload"); skip {
		return
	}

	// 调用栈追踪函数（基于实际 defer 调用的时间点）
	defer scope.trace(NowFunc())

	var (
		isSlice, isPtr bool
		resultType     reflect.Type
		results        = scope.IndirectValue() // 返回值，例如 DB.First(&user) 中的 user struct
	)

	// 增加主键排序，例如： ORDER BY `users`.`id` ASC
	if orderBy, ok := scope.Get("gorm:order_by_primary_key"); ok {
		if primaryField := scope.PrimaryField(); primaryField != nil {
			scope.Search.Order(fmt.Sprintf("%v.%v %v", scope.QuotedTableName(), scope.Quote(primaryField.DBName), orderBy))
		}
	}

	// 修改 results 为该配置中的值。在 DB.Scan() 方法中用到
	if value, ok := scope.Get("gorm:query_destination"); ok {
		results = indirect(reflect.ValueOf(value))
	}

	// 反射解析 results 对象
	if kind := results.Kind(); kind == reflect.Slice {
		isSlice = true
		resultType = results.Type().Elem()
		results.Set(reflect.MakeSlice(results.Type(), 0, 0))

		if resultType.Kind() == reflect.Ptr {
			isPtr = true
			resultType = resultType.Elem()
		}
	} else if kind != reflect.Struct {
		scope.Err(errors.New("unsupported destination, should be slice or struct"))
		return
	}

	// 组装 SQL ，即组合 DB.search 里存储的 JOIN / WHERE / GROUP / HAVING / ORDER / LIMIT / OFFSET 等条件。
	// 如果是原生（RAW）SQL，则直接执行
	scope.prepareQuerySQL()

	if !scope.HasError() {
		scope.db.RowsAffected = 0

		// SQL 提示。在原 SQL 前添加字符串
		if str, ok := scope.Get("gorm:query_hint"); ok {
			scope.SQL = fmt.Sprint(str) + scope.SQL
		}

		// SQL 选项。在原 SQL 后添加字符串
		if str, ok := scope.Get("gorm:query_option"); ok {
			scope.SQL += addExtraSpaceIfExist(fmt.Sprint(str))
		}

		// 调用底层的数据库对象进行查询操作，即 interface.SQLCommon 的实现
		if rows, err := scope.SQLDB().Query(scope.SQL, scope.SQLVars...); scope.Err(err) == nil {
			defer rows.Close()

			// 获取查询结果里的行名，例如 []string{`id`, `age`}
			columns, _ := rows.Columns()

			// rows 是 database/sql.Rows 迭代器。
			// GORM 也是通过此迭代器进行处理的——这点和我们直接在外层调用 DB.Rows 方法是一致的
			for rows.Next() {
				scope.db.RowsAffected++

				elem := results
				if isSlice {
					elem = reflect.New(resultType).Elem()
				}

				// 通过大量反射操作，将查询到的数据扫描到 results 对象上
				scope.scan(rows, columns,
					scope.New(elem.Addr().Interface()).Fields(),// 解析传入值 results 字段
				)

				// 如果是切片，对每一个元素
				if isSlice {
					if isPtr {
						results.Set(reflect.Append(results, elem.Addr()))
					} else {
						results.Set(reflect.Append(results, elem))
					}
				}
			}

			// 错误处理
			if err := rows.Err(); err != nil {
				scope.Err(err)
			} else if scope.db.RowsAffected == 0 && !isSlice {
				scope.Err(ErrRecordNotFound)
			}
		}
	}
}

// afterQueryCallback 查询后回调会在查询之后唤醒 `AfterFind` 方法
func afterQueryCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("AfterFind")
	}
}
