package db

// RecordAudit 记录一条审计日志（异步，不阻塞请求）
func RecordAudit(action, username, ip, detail string) {
	if DB == nil {
		return
	}
	log := &AuditLog{
		Action:   action,
		Username: username,
		IP:       ip,
		Detail:   detail,
	}
	// 使用独立 goroutine 避免阻塞请求
	go func() {
		DB.Create(log)
	}()
}
