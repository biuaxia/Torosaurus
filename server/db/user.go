package db

import (
	"fmt"
	"log"
	"zeroDemoProjectForUrl/Torosaurus/server/db/mysql"
)

// UserSignUp 用户注册，成功返回 true
func UserSignUp(username string, phone string, password string) bool {
	u, _ := GetUserInfo(username)
	if u.Username == username {
		log.Println("User exist, don't insert")
		return false
	}

	sqlTemp := "INSERT INTO `tbl_user` (`user_name`, `phone`, `user_pwd`) VALUES (?, ?, ?)"
	stmt, err := mysql.DBConn().Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, phone, password)
	if err != nil {
		log.Printf("UserSignUp -> Failed to exec sql, err: %s\n", err.Error())
		return false
	}

	rf, err := ret.RowsAffected()
	if err != nil {
		log.Printf("Failed to check RowsAffected sql, err: %s\n", err.Error())
		return false
	}
	if rf <= 0 {
		log.Printf("sql exec success, but affect row is zero.")
		return false
	}
	log.Printf("user signup sql exec success")
	return true
}

// UserSignIn 用户登录，成功返回 true
func UserSignIn(username string, enc_password string) bool {
	sqlTemp := "SELECt * FROM `tbl_user` WHERE `user_name` = ? limit 1"
	stmt, err := mysql.DBConn().Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return false
	}
	defer stmt.Close()

	rs, err := stmt.Query(username)
	if nil != err {
		log.Printf("Failed to queryRow, err: %s\n", err.Error())
		return false
	}
	defer rs.Close()

	pr := mysql.ParseRows(rs)
	if len(pr) > 0 && fmt.Sprint(pr[0]["user_pwd"]) == enc_password {
		return true
	}
	return false
}

func UpdateToken(username string, token string) bool {
	sqlTemp := "replace into `tbl_user_token` (`user_name`, `user_token`) values (?, ?)"
	stmt, err := mysql.DBConn().Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, token)
	if err != nil {
		log.Printf("UpdateToken -> Failed to exec sql, err: %s\n", err.Error())
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

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// GetUserInfo 获取用户信息
func GetUserInfo(username string) (User, error) {
	user := User{}

	sqlTemp := "SELECT `user_name`, `signup_at` FROM `tbl_user` WHERE `user_name` = ? limit 1"
	stmt, err := mysql.DBConn().Prepare(sqlTemp)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		log.Printf("Warning -> exec query, err: %s\n", err.Error())
		return user, err
	}
	return user, nil
}
