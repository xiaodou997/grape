package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/graperegistry/grape/internal/auth"
	"github.com/graperegistry/grape/internal/config"
	"github.com/graperegistry/grape/internal/logger"
	"github.com/graperegistry/grape/internal/metrics"
	"github.com/graperegistry/grape/internal/registry"
	"github.com/graperegistry/grape/internal/server/handler"
	"github.com/graperegistry/grape/internal/storage/local"
	"github.com/graperegistry/grape/internal/web"
	"github.com/graperegistry/grape/internal/webhook"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	cfg             *config.Config
	router          *gin.Engine
	apiRouter       *gin.Engine         // npm Registry API 专用路由
	http            *http.Server        // Web UI 服务器
	apiServer       *http.Server        // npm Registry API 服务器
	proxy           *registry.Proxy
	storage         *local.Storage
	userStore       auth.UserStore
	jwtService      *auth.JWTService
	registryHandler *handler.RegistryHandler
	authHandler     *handler.AuthHandler
	publishHandler  *handler.PublishHandler
	apiHandler      *handler.APIHandler
	webhookHandler  *handler.WebhookHandler
	tokenHandler    *handler.TokenHandler
	ownerHandler    *handler.OwnerHandler
	backupHandler   *handler.BackupHandler
	gcHandler       *handler.GCHandler
	webFS           http.FileSystem
	webDist         fs.FS
}

func New(cfg *config.Config, version string) *Server {
	gin.SetMode(gin.ReleaseMode)

	// Web UI 路由器
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(maxBytesMiddleware(50 << 20))
	router.Use(securityHeadersMiddleware())
	router.Use(prometheusMiddleware())

	// npm Registry API 路由器（独立，无中间件）
	apiRouter := gin.New()
	apiRouter.Use(gin.Recovery())
	apiRouter.Use(requestLogger())
	apiRouter.Use(maxBytesMiddleware(50 << 20))
	apiRouter.Use(corsMiddleware())
	apiRouter.Use(prometheusMiddleware())

	// 初始化组件
	proxy := registry.NewProxy(&cfg.Registry)
	storage := local.New(cfg.Storage.Path)

	// 检查 JWT 密钥安全性
	if cfg.Auth.JWTSecret == "grape-secret-key-change-in-production" {
		logger.Warn("⚠️  WARNING: Using default JWT secret! Please change 'auth.jwt_secret' in config.yaml")
		logger.Warn("⚠️  This is a security risk in production environments!")
	}

	// 使用数据库用户存储
	userStore := auth.NewDBUserStore()
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)

	// 检查是否需要创建默认管理员
	createDefaultAdminIfNeeded(userStore)

	// API 端口作为 baseURL
	apiPort := cfg.Server.APIPort
	if apiPort == 0 {
		apiPort = cfg.Server.Port // 如果没有配置 API 端口，使用主端口
	}
	baseURL := fmt.Sprintf("http://localhost:%d", apiPort)

	// 创建 Webhook 分发器
	webhookDispatcher := webhook.NewDispatcher()

	// 创建 handlers
	registryHandler := handler.NewRegistryHandler(proxy, storage, baseURL)
	authHandler := handler.NewAuthHandler(userStore, jwtService, cfg.Auth.AllowRegistration)
	publishHandler := handler.NewPublishHandler(storage, webhookDispatcher)
	webhookHandler := handler.NewWebhookHandler(webhookDispatcher)
	tokenHandler := handler.NewTokenHandler()
	ownerHandler := handler.NewOwnerHandler()
	backupHandler := handler.NewBackupHandler(cfg.Storage.Path)
	gcHandler := handler.NewGCHandler(storage, cfg.Storage.Path)

	// 获取前端文件系统
	webFS := web.GetFileSystem()
	webDist := web.GetDistFS()

	s := &Server{
		cfg:             cfg,
		router:          router,
		apiRouter:       apiRouter,
		proxy:           proxy,
		storage:         storage,
		userStore:       userStore,
		jwtService:      jwtService,
		registryHandler: registryHandler,
		authHandler:     authHandler,
		publishHandler:  publishHandler,
		webhookHandler:  webhookHandler,
		tokenHandler:    tokenHandler,
		ownerHandler:    ownerHandler,
		backupHandler:   backupHandler,
		gcHandler:       gcHandler,
		webFS:           webFS,
		webDist:         webDist,
		http: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		apiServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, apiPort),
			Handler:      apiRouter,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
	}

	// apiHandler 需要引用 s（通过 applyFn），所以在 s 创建后再初始化
	s.apiHandler = handler.NewAPIHandler(storage, cfg.Storage.Path, proxy, cfg, version, s.applyConfig)

	s.setupRoutes()
	return s
}

// applyConfig 热更新运行中组件
func (s *Server) applyConfig(cfg *config.Config) {
	// 更新上游代理
	s.proxy.SetUpstreams(cfg.Registry.Upstreams)
	// 更新 JWT 服务
	s.jwtService.UpdateSecret(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)
	// 更新自助注册开关
	s.authHandler.SetAllowRegistration(cfg.Auth.AllowRegistration)
	// 更新日志级别
	if err := logger.SetLevel(cfg.Log.Level); err != nil {
		logger.Warnf("Failed to update log level: %v", err)
	}
	logger.Infof("✅ Config hot-reloaded successfully")
}

// createDefaultAdminIfNeeded 如果数据库中没有用户，创建默认管理员
func createDefaultAdminIfNeeded(userStore auth.UserStore) {
	users := userStore.List()
	if len(users) == 0 {
		adminUser := &auth.User{
			Username: "admin",
			Email:    "admin@grape.local",
			Password: "admin",
			Role:     "admin",
		}
		if err := userStore.Create(adminUser); err != nil {
			logger.Warnf("Failed to create default admin: %v", err)
		} else {
			logger.Info("👤 Created default admin user: admin / admin")
		}
	}
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

	// =====================================
	// Web UI 路由（仅处理前端页面）
	// =====================================
	s.router.GET("/-/health", s.handleHealth)
	s.router.GET("/-/metrics", gin.WrapH(promhttp.Handler()))
	
	// 管理 API（带认证）
	webAPI := s.router.Group("/-")
	webAPI.Use(authMiddleware)
	{
		webAPI.GET("/api/packages", s.apiHandler.ListPackages)
		webAPI.GET("/api/stats", s.apiHandler.GetStats)
		webAPI.GET("/api/search", s.apiHandler.SearchPackages)
		webAPI.GET("/api/upstreams", s.apiHandler.GetUpstreams)
		webAPI.GET("/api/user", s.authHandler.GetCurrentUser)
		webAPI.DELETE("/api/session", s.authHandler.Logout)

		// Token 管理 API
		webAPI.GET("/npm/v1/tokens", s.tokenHandler.ListTokens)
		webAPI.POST("/npm/v1/tokens", s.tokenHandler.CreateToken)
		webAPI.DELETE("/npm/v1/tokens/token/:id", s.tokenHandler.DeleteToken)

		admin := webAPI.Group("/api/admin")
		admin.Use(auth.RequireAdmin())
		{
			admin.GET("/users", s.authHandler.ListUsers)
			admin.POST("/users", s.authHandler.CreateUser)
			admin.PUT("/users/:username", s.authHandler.UpdateUser)
			admin.DELETE("/users/:username", s.authHandler.DeleteUser)
			admin.GET("/system", s.apiHandler.GetSystemInfo)
			admin.GET("/config", s.apiHandler.GetConfig)
			admin.PUT("/config", s.apiHandler.UpdateConfig)
			admin.GET("/audit-logs", handler.GetAuditLogs)
			admin.GET("/webhooks", s.webhookHandler.ListWebhooks)
			admin.POST("/webhooks", s.webhookHandler.CreateWebhook)
			admin.PUT("/webhooks/:id", s.webhookHandler.UpdateWebhook)
			admin.DELETE("/webhooks/:id", s.webhookHandler.DeleteWebhook)
			admin.POST("/webhooks/:id/test", s.webhookHandler.TestWebhook)
			// Package owner 管理
			admin.GET("/packages/:name/owners", s.ownerHandler.ListPackageOwnersAdmin)
			admin.POST("/packages/:name/owners", s.ownerHandler.SetPackageOwnerAdmin)
			admin.DELETE("/packages/:name/owners/:username", s.ownerHandler.RemovePackageOwnerAdmin)
			// Backup & Restore
			admin.GET("/backup/info", s.backupHandler.GetBackupInfo)
			admin.GET("/backup/download", s.backupHandler.CreateBackup)
			admin.POST("/backup/restore", s.backupHandler.RestoreBackup)
			admin.GET("/backup/list", s.backupHandler.ListBackups)
			// Garbage Collection
			admin.GET("/gc/stats", s.gcHandler.GetGCStats)
			admin.GET("/gc/analyze", s.gcHandler.AnalyzeGC)
			admin.POST("/gc/run", s.gcHandler.RunGC)
			// Package deprecation
			admin.POST("/packages/:name/deprecate", s.gcHandler.DeprecatePackage)
			admin.DELETE("/packages/:name/deprecate", s.gcHandler.UndeprecatePackage)
		}
		webAPI.PUT("/user/:username", s.authHandler.Login)
		webAPI.PUT("/user/:username/*rev", s.authHandler.Login)
	}
	
	// 前端静态资源和 SPA
	s.router.NoRoute(authMiddleware, s.serveFrontend)

	// =====================================
	// npm Registry API 路由（独立路由器）
	// =====================================
	// Health check
	s.apiRouter.GET("/-/health", s.handleHealth)
	
	// 认证 API（npm login）
	s.apiRouter.PUT("/-/user/:username", s.authHandler.Login)
	s.apiRouter.PUT("/-/user/:username/*rev", s.authHandler.Login)

	// 管理 API（带认证）
	apiRegistry := s.apiRouter.Group("/-")
	apiRegistry.Use(authMiddleware)
	{
		apiRegistry.GET("/api/user", s.authHandler.GetCurrentUser)
		apiRegistry.DELETE("/api/session", s.authHandler.Logout)
		// npm owner 命令兼容 API
		apiRegistry.GET("/package/:name/collaborators", s.ownerHandler.ListOwners)
		apiRegistry.PUT("/package/:name/collaborators/:username", s.ownerHandler.AddOwner)
		apiRegistry.DELETE("/package/:name/collaborators/:username", s.ownerHandler.RemoveOwner)
	}
	
	// npm Registry API - 使用 NoRoute 处理所有包请求（包括 scoped 包、发布、删除）
	s.apiRouter.NoRoute(s.handleRegistryRequest)
}

// handleRegistryRequest 统一处理 npm registry 请求
func (s *Server) handleRegistryRequest(c *gin.Context) {
	// 使用 Request.URL.Path 获取完整路径（NoRoute 不会设置参数）
	reqPath := c.Request.URL.Path
	
	// 去除前导 /
	reqPath = strings.TrimPrefix(reqPath, "/")
	
	// 空路径或根路径
	if reqPath == "" || reqPath == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}
	
	logger.Debugf("handleRegistryRequest: path=%s, method=%s", reqPath, c.Request.Method)
	
	// 判断是否为 tarball 请求 (包含 /-/)
	if strings.Contains(reqPath, "/-/") {
		// tarball 下载或删除
		idx := strings.Index(reqPath, "/-/")
		packageName := reqPath[:idx]
		filename := reqPath[idx+3:]
		
		// 直接设置 Params
		c.Params = gin.Params{
			{Key: "package", Value: packageName},
			{Key: "filename", Value: filename},
		}
		
		switch c.Request.Method {
		case http.MethodGet:
			s.registryHandler.GetTarball(c)
		case http.MethodDelete:
			// 删除需要认证
			authMiddleware := auth.AuthMiddleware(s.jwtService, s.userStore)
			authMiddleware(c)
			if c.IsAborted() {
				return
			}
			s.publishHandler.Unpublish(c)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		}
	} else {
		// 包元数据或发布
		// 直接设置 Params
		c.Params = gin.Params{
			{Key: "package", Value: reqPath},
		}
		
		switch c.Request.Method {
		case http.MethodGet:
			s.registryHandler.GetPackage(c)
		case http.MethodPut:
			// 发布需要认证
			authMiddleware := auth.AuthMiddleware(s.jwtService, s.userStore)
			authMiddleware(c)
			if c.IsAborted() {
				return
			}
			s.publishHandler.Publish(c)
		case http.MethodDelete:
			// 删除需要认证
			authMiddleware := auth.AuthMiddleware(s.jwtService, s.userStore)
			authMiddleware(c)
			if c.IsAborted() {
				return
			}
			s.publishHandler.Unpublish(c)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		}
	}
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

	// 单段路径，可能是包名（如 /vite, /lodash）
	// 排除已知的前端路径
	knownFrontendPaths := map[string]bool{
		"/": true, "/login": true, "/register": true,
		"/packages": true, "/package": true,
		"/admin": true, "/settings": true, "/guide": true,
		"/health": true, "/metrics": true,
	}
	
	// 精确匹配已知前端路径
	if knownFrontendPaths[path] {
		return false
	}
	
	// 排除静态资源路径
	if strings.HasPrefix(path, "/assets/") || 
	   strings.HasPrefix(path, "/static/") ||
	   strings.HasSuffix(path, ".js") ||
	   strings.HasSuffix(path, ".css") ||
	   strings.HasSuffix(path, ".png") ||
	   strings.HasSuffix(path, ".svg") ||
	   strings.HasSuffix(path, ".ico") {
		return false
	}
	
	// 如果路径只有一段（如 /vite），假设是 npm 包
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathSegments) == 1 && pathSegments[0] != "" {
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
	webAddr := s.http.Addr
	apiAddr := s.apiServer.Addr
	
	logger.Infof("🚀 Grape Web UI server starting on http://%s", webAddr)
	logger.Infof("📦 npm Registry API server starting on http://%s", apiAddr)
	logger.Infof("📦 npm registry upstream: %s", s.cfg.Registry.Upstream)
	logger.Infof("💾 Storage: %s", s.cfg.Storage.Path)
	logger.Infof("💾 Database: %s", s.cfg.Database.DSN)
	
	// 启动两个服务器
	go func() {
		if err := s.apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start API server: %v", err)
		}
	}()
	
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("🛑 Shutting down servers...")
	handler.StopLoginLimiter()
	
	// 关闭 Web UI 服务器
	if err := s.http.Shutdown(ctx); err != nil {
		logger.Errorf("Web UI server shutdown error: %v", err)
	}
	
	// 关闭 API 服务器
	if err := s.apiServer.Shutdown(ctx); err != nil {
		logger.Errorf("API server shutdown error: %v", err)
	}
	
	logger.Info("👋 Grape stopped")
	return nil
}

// maxBytesMiddleware 限制请求体大小
func maxBytesMiddleware(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}

// prometheusMiddleware 收集 HTTP 请求指标
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// securityHeadersMiddleware 添加 HTTP 安全响应头
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// 允许连接到同源的 API 端口 (4874) 和默认端口 (4873)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' http://localhost:4874 http://127.0.0.1:4874")
		c.Next()
	}
}

// corsMiddleware 处理跨域请求
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// 允许来自 Web UI 端口的请求
		if origin == "http://localhost:4873" || origin == "http://127.0.0.1:4873" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}
