package server

import (
	"strings"
)

// RegistryRequestType 注册表请求类型
type RegistryRequestType int

const (
	RequestMetadata RegistryRequestType = iota
	RequestTarball
)

// RegistryPathInfo 解析后的路径信息
type RegistryPathInfo struct {
	Type        RegistryRequestType
	PackageName string
	Filename    string // 仅用于 Tarball
}

// parseRegistryPath 将原始 URL 路径解析为结构化的包信息
func parseRegistryPath(path string) *RegistryPathInfo {
	// 去除前导和尾随斜杠
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}

	// 检查是否为 tarball 请求 (路径中包含 "/-/")
	if strings.Contains(path, "/-/") {
		idx := strings.Index(path, "/-/")
		return &RegistryPathInfo{
			Type:        RequestTarball,
			PackageName: path[:idx],
			Filename:    path[idx+3:],
		}
	}

	// 否则视为包元数据请求
	return &RegistryPathInfo{
		Type:        RequestMetadata,
		PackageName: path,
	}
}
