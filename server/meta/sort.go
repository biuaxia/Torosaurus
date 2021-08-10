package meta

import (
	"time"
)

const baseFormat = "2006-01-02 15:04:05"

type ByUploadTime []FileMeta

// Len 返回列表的长度
func (a ByUploadTime) Len() int {
	return len(a)
}

// Swap 支持某两个元素交换
func (a ByUploadTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less 比较规则，此处时间越往后则排序到最下面，即时间降序
func (a ByUploadTime) Less(i, j int) bool {
	iTime, _ := time.Parse(baseFormat, a[i].UploadAt)
	jTime, _ := time.Parse(baseFormat, a[j].UploadAt)
	return iTime.UnixNano() > jTime.UnixNano()
}
