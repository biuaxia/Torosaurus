package handler

import (
	"fmt"
	"io"
	"io/ioutil"
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
		bytes, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			// 出错就展示错误语句给前端
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internel server error: %s", err.Error()))
			return
		}
		// 正常情况返回 html 页面
		w.Header().Add("Content-Type", "text/html")
		io.WriteString(w, string(bytes))
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

		log.Printf("SignupHandler -> username: %q, password: %q, enc_password: %q\n",
			username, password, enc_password)

		suc := db.UserSignUp(username, enc_password)
		if suc {
			io.WriteString(w, "success")
			return
		}
		io.WriteString(w, "failed user login")
		return
	}

}

// SignInHandler GET 请求返回登录页面，POST 请求处理登录
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		bytes, err := ioutil.ReadFile("./static/view/signin.html")
		if err != nil {
			// 出错就展示错误语句给前端
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internel server error: %s", err.Error()))
			return
		}
		// 正常情况返回 html 页面
		w.Header().Add("Content-Type", "text/html")
		io.WriteString(w, string(bytes))
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

		w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	}

}

const token_salt = "_tokensalt"

func GenToken(username string) string {
	// 40位字符: md5(username + timestamp + token_salt) + timestamp[:8]
	timestamp := strconv.Itoa(int(time.Now().UnixNano() / 1e6))
	tokenPrefix := util.MD5([]byte(username + timestamp + token_salt))
	return tokenPrefix + timestamp[:8]
}
