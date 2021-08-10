package meta

import (
	"sort"
	"strings"
)

// FileMeta 文件元信息结构
type FileMeta struct {
	FileSha1 string `json:"fileSha1,omitempty"`
	FileName string `json:"fileName,omitempty"`
	FileSize int64  `json:"fileSize,omitempty"`
	// Location 本地路径
	Location string `json:"fileLocation,omitempty"`
	// UploadAt 文件上传时间戳
	UploadAt string `json:"fileUploadTime,omitempty"`
}

var fileMetas map[string]*FileMeta

func init() {
	fileMetas = make(map[string]*FileMeta)
}

// UpdateFileMeta 新增/更新文件元信息
func UpdateFileMeta(fm *FileMeta) {
	fileMetas[strings.ToLower(fm.FileSha1)] = fm
}

// GetFileMeta 通过 fs(FileSha1) 获取文件元信息
func GetFileMeta(fs string) *FileMeta {
	if v, ok := fileMetas[strings.ToLower(fs)]; ok {
		return v
	}
	return nil
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

// GetLatestFileMetas 获取指定条数的最新文件列表
func GetLatestFileMetas(count int) []FileMeta {
	fmArray := make([]FileMeta, 0)
	for _, v := range fileMetas {
		fmArray = append(fmArray, *v)
	}
	sort.Sort(ByUploadTime(fmArray))
	// 对数据进行处理，如果想要获取的条数count大于数组长度则以数据长度为准
	if len(fmArray) < count {
		count = len(fmArray)
	}
	return fmArray[:count]
}

// RemoveFileMeta 从Map中移除指定key的元素
func RemoveFileMeta(fs string) bool {
	fs = strings.ToLower(fs)
	delete(fileMetas, fs)
	_, ok := fileMetas[fs]
	return !ok
}
