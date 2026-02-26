package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/config"
	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/registry"
	"github.com/graperegistry/grape/internal/server/handler"
	"github.com/graperegistry/grape/internal/storage/local"
	"github.com/graperegistry/grape/internal/web"
)

type Server struct {
	cfg             *config.Config
	router          *gin.Engine
	http            *http.Server
	proxy           *registry.Proxy
	storage         *local.Storage
	userStore       auth.UserStore
	jwtService      *auth.JWTService
	registryHandler *handler.RegistryHandler
	authHandler     *handler.AuthHandler
	publishHandler  *handler.PublishHandler
	apiHandler      *handler.APIHandler
	webFS           http.FileSystem
	webDist         fs.FS
}

func New(cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())

	// 初始化组件
	proxy := registry.NewProxy(&cfg.Registry)
	storage := local.New(cfg.Storage.Path)
	userStore := auth.NewMemoryUserStore()
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)

	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)

	// 创建 handlers
	registryHandler := handler.NewRegistryHandler(proxy, storage, baseURL)
	authHandler := handler.NewAuthHandler(userStore, jwtService)
	publishHandler := handler.NewPublishHandler(storage)
	apiHandler := handler.NewAPIHandler(storage, cfg.Storage.Path)

	// 获取前端文件系统
	webFS := web.GetFileSystem()
	webDist := web.GetDistFS()

	s := &Server{
		cfg:             cfg,
		router:          router,
		proxy:           proxy,
		storage:         storage,
		userStore:       userStore,
		jwtService:      jwtService,
		registryHandler: registryHandler,
		authHandler:     authHandler,
		publishHandler:  publishHandler,
		apiHandler:      apiHandler,
		webFS:           webFS,
		webDist:         webDist,
		http: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
	}

	s.setupRoutes()
	return s
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.Infof("%s %s %d %v", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
	}
}

func (s *Server) setupRoutes() {
	authMiddleware := auth.AuthMiddleware(s.jwtService, s.userStore)

	// Health check (无需认证)
	s.router.GET("/-/health", s.handleHealth)

	// API routes
	api := s.router.Group("/-")
	api.Use(authMiddleware)
	{
		api.GET("/api/packages", s.apiHandler.ListPackages)
		api.GET("/api/stats", s.apiHandler.GetStats)
		api.GET("/api/search", s.apiHandler.SearchPackages)
		api.GET("/api/user", s.authHandler.GetCurrentUser)
		api.DELETE("/api/session", s.authHandler.Logout)

		admin := api.Group("/api/admin")
		admin.Use(auth.RequireAdmin())
		{
			admin.GET("/users", s.authHandler.ListUsers)
			admin.POST("/users", s.authHandler.CreateUser)
			admin.DELETE("/users/:username", s.authHandler.DeleteUser)
		}

		api.PUT("/user/:username", s.authHandler.Login)
		api.PUT("/user/:username/*rev", s.authHandler.Login)
	}

	// 所有其他路由
	s.router.NoRoute(authMiddleware, s.handleRequest)
}

// handleRequest 处理所有其他请求
func (s *Server) handleRequest(c *gin.Context) {
	path := c.Request.URL.Path

	// npm tarball 请求
	if strings.Contains(path, "/-/") && !strings.HasPrefix(path, "/-/") {
		s.handleNpmTarball(c)
		return
	}

	// 判断是 npm 包请求还是前端页面
	if s.isNpmPackageRequest(c) {
		s.handleNpmPackage(c)
		return
	}

	// 前端静态资源或 SPA
	s.serveFrontend(c)
}

// handleNpmTarball 处理 npm tarball 下载
func (s *Server) handleNpmTarball(c *gin.Context) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/")
	idx := strings.Index(path, "/-/")
	if idx == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	packageName := path[:idx]
	filename := path[idx+3:]
	c.Params = append(c.Params,
		gin.Param{Key: "package", Value: packageName},
		gin.Param{Key: "filename", Value: filename},
	)
	s.registryHandler.GetTarball(c)
}

// handleNpmPackage 处理 npm 包请求
func (s *Server) handleNpmPackage(c *gin.Context) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/")
	c.Params = append(c.Params, gin.Param{Key: "package", Value: path})

	switch c.Request.Method {
	case http.MethodGet:
		s.registryHandler.GetPackage(c)
	case http.MethodPut:
		s.publishHandler.Publish(c)
	case http.MethodDelete:
		s.publishHandler.Unpublish(c)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

// serveFrontend 提供前端静态文件
func (s *Server) serveFrontend(c *gin.Context) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// 尝试读取文件
	data, err := fs.ReadFile(s.webDist, path)
	if err != nil {
		// 文件不存在，返回 index.html (SPA fallback)
		data, err = fs.ReadFile(s.webDist, "index.html")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		path = "index.html"
	}

	// 设置 Content-Type
	contentType := "text/html"
	if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript"
	} else if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(path, ".svg") {
		contentType = "image/svg+xml"
	} else if strings.HasSuffix(path, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(path, ".ico") {
		contentType = "image/x-icon"
	}

	c.Data(http.StatusOK, contentType, data)
}

// isNpmPackageRequest 判断是否为 npm 包请求
func (s *Server) isNpmPackageRequest(c *gin.Context) bool {
	path := c.Request.URL.Path

	// scoped package (@scope/name)
	if strings.HasPrefix(path, "/@") {
		return true
	}

	// Accept header 包含 application/json
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}

	// User-Agent 包含 npm/yarn/pnpm/bun
	userAgent := c.GetHeader("User-Agent")
	if strings.Contains(userAgent, "npm/") ||
		strings.Contains(userAgent, "yarn/") ||
		strings.Contains(userAgent, "pnpm/") ||
		strings.Contains(userAgent, "bun/") {
		return true
	}

	return false
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) Start() error {
	addr := s.http.Addr
	logger.Infof("🚀 Grape server starting on http://%s", addr)
	logger.Infof("📦 npm registry: %s", s.cfg.Registry.Upstream)
	logger.Infof("💾 Storage: %s", s.cfg.Storage.Path)
	logger.Infof("👤 Default user: admin / admin")
	logger.Infof("🌐 Web UI: http://%s", addr)

	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("🛑 Shutting down server...")
	return s.http.Shutdown(ctx)
}

func (s *Server) Router() *gin.Engine {
	return s.router
}
