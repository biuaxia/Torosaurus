package handler

import (
	"fmt"
	redis2 "github.com/garyburd/redigo/redis"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"zeroDemoProjectForUrl/Torosaurus/server/cache/redis"
	"zeroDemoProjectForUrl/Torosaurus/server/db"
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
	// 已经上传完成的分块索引列表
	ChunkExists []int
}

const (
	// ChunkDir 上传的分块所在目录
	ChunkDir = "./data/chunks/"
	// MergeDir 合并后的文件所在目录
	MergeDir = "./data/merge/"
	// ChunkKeyPrefix 分块信息对应的 Redis 键前缀
	ChunkKeyPrefix = "MP_"
	// HashUpIDKeyPrefix 文件 Hash 映射 uploadId 对应的 Redis 键前缀
	HashUpIDKeyPrefix = "HASH_UPID_"
	// ChunkIndexKeyPrefix 文件分块键前缀
	ChunkIndexKeyPrefix = "chkidx_"
)

// InitialMultipartUploadHandler 分块上传初始化
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesizeStr := r.Form.Get("filesize")
	if util.IsAllBlank(username, filehash, filesizeStr) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> username, filehash, filesize is not empty")
		return
	}

	filesize, err := strconv.Atoi(filesizeStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> filesize convert to int type failed")
		return
	}

	// 2. 获得Redis的链接
	rc := redis.RedisPool().Get()
	defer rc.Close()

	// 2.1
	var uploadId string
	// 通过文件hash判断是否断点续传，并获取 uploadId
	keyExists, _ := redis2.Bool(rc.Do("EXISTS", HashUpIDKeyPrefix+filehash))
	if keyExists {
		uploadId, err = redis2.String(rc.Do("GET", HashUpIDKeyPrefix+filehash))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internal server error -> query uploadId failed")
			return
		}
	}

	var chunkExists []int
	if util.IsBlank(uploadId) {
		// 新上传
		uploadId = username + fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		// 取出所有分块信息
		values, err := redis2.Values(rc.Do("HGETALL", ChunkKeyPrefix+uploadId))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internal server error -> query chunk failed")
			return
		}
		// ChunkExists
		for i := 0; i < len(values); i += 2 {
			k := string(values[i].([]byte))
			v := string(values[i+1].([]byte))
			if strings.HasPrefix(k, HashUpIDKeyPrefix) && v == "1" {
				// 拿到索引下标
				index := k[7:]
				chunkIdx, _ := strconv.Atoi(index)
				chunkExists = append(chunkExists, chunkIdx)
			}
		}
	}

	// 3. 生成分块上传的初始化信息
	mu := MultipartUploadInfo{
		FileHash: filehash,
		FileSize: filesize,
		UploadId: uploadId,
		// 5MB
		ChunkSize: 5 * 1024 * 1024,
		// 分块大小计算
		ChunkCount:  int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
		ChunkExists: chunkExists,
	}

	// 4. 将初始化信息写入到 Redis 缓存
	if len(mu.ChunkExists) <= 0 {
		// 首次上传才会提交到 Redis
		hk := ChunkKeyPrefix + mu.UploadId
		rc.Do("HSET", hk, "chunkcount", mu.ChunkCount)
		rc.Do("HSET", hk, "filehash", mu.FileHash)
		rc.Do("HSET", hk, "filesize", mu.FileSize)
		rc.Do("EXPIRE", hk, 43200)
		rc.Do("SET", HashUpIDKeyPrefix+filehash, mu.UploadId, "EX", 43200)
	}

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", mu).JSONBytes())
}

// UplaodPartHandler 上传文件分块
func UplaodPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	username := r.Form.Get("username")
	uploadId := r.Form.Get("uploadId")
	chunkIndex := r.Form.Get("chunkIndex")
	if util.IsAllBlank(username, uploadId, chunkIndex) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> username, uploadId, chunkIndex is not empty")
		return
	}

	// 2. 获得 Redis 连接
	rc := redis.RedisPool().Get()
	defer rc.Close()

	// 3. 获得文件句柄，用于存储分块内容
	// Tips 先创建文件夹再创建文件
	filepath := ChunkDir + uploadId + "/" + chunkIndex
	err = os.MkdirAll(path.Dir(filepath), 0744)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> create upload dir failed")
		return
	}

	f, err := os.Create(filepath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> create upload file failed")
		return
	}
	defer f.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		f.Write(buf[:n])

		if err != nil {
			break
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internal server error -> from request read binary stream failed")
			return
		}
	}

	// 4. 更新 reddis 缓存状态
	rc.Do("HSET", ChunkKeyPrefix+uploadId, ChunkIndexKeyPrefix+chunkIndex, 1)

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler 上传完成进行合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	username := r.Form.Get("username")
	uploadId := r.Form.Get("uploadId")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")
	if util.IsAllBlank(username, uploadId, filehash, filesize, filename) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> username, uploadId, filehash, filesize, filename is not empty")
		return
	}

	// 2. 获得 Redis 连接
	rc := redis.RedisPool().Get()
	defer rc.Close()

	// 3. 通过 UploadId 查询 Redis 并判断是否所有分块上传完成
	data, err := redis2.Values(rc.Do("HGETALL", ChunkKeyPrefix+uploadId))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> complete upload failed")
		return
	}
	totalCount := 0
	chunkCount := 0

	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))

		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, HashUpIDKeyPrefix) && v == "1" {
			chunkCount++
		}
	}

	fmt.Printf("totalCount: %d, chunkCount: %d\n", totalCount, chunkCount)
	if totalCount != chunkCount {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> invalid request")
		return
	}

	// 4. TODO 合并分块

	// 5. 更新唯一文件表及用户文件表
	fz, err := strconv.Atoi(filesize)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> filesize convert to int type failed")
		return
	}
	db.OnFileUploadFinished(filehash, filename, int64(fz), "")
	db.OnUserFileUploadFinished(username, filehash, filename, int64(fz))

	// 6. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

func CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	filehash := r.Form.Get("filehash")
	if util.IsBlank(filehash) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> filehash is not empty")
		return
	}

	// 2. 获得 Redis 连接
	rc := redis.RedisPool().Get()
	defer rc.Close()

	// 3. 检查 UploadId 是否存在，如果存在则删除
	uploadId, err := redis2.String(rc.Do("GET", HashUpIDKeyPrefix+filehash))
	if err != nil || util.IsBlank(uploadId) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> query uploadId failed")
		return
	}

	_, err = rc.Do("DEL", HashUpIDKeyPrefix+filehash)
	if err != nil || util.IsBlank(uploadId) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> by filehash delete hash record failed")
		return
	}

	_, err = rc.Do("DEL", ChunkKeyPrefix+uploadId)
	if err != nil || util.IsBlank(uploadId) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> by filehash delete chunk record failed")
		return
	}

	// 4. 删除已上传的文件分块
	delChunkRes := util.RemovePathByShell(ChunkDir + uploadId)
	if !delChunkRes {
		fmt.Println("[Warning] 删除分块文件夹失败，请稍后手动处理")
	}

	// 6. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
