package apierr

import "time"

// Response 统一 API 响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// PackageInfo 包信息
type PackageInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version,omitempty"`
	Description string    `json:"description,omitempty"`
	Private     bool      `json:"private"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role"`
	LastLogin time.Time `json:"lastLogin,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// Stats 统计信息
type Stats struct {
	TotalPackages  int   `json:"totalPackages"`
	CachedPackages int   `json:"cachedPackages"`
	StorageSize    int64 `json:"storageSize"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Version       string       `json:"version"`
	StartTime     time.Time    `json:"startTime"`
	Uptime        string       `json:"uptime"`
	Host          string       `json:"host"`
	StoragePath   string       `json:"storagePath"`
	DatabasePath  string       `json:"databasePath"`
	Upstreams     []Upstream   `json:"upstreams"`
}

// Upstream 上游配置
type Upstream struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Scope   string `json:"scope"`
	Timeout int    `json:"timeout"`
	Enabled bool   `json:"enabled"`
}

// WebhookInfo Webhook 信息
type WebhookInfo struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	Events          string    `json:"events"`
	Enabled         bool      `json:"enabled"`
	LastDeliveryAt  time.Time `json:"lastDeliveryAt,omitempty"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID        int       `json:"id"`
	Action    string    `json:"action"`
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}
