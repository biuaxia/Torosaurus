package db

import (
	"log"
	"zeroDemoProjectForUrl/Torosaurus/server/db/mysql"
)

// OnFileUploadFinished 文件上传完成时的数据库操作
func OnFileUploadFinished(
	filehash string,
	filename string,
	filesize int64,
	fileaddr string,
) bool {
	db := mysql.DBConn()
	log.Println("检查数据库是否为空", db == nil)

	sqlTemp := "INSERT INTO `torosaurus`.`tbl_file` (`file_sha1`, `file_name`, `file_size`, `file_addr`, `status`) VALUES (?, ?, ?, ?, 1)"
	stms, err := db.Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return false
	}
	defer stms.Close()

	ret, err := stms.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		log.Printf("Failed to exec sql, err: %s\n", err.Error())
		return false
	}

	rf, err := ret.RowsAffected()
	if err != nil {
		log.Printf("Failed to exec sql, err: %s\n", err.Error())
		return false
	}
	if rf <= 0 {
		log.Printf("sql exec success, but affect row is zero.")
	}
	return true
}
