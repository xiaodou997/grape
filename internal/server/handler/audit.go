package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/db"
)

// GetAuditLogs 获取审计日志
// GET /-/api/admin/audit-logs
func GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var logs []db.AuditLog
	var total int64

	db.DB.Model(&db.AuditLog{}).Count(&total)
	db.DB.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
