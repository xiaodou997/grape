package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var distFS embed.FS

// GetFileSystem 返回嵌入的文件系统
func GetFileSystem() http.FileSystem {
	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(dist)
}

// GetDistFS 返回 dist 目录的 fs.FS
func GetDistFS() fs.FS {
	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return dist
}

// ReadFile 读取嵌入的文件
func ReadFile(name string) ([]byte, error) {
	return distFS.ReadFile("dist" + name)
}

// Exists 检查文件是否存在
func Exists(name string) bool {
	_, err := distFS.Open("dist" + name)
	return err == nil
}