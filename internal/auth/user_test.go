package auth

import (
	"testing"
)

func TestMemoryUserStore_Create(t *testing.T) {
	store := NewMemoryUserStore()

	// 测试创建新用户
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "developer",
	}

	err := store.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 测试重复创建
	err = store.Create(user)
	if err != ErrUserAlreadyExists {
		t.Fatalf("Expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestMemoryUserStore_Get(t *testing.T) {
	store := NewMemoryUserStore()

	// 测试获取已存在的用户
	user, err := store.Get("admin")
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}
	if user.Username != "admin" {
		t.Fatalf("Expected admin, got: %s", user.Username)
	}

	// 测试获取不存在的用户
	_, err = store.Get("nonexistent")
	if err != ErrUserNotFound {
		t.Fatalf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestMemoryUserStore_Validate(t *testing.T) {
	store := NewMemoryUserStore()

	// 测试正确密码
	user, err := store.Validate("admin", "admin")
	if err != nil {
		t.Fatalf("Failed to validate admin: %v", err)
	}
	if user.Username != "admin" {
		t.Fatalf("Expected admin, got: %s", user.Username)
	}

	// 测试错误密码
	_, err = store.Validate("admin", "wrongpassword")
	if err != ErrInvalidPassword {
		t.Fatalf("Expected ErrInvalidPassword, got: %v", err)
	}
}

func TestMemoryUserStore_Delete(t *testing.T) {
	store := NewMemoryUserStore()

	// 创建测试用户
	user := &User{
		Username: "tobedeleted",
		Password: "password123",
	}
	_ = store.Create(user)

	// 测试删除
	err := store.Delete("tobedeleted")
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// 验证已删除
	_, err = store.Get("tobedeleted")
	if err != ErrUserNotFound {
		t.Fatalf("Expected ErrUserNotFound after delete, got: %v", err)
	}
}

func TestMemoryUserStore_List(t *testing.T) {
	store := NewMemoryUserStore()

	users := store.List()
	if len(users) < 1 {
		t.Fatalf("Expected at least 1 user (admin), got: %d", len(users))
	}
}
