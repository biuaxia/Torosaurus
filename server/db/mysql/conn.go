package mysql

import (
	"database/sql"
	_ "database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	// "user:password@tcp(127.0.0.1:3306)/dbname"
	db, _ = sql.Open("mysql", "Torosaurus:z5pc5XsjTkmGGJxS@tcp(ddns.biuaxia.cn:3306)/torosaurus?charset=utf8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		log.Fatalln("监听数据库出错:", err.Error())
		os.Exit(1)
	}
	log.Println("连接数据库完成")
}

// DBConn 返回数据库链接对象
func DBConn() *sql.DB {
	return db
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()

	values := make([]interface{}, len(columns))
	valuesPtr := make([]interface{}, len(columns))

	for i := range values {
		valuesPtr[i] = &values[i]
	}

	var record []map[string]interface{}
	for rows.Next() {
		// 将行数据扫描到 values 中
		rows.Scan(valuesPtr...)
		r := make(map[string]interface{})
		for i, v := range values {
			if v != nil {
				var fv string
				switch s := v.(type) {
				case int64:
					fv = fmt.Sprintf("%d", s)
				case []uint8:
					fv = string(s)
				default:
					fv = s.(string)
				}
				r[columns[i]] = fv
			}
		}
		record = append(record, r)
	}

	return record
}
