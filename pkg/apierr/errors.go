package apierr

import (
	"github.com/gin-gonic/gin"
)

// APIError 统一错误结构
type APIError struct {
	Code    int    `json:"code"`              // 业务错误码
	Message string `json:"message"`           // 用户可读错误
	Reason  string `json:"reason,omitempty"`  // npm 协议兼容字段
}

// 预定义错误码
var (
	ErrBadRequest       = &APIError{Code: 4000, Message: "bad request"}
	ErrUnauthorized     = &APIError{Code: 4010, Message: "authentication required"}
	ErrForbidden        = &APIError{Code: 4030, Message: "insufficient permissions"}
	ErrNotFound         = &APIError{Code: 4040, Message: "resource not found"}
	ErrPackageNotFound  = &APIError{Code: 4041, Message: "package not found"}
	ErrUserNotFound     = &APIError{Code: 4042, Message: "user not found"}
	ErrConflict         = &APIError{Code: 4090, Message: "resource conflict"}
	ErrVersionExists    = &APIError{Code: 4091, Message: "version already exists"}
	ErrUserExists       = &APIError{Code: 4092, Message: "user already exists"}
	ErrRateLimited      = &APIError{Code: 4290, Message: "too many requests"}
	ErrInternal         = &APIError{Code: 5000, Message: "internal server error"}
	ErrInvalidRequest   = &APIError{Code: 4001, Message: "invalid request body"}
	ErrInvalidCredentials = &APIError{Code: 4011, Message: "invalid credentials"}
)

// Respond 返回错误响应
func Respond(c *gin.Context, httpStatus int, err *APIError) {
	c.JSON(httpStatus, err)
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	return e.Message
}

// WithReason 添加 reason 字段
func (e *APIError) WithReason(reason string) *APIError {
	return &APIError{
		Code:    e.Code,
		Message: e.Message,
		Reason:  reason,
	}
}

// WithMessage 自定义消息
func (e *APIError) WithMessage(msg string) *APIError {
	return &APIError{
		Code:    e.Code,
		Message: msg,
		Reason:  e.Reason,
	}
}
