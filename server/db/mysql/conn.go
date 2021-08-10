package mysql

import (
	"database/sql"
	_ "database/sql"
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
