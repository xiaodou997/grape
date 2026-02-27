package auth

import (
	"testing"
	"time"
)

func TestJWTService_GenerateAndValidate(t *testing.T) {
	secret := "test-secret-key"
	expiry := time.Hour

	service := NewJWTService(secret, expiry)

	user := &User{
		Username: "testuser",
		Role:     "developer",
	}

	// 生成 token
	token, err := service.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Token should not be empty")
	}

	// 验证 token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.Username != user.Username {
		t.Fatalf("Expected username %s, got: %s", user.Username, claims.Username)
	}

	if claims.Role != user.Role {
		t.Fatalf("Expected role %s, got: %s", user.Role, claims.Role)
	}
}

func TestJWTService_InvalidToken(t *testing.T) {
	service := NewJWTService("secret", time.Hour)

	// 测试无效 token
	_, err := service.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}
}

func TestJWTService_WrongSecret(t *testing.T) {
	service1 := NewJWTService("secret1", time.Hour)
	service2 := NewJWTService("secret2", time.Hour)

	user := &User{Username: "test", Role: "developer"}

	token, _ := service1.GenerateToken(user)

	// 使用不同密钥验证
	_, err := service2.ValidateToken(token)
	if err == nil {
		t.Fatal("Expected error when validating token with wrong secret")
	}
}
