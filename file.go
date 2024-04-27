package yiigo

import (
	"os"
	"path"
	"path/filepath"
)

// CreateFile 创建或清空指定的文件
// 文件已存在，则清空；文件或目录不存在，则以0775权限创建
func CreateFile(filename string) (*os.File, error) {
	abspath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(path.Dir(abspath), 0o775); err != nil {
		return nil, err
	}
	return os.OpenFile(abspath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o775)
}

// OpenFile 打开指定的文件
// 文件已存在，则追加内容；文件或目录不存在，则以0775权限创建
func OpenFile(filename string) (*os.File, error) {
	abspath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(path.Dir(abspath), 0o775); err != nil {
		return nil, err
	}
	return os.OpenFile(abspath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o775)
}
