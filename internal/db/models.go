package db

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User 用户模型
type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email     string     `gorm:"size:255" json:"email"`
	Password  string     `gorm:"size:255;not null" json:"-"`
	Role      string     `gorm:"size:50;default:developer" json:"role"`
	CreatedAt time.Time  `json:"createdAt"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// SetPassword 设置密码（自动哈希）
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Package 包元数据模型（用于快速查询）
type Package struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:255;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Private     bool      `gorm:"default:false" json:"private"`
	Latest      string    `gorm:"size:50" json:"latest"` // 最新版本
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (Package) TableName() string {
	return "packages"
}

// PackageVersion 包版本模型
type PackageVersion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PackageName string    `gorm:"uniqueIndex:idx_pkg_version;size:255;not null" json:"packageName"`
	Version     string    `gorm:"uniqueIndex:idx_pkg_version;size:50;not null" json:"version"`
	Tarball     string    `gorm:"size:500" json:"tarball"` // tarball 文件名
	Shasum      string    `gorm:"size:100" json:"shasum"`
	Publisher   string    `gorm:"size:100" json:"publisher"` // 发布者
	CreatedAt   time.Time `json:"createdAt"`
}

// TableName 指定表名
func (PackageVersion) TableName() string {
	return "package_versions"
}

// AuditLog 审计日志
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Action    string    `gorm:"size:50;index" json:"action"`  // login/publish/unpublish/user_create/user_delete/config_update
	Username  string    `gorm:"size:100;index" json:"username"`
	IP        string    `gorm:"size:50" json:"ip"`
	Detail    string    `gorm:"size:500" json:"detail"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Webhook 事件 Webhook 配置
type Webhook struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"size:100;not null" json:"name"`
	URL            string     `gorm:"size:500;not null" json:"url"`
	Secret         string     `gorm:"size:255" json:"-"` // HMAC 密钥，不对外暴露
	Events         string     `gorm:"size:500" json:"events"` // 逗号分隔的事件类型，空表示订阅所有
	Enabled        bool       `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time  `json:"createdAt"`
	LastDeliveryAt *time.Time `json:"lastDeliveryAt,omitempty"`
}

// TableName 指定表名
func (Webhook) TableName() string {
	return "webhooks"
}

// Token CI/CD 持久化 Token（用于自动化发布）
type Token struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"index;not null" json:"userId"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name      string     `gorm:"size:100;not null" json:"name"`              // 描述，如 "github-ci"
	TokenHash string     `gorm:"size:64;uniqueIndex;not null" json:"-"`      // sha256 哈希存储
	Readonly  bool       `gorm:"default:false" json:"readonly"`              // 只读 token 不能发布
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`                        // nil = 永不过期
	LastUsed  *time.Time `json:"lastUsed,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// TableName 指定表名
func (Token) TableName() string {
	return "tokens"
}

// PackageOwner 包的 Owner（访问控制）
type PackageOwner struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PackageName string    `gorm:"uniqueIndex:idx_pkg_owner;size:255;not null" json:"packageName"` // 支持通配符如 @ui/*
	UserID      uint      `gorm:"uniqueIndex:idx_pkg_owner;not null" json:"userId"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CanPublish  bool      `gorm:"default:true" json:"canPublish"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TableName 指定表名
func (PackageOwner) TableName() string {
	return "package_owners"
}
