package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"zeroDemoProjectForUrl/Torosaurus/server/util"
)

// HttpInterceptor http请求拦截器
func HttpInterceptor(hf http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Println("请求拦截...")
			err := r.ParseForm()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w,
					fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
				return
			}

			username := r.Form.Get("username")
			token := r.Form.Get("token")
			if util.IsAllBlank(username, token) {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "internal server error -> username or token is not empty")
				return
			}

			// 验证登录token是否有效
			if len(username) < 3 || IsTokenValid(token) {
				// w.WriteHeader(http.StatusForbidden)
				// token校验失败则跳转到登录页面
				http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
				return
			}

			hf(w, r)
		},
	)
}
