package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"zeroDemoProjectForUrl/Torosaurus/server/db"
	"zeroDemoProjectForUrl/Torosaurus/server/util"
)

const (
	pwd_salt = "*/-*-+7454"
)

// SignupHandler GET 请求返回注册页面，POST 请求处理注册
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// bytes, err := ioutil.ReadFile("./static/view/signup.html")
		// if err != nil {
		// 	// 出错就展示错误语句给前端
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	io.WriteString(w,
		// 		fmt.Sprintf("internel server error: %s", err.Error()))
		// 	return
		// }
		// // 正常情况返回 html 页面
		// w.Header().Add("Content-Type", "text/html")
		// io.WriteString(w, string(bytes))
		http.Redirect(w, r, "/static/view/signup.html", http.StatusFound)
		return
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internel server error -> parse Form: %s", err.Error()))
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")
		phone := r.Form.Get("phone")
		if util.IsAllBlank(username, password, phone) {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internel server error -> username, password, phone is not empty")
			return
		}

		// 加密密码
		enc_password := util.Sha1([]byte(password + pwd_salt))

		log.Printf("SignupHandler -> username: %q, password: %q, enc_password: %q\n",
			username, password, enc_password)

		suc := db.UserSignUp(username, phone, enc_password)
		if !suc {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internel server error -> user signup failed")
			return
		}
		w.Header().Add("Content-Type", "application/json;charset=utf-8")
		w.Write(util.GenSimpleRespStream(0, "OK"))
	}

}

// SignInHandler GET 请求返回登录页面，POST 请求处理登录
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internel server error -> parse Form: %s", err.Error()))
			return
		}

		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if util.IsAllBlank(username, password) {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internel server error -> username or password is not empty")
			return
		}

		// 加密密码
		enc_password := util.Sha1([]byte(password + pwd_salt))

		log.Printf("SignInHandler -> username: %q, password: %q, enc_password: %q\n",
			username, password, enc_password)

		pwdCheckRs := db.UserSignIn(username, enc_password)
		if !pwdCheckRs {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internel server error -> username not exist or password not match")
			return
		}

		token := GenToken(username)
		upRes := db.UpdateToken(username, token)
		if !upRes {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internel server error -> update token failed")
			return
		}

		// w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
		resp := util.RespMsg{
			Code: 0,
			Msg:  "OK",
			Data: struct {
				Location string
				Username string
				Token    string
			}{
				Location: "http://" + r.Host + "/static/view/home.html",
				Username: username,
				Token:    token,
			},
		}
		w.Header().Add("Content-Type", "application/json;charset=utf-8")
		w.Write(resp.JSONBytes())
	}

}

const token_salt = "_tokensalt"

func GenToken(username string) string {
	// 40位字符: md5(username + timestamp + token_salt) + timestamp[:8]
	timestamp := strconv.Itoa(int(time.Now().UnixNano() / 1e6))
	tokenPrefix := util.MD5([]byte(username + timestamp + token_salt))
	return tokenPrefix + timestamp[:8]
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internel server error -> parse Form: %s", err.Error()))
		return
	}

	username := r.Form.Get("username")
	if util.IsAllBlank(username) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internel server error -> username is not empty")
		return
	}

	user, err := db.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internel server error -> by username not get exist userinfo")
		return
	}

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(resp.JSONBytes())
}

// IsTokenValid Token是否过期，过期返回true
func IsTokenValid(token string) bool {
	return len(token) != 40
}
