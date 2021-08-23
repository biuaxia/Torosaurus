package db

import (
	"log"
	"time"
	"zeroDemoProjectForUrl/Torosaurus/server/db/mysql"
)

// UserFile 用户文件表结构体
type UserFile struct {
	UserName    string
	FileSha1    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFileUploadFinished 更新用户文件表
func OnUserFileUploadFinished(username string,
	filesha1 string,
	filename string,
	filesize int64) bool {
	sqlTemp := "INSERT INTO `tbl_user_file` (`user_name`, `file_sha1`, `file_name`, `file_size`, `upload_at`) values (?, ?, ?, ?, ?)"
	stmt, err := mysql.DBConn().Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, filesha1, filename, filesize, time.Now())
	if err != nil {
		log.Printf("OnUserFileUploadFinished -> Failed to exec sql, err: %s\n", err.Error())
		return false
	}

	rf, err := ret.RowsAffected()
	if err != nil {
		log.Printf("Failed to check RowsAffected sql, err: %s\n", err.Error())
		return false
	}
	if rf <= 0 {
		log.Printf("sql exec success, but affect row is zero.")
	}
	return true
}

// QueryUserFileMetas 批量获取文件信息
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	sqlTemp := "SELECT `file_sha1`, `file_name`, `file_size`, `upload_at`, `last_update` FROM `tbl_user_file` WHERE `user_name` = ? limit ?"
	stmt, err := mysql.DBConn().Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query(username, limit)
	if nil != err {
		log.Printf("Failed to queryRow, err: %s\n", err.Error())
		return nil, err
	}
	defer rs.Close()

	var userFiles []UserFile

	for rs.Next() {
		uf := UserFile{}

		err := rs.Scan(&uf.FileSha1, &uf.FileName, &uf.FileSize, &uf.UploadAt, &uf.LastUpdated)
		if err != nil {
			log.Printf("Failed to queryRow, err: %s\n", err.Error())
			break
		}
		userFiles = append(userFiles, uf)
	}
	return userFiles, nil
}
