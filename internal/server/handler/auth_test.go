package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestLoginHandler(t *testing.T) {
	router := setupTestRouter()
	
	// 初始化 JWT 服务
	jwtService := auth.NewJWTService("test-secret", time.Hour)
	
	// 注册路由
	router.PUT("/-/user/:username", func(c *gin.Context) {
		// 简化的登录处理
		var req struct {
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		
		// 测试用户验证
		if req.Name == "admin" && req.Password == "admin" {
			token, _ := jwtService.GenerateToken(&auth.User{Username: req.Name, Role: "admin"})
			c.JSON(201, gin.H{"token": token})
		} else {
			c.JSON(401, gin.H{"error": "invalid credentials"})
		}
	})

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
		wantToken  bool
	}{
		{
			name:       "valid credentials",
			username:   "admin",
			password:   "admin",
			wantStatus: 201,
			wantToken:  true,
		},
		{
			name:       "invalid credentials",
			username:   "admin",
			password:   "wrong",
			wantStatus: 401,
			wantToken:  false,
		},
		{
			name:       "empty password",
			username:   "admin",
			password:   "",
			wantStatus: 401,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"name":     tt.username,
				"password": tt.password,
			})
			req := httptest.NewRequest("PUT", "/-/user/"+tt.username, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			
			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			
			if tt.wantToken {
				if _, ok := resp["token"]; !ok || resp["token"] == "" {
					t.Error("expected token in response")
				}
			} else {
				if _, ok := resp["token"]; ok {
					t.Error("expected no token in response")
				}
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	router := setupTestRouter()
	
	// 受保护的路由
	authorized := router.Group("/")
	authorized.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	})
	{
		authorized.GET("/api/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "no auth header",
			authHeader: "",
			wantStatus: 401,
		},
		{
			name:       "valid bearer token",
			authHeader: "Bearer valid-token",
			wantStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}