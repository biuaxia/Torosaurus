package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"zeroDemoProjectForUrl/Torosaurus/server/db"
	"zeroDemoProjectForUrl/Torosaurus/server/meta"
	"zeroDemoProjectForUrl/Torosaurus/server/util"
)

// UploadHandler GET 请求返回上传页面，POST 请求处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回上传页面
		bytes, err := ioutil.ReadFile("./static/view/fileupload.html")
		if err != nil {
			// 出错就展示错误语句给前端
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internal server error: %s", err.Error()))
			return
		}
		// 正常情况返回 html 页面
		io.WriteString(w, string(bytes))
	} else if r.Method == "POST" {
		// 接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internal server error -> get upload file data: %s", err.Error()))
			return
		}
		defer file.Close()

		location := "./tmp/" + head.Filename

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: location,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		// 在服务器本地创建文件
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internal server error -> upload file to server local: %s", err.Error()))
			return
		}
		defer newFile.Close()

		// 拷贝文件到服务器本地
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internal server error -> copy file to server local: %s", err.Error()))
			return
		}

		// Seek 移动到 0 的位置，类似于重置
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		// meta.UpdateFileMeta(&fileMeta)
		ofuf := meta.UpdateFileMetaDB(&fileMeta)

		err = r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,
				fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
			return
		}

		username := r.Form.Get("username")
		if util.IsAllBlank(username) {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internal server error -> username is not empty")
			return
		}

		oufuf := db.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)

		log.Printf("拷贝文件到本地目录 [%s] 完成，大小为 [%d]，更新文件数据库 [%t]，更新文件关联数据库 [%t]",
			newFile.Name(),
			fileMeta.FileSize,
			ofuf,
			oufuf)
		if !ofuf || !oufuf {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internal server error -> 处理文件出错")
			return
		}
		w.Header().Add("Content-Type", "application/json;charset=utf-8")
		w.Write(util.GenSimpleRespStream(0, "OK"))
	}
}

// UploadSucHandler 上传成功的提示信息回写
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "success")
}

// GetAllFileMetaHandler 获取所有文件元信息
func GetAllFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	result := meta.GetAllFileMetaDB()
	bytes, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> marshal json: %s", err.Error()))
		return
	}
	w.Write(bytes)
}

// GetFileMetaHandler 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	filehash := r.Form.Get("filehash")
	fm := meta.GetFileMetaDB(filehash)
	if fm == nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> file not exist")
		return
	}

	bytes, err := json.Marshal(fm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> marshal json: %s", err.Error()))
		return
	}
	w.Write(bytes)
}

// FileQueryHandler 批量查询文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	limit := r.Form.Get("limit")
	if len(limit) == 0 {
		// 设置默认5条数据
		limit = "5"
	}
	limitCount, err := strconv.Atoi(limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> limit Form: %s", err.Error()))
		return
	}

	username := r.Form.Get("username")
	if util.IsBlank(username) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> username is not empty")
		return
	}

	ufs, err := db.QueryUserFileMetas(username, limitCount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> query userfiles failed: %s", err.Error()))
		return
	}

	// fmArray := meta.GetLatestFileMetas(int64(limitCount))
	bytes, err := json.Marshal(ufs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> marshal json: %s", err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(bytes)
}

// DownloadHandler 根据 sha1 下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	filehash := r.Form.Get("filehash")
	if len(filehash) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> filehash isEmpty")
		return
	}

	fm := meta.GetFileMetaDB(filehash)
	if fm == nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> file not exist")
		return
	}

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> load file to memory: %s", err.Error()))
		return
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> io readall: %s", err.Error()))
		return
	}

	w.Header().Add("Content-Disposition",
		fmt.Sprintf("attachment;filename=%s", url.QueryEscape(fm.FileName)))
	w.Header().Add("Content-Type", "application/octect-stream")
	w.Header().Add("Content-Length", strconv.FormatInt(fm.FileSize, 10))
	w.Write(bytes)
}

// UpdateFileMetaHandler 根据 sha1 更新文件名称
func UpdateFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	filehash := r.Form.Get("filehash")
	if len(filehash) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> filehash isEmpty")
		return
	}

	filename := strings.TrimSpace(r.Form.Get("filename"))
	if len(filename) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> filename isEmpty")
		return
	}

	fm := meta.GetFileMetaDB(filehash)
	if fm == nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> file not exist")
		return
	}

	oldFileName := fm.FileName

	// 后缀
	filesuffix := path.Ext(fm.FileName)

	fm.FileName = filename + filesuffix
	b := meta.UpdateFileMetaDB(fm)

	if b {
		log.Printf("更新文件名称完成；原名称 [%s]，新名称 [%s]",
			oldFileName,
			fm.FileName)
	}

	bytes, err := json.Marshal(fm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> marshal json: %s", err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(bytes)
}

// FileDelHandler 删除文件
func FileDelHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	filehash := r.Form.Get("filehash")
	if len(filehash) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> filehash isEmpty")
		return
	}

	fm := meta.GetFileMetaDB(filehash)
	if fm == nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> file not exist")
		return
	}

	// 移除逻辑存储
	ok := meta.RemoveFileMeta(filehash)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> remove file failed")
		return
	}
	// 移除物理存储
	err = os.Remove(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> os remove file failed")
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "success")
}

func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> parse Form: %s", err.Error()))
		return
	}

	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize := r.Form.Get("filesize")
	if util.IsAllBlank(username, filehash, filename, filesize) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error -> username, filehash, filename, filesize is not empty")
		return
	}

	fm := meta.GetFileMetaDB(filehash)
	if fm == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp.JSONBytes())
		return
	}
	fz, err := strconv.Atoi(filesize)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w,
			fmt.Sprintf("internal server error -> filesize convert to int type failed: %s", err.Error()))
		return
	}

	ret := db.OnUserFileUploadFinished(username, filehash, filename, int64(fz))
	var resp util.RespMsg
	if !ret {
		resp = util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		w.WriteHeader(http.StatusBadRequest)
	} else {
		resp = util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.WriteHeader(http.StatusOK)
	}
	w.Write(resp.JSONBytes())
	return

}
