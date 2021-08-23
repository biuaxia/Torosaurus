package db

import (
	"database/sql"
	"log"
	"zeroDemoProjectForUrl/Torosaurus/server/db/mysql"
	"zeroDemoProjectForUrl/Torosaurus/server/util"
)

// OnFileUploadFinished 文件上传完成时的数据库操作
func OnFileUploadFinished(
	filehash string,
	filename string,
	filesize int64,
	fileaddr string,
) bool {
	db := mysql.DBConn()

	tf, err := OnGetFileMeta(filehash)
	if nil != err || tf == nil {
		// 新数据
		sqlTemp := "INSERT INTO `tbl_file` (`file_sha1`, `file_name`, `file_size`, `file_addr`, `status`) VALUES (?, ?, ?, ?, 1)"
		stmt, err := db.Prepare(sqlTemp)
		if err != nil {
			log.Printf("Failed to prepare statement, err: %s\n", err.Error())
			return false
		}
		defer stmt.Close()

		ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
		if err != nil {
			log.Printf("OnFileUploadFinished | 1 -> Failed to exec sql, err: %s\n", err.Error())
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
	if tf.FileName.String == filename {
		log.Println("文件名称未发生变化，修改失败")
		return false
	}

	// 老数据
	sqlTemp := "UPDATE `tbl_file` SET `file_name` = ? WHERE `file_sha1` = ?"
	stmt, err := db.Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filename, filehash)
	if err != nil {
		log.Printf("OnFileUploadFinished | 2 -> Failed to exec sql, err: %s\n", err.Error())
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

type TableFile struct {
	FileSha1 string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// OnGetFileMeta 获取文件元信息
func OnGetFileMeta(filehash string) (*TableFile, error) {
	db := mysql.DBConn()

	sqlTemp := "SELECT `file_sha1`, `file_name`, `file_size`, `file_addr`" +
		" FROM `tbl_file` where `file_sha1` = ? and status = 1 limit 1"

	stmt, err := db.Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return nil, err
	}
	defer stmt.Close()

	tf := TableFile{}

	err = stmt.QueryRow(filehash).
		// 扫描到对象字段（与SQL语句一致），指针传递，外部就可以直接使用 tf 对象了
		Scan(&tf.FileSha1, &tf.FileName, &tf.FileSize, &tf.FileAddr)
	if err != nil {
		log.Printf("Failed to queryRow, err: %s\n", err.Error())
		return nil, err
	}

	return &tf, nil
}

func OnGetAllFileMetasByUsername(limit int64, username string) []TableFile {
	db := mysql.DBConn()

	var rs *sql.Rows

	if util.IsBlank(username) {
		sqlTemp := "SELECT `file_sha1`, `file_name`, `file_size`, `file_addr`" +
			" FROM `tbl_file` where status = 1 order by `create_at` limit ? "
		stmt, err := db.Prepare(sqlTemp)
		if err != nil {
			log.Printf("Failed to prepare statement, err: %s\n", err.Error())
			return nil
		}
		defer stmt.Close()

		if limit == 0 {
			limit = 5
		}

		rs, err := stmt.Query(limit)
		if nil != err {
			log.Printf("Failed to queryRow, err: %s\n", err.Error())
			return nil
		}
		defer rs.Close()
	} else {
		sqlTemp := "SELECT `file_sha1`, `file_name`, `file_size`, `file_addr`" +
			" FROM `tbl_file` where status = 1 order by `create_at` limit ? "
		stmt, err := db.Prepare(sqlTemp)
		if err != nil {
			log.Printf("Failed to prepare statement, err: %s\n", err.Error())
			return nil
		}
		defer stmt.Close()

		if limit == 0 {
			limit = 5
		}

		rs, err := stmt.Query(limit)
		if nil != err {
			log.Printf("Failed to queryRow, err: %s\n", err.Error())
			return nil
		}
		defer rs.Close()
	}

	var tfs []TableFile

	for rs.Next() {
		var tf TableFile
		err := rs.Scan(&tf.FileSha1, &tf.FileName, &tf.FileSize, &tf.FileAddr)
		if err != nil {
			log.Printf("Failed to queryRow, err: %s\n", err.Error())
			return nil
		}
		tfs = append(tfs, tf)
	}
	return tfs
}

// OnGetAllFileMetas 获取所有文件元信息，支持分页，默认按照上传时间排序
func OnGetAllFileMetas(limit int64) []TableFile {
	db := mysql.DBConn()

	sqlTemp := "SELECT `file_sha1`, `file_name`, `file_size`, `file_addr`" +
		" FROM `tbl_file` where status = 1 order by `create_at` limit ? "
	stmt, err := db.Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return nil
	}
	defer stmt.Close()

	if limit == 0 {
		limit = 5
	}

	rs, err := stmt.Query(limit)
	if nil != err {
		log.Printf("Failed to queryRow, err: %s\n", err.Error())
		return nil
	}
	defer rs.Close()

	var tfs []TableFile

	for rs.Next() {
		var tf TableFile
		err := rs.Scan(&tf.FileSha1, &tf.FileName, &tf.FileSize, &tf.FileAddr)
		if err != nil {
			log.Printf("Failed to queryRow, err: %s\n", err.Error())
			return nil
		}
		tfs = append(tfs, tf)
	}
	return tfs
}
