package main

import (
	"log"
	"net/http"
	"zeroDemoProjectForUrl/Torosaurus/server/handler"
)

func main() {
	// 设置路由规则
	// 处理静态资源映射
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	// file 文件上传
	http.HandleFunc("/file/upload", handler.HttpInterceptor(handler.UploadHandler))
	http.HandleFunc("/file/upload/suc", handler.HttpInterceptor(handler.UploadSucHandler))
	http.HandleFunc("/file/meta", handler.HttpInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/metas", handler.HttpInterceptor(handler.GetAllFileMetaHandler))
	http.HandleFunc("/file/query", handler.HttpInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/file/download", handler.HttpInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/update", handler.HttpInterceptor(handler.UpdateFileMetaHandler))
	http.HandleFunc("/file/delete", handler.HttpInterceptor(handler.FileDelHandler))
	http.HandleFunc("/file/fastupload", handler.HttpInterceptor(handler.TryFastUploadHandler))

	// 分块上传接口
	http.HandleFunc("/file/mpupload/init", handler.HttpInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart", handler.HttpInterceptor(handler.UplaodPartHandler))
	http.HandleFunc("/file/mpupload/complete", handler.HttpInterceptor(handler.CompleteUploadHandler))

	// user 用户
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HttpInterceptor(handler.UserInfoHandler))

	// 开启监听
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to start server, err: %s", err.Error())
	}
}
