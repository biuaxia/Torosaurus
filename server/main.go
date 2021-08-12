package main

import (
	"log"
	"net/http"
	"zeroDemoProjectForUrl/Torosaurus/server/handler"
)

func main() {
	// 设置路由规则
	// file 文件上传
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/metas", handler.GetAllFileMetaHandler)
	http.HandleFunc("/file/query", handler.FileQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.UpdateFileMetaHandler)
	http.HandleFunc("/file/delete", handler.FileDelHandler)
	// user 用户
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)

	// 开启监听
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("failed to start server, err: %s", err.Error())
	}
}
