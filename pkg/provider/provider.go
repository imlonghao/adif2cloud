package provider

import (
	"io"
)

// Provider 定义了云存储提供商的接口
type Provider interface {
	// GetSize 获取远程 ADIF 文件的大小
	GetSize() (int64, error)

	// Download 下载远程 ADIF 文件到本地
	Download(w io.Writer) error

	// Upload 上传文件或新增行到远程端
	Upload(filename string, line string) error
}
