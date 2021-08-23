package handler

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
	"zeroDemoProjectForUrl/Torosaurus/server/cache/redis"
	"zeroDemoProjectForUrl/Torosaurus/server/util"
)

// MultipartUploadInfo 初始化信息
type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadId string
	// 分块大小
	ChunkSize int
	// 分块数量
	ChunkCount int
}

// InitialMultipartUploadHandler 分块上传初始化
func InitialMultipartUploadHandler(w http.ResponseWriter, r http.Request) {
	// 1. 解析用户请求参数
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internel server error -> parse Form: %s", err.Error()))
		return
	}

	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesizeStr := r.Form.Get("filesize")
	if util.IsAllBlank(username, filehash, filesizeStr) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internel server error -> username, filehash, filesize is not empty")
		return
	}

	filesize, err := strconv.Atoi(filesizeStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internel server error -> filesize convert to int type failed")
		return
	}

	// 2. 获得Redis的链接
	rc := redis.RedisPool().Get()
	defer rc.Close()

	// 3. 生成分块上传的初始化信息
	mu := MultipartUploadInfo{
		FileHash: filehash,
		FileSize: filesize,
		UploadId: username + fmt.Sprintf("%x", time.Now().UnixNano()),
		// 5MB
		ChunkSize: 5 * 1024 * 1024,
		// 分块大小计算
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	// 4. 将初始化信息写入到redis缓存
	rc.Do("HSET", "MP_"+mu.UploadId, "chunkcount", mu.ChunkCount)
	rc.Do("HSET", "MP_"+mu.UploadId, "filehash", mu.FileHash)
	rc.Do("HSET", "MP_"+mu.UploadId, "filesize", mu.FileSize)

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", mu).JSONBytes())
}

// UplaodPartHandler 上传文件分块
func UplaodPartHandler(w http.ResponseWriter, r http.Request) {
	// 1. 解析用户请求参数
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internel server error -> parse Form: %s", err.Error()))
		return
	}

	username := r.Form.Get("username")
	uploadId := r.Form.Get("uploadId")
	chunkIndex := r.Form.Get("chunkIndex")
	if util.IsAllBlank(username, uploadId, chunkIndex) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internel server error -> username, uploadId, chunkIndex is not empty")
		return
	}

	// 2. 获得 redis 连接
	rc := redis.RedisPool().Get()
	defer rc.Close()

	// 3. 获得文件句柄，用于存储分块内容
	f, err := os.Create("./data/" + uploadId + "/" + chunkIndex)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internel server error -> create upload dir failed")
		return
	}
	defer f.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internel server error -> from request read binary stream failed")
			return
		}
		f.Write(buf[:n])
	}

	// 4. 更新 reddis 缓存状态
	rc.Do("HSET", "MP_"+uploadId, "chkidx_"+chunkIndex, 1)

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
