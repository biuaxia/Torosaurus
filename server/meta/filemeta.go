package meta

import (
	"sort"
	"strings"
	"zeroDemoProjectForUrl/Torosaurus/server/db"
)

// FileMeta 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	// Location 本地路径
	Location string
	// UploadAt 文件上传时间戳
	UploadAt string
}

var fileMetas map[string]*FileMeta

func init() {
	fileMetas = make(map[string]*FileMeta)
}

// UpdateFileMeta 新增/更新文件元信息
func UpdateFileMeta(fm *FileMeta) {
	fileMetas[strings.ToLower(fm.FileSha1)] = fm
}

// UpdateFileMetaDB 新增/更新文件元信息持久化到数据库，操作数据库成功返回true
func UpdateFileMetaDB(fm *FileMeta) bool {
	return db.OnFileUploadFinished(
		fm.FileSha1,
		fm.FileName,
		fm.FileSize,
		fm.Location,
	)
}

// GetFileMeta 通过 fs(FileSha1) 获取文件元信息
func GetFileMeta(fs string) *FileMeta {
	if v, ok := fileMetas[strings.ToLower(fs)]; ok {
		return v
	}
	return nil
}

// GetFileMeta 通过 fs(FileSha1) 从 DB 获取文件元信息
func GetFileMetaDB(fs string) *FileMeta {
	tf, err := db.OnGetFileMeta(fs)
	if err != nil {
		return nil
	}

	return &FileMeta{
		FileSha1: tf.FileSha1,
		FileName: tf.FileName.String,
		FileSize: tf.FileSize.Int64,
		Location: tf.FileAddr.String,
	}
}

// GetFileMeta 通过 fs(FileSha1) 获取文件元信息
func GetAllFileMeta() []FileMeta {
	result := make([]FileMeta, 0)
	// 取出map中的所有key存入切片keys
	var keys = make([]string, 0)
	for key := range fileMetas {
		keys = append(keys, key)
	}
	//对切片进行排序
	sort.Strings(keys)

	//按照排序后的key遍历map
	for _, key := range keys {
		result = append(result, *fileMetas[key])
	}
	return result
}

// GetFileMeta 通过 fs(FileSha1) 获取文件元信息
func GetAllFileMetaDB() []FileMeta {
	result := make([]FileMeta, 0)
	tfs := db.OnGetAllFileMetas(99)

	//按照排序后的key遍历map
	for _, tf := range tfs {
		fm := FileMeta{
			FileSha1: tf.FileSha1,
			FileName: tf.FileName.String,
			FileSize: tf.FileSize.Int64,
			Location: tf.FileAddr.String,
		}
		result = append(result, fm)
	}
	return result
}

// GetLatestFileMetas 获取指定条数的最新文件列表
func GetLatestFileMetas(count int64) []FileMeta {
	result := make([]FileMeta, 0)
	tfs := db.OnGetAllFileMetas(count)
	//按照排序后的key遍历map
	for _, tf := range tfs {
		fm := FileMeta{
			FileSha1: tf.FileSha1,
			FileName: tf.FileName.String,
			FileSize: tf.FileSize.Int64,
			Location: tf.FileAddr.String,
		}
		result = append(result, fm)
	}

	return result
}

// RemoveFileMeta 从Map中移除指定key的元素
func RemoveFileMeta(fs string) bool {
	fs = strings.ToLower(fs)
	delete(fileMetas, fs)
	_, ok := fileMetas[fs]
	return !ok
}
