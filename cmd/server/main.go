package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code-sentinel/internal/config"
	"code-sentinel/internal/handler"
	"code-sentinel/internal/service"
	"code-sentinel/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化存储层
	db, err := store.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	sugar.Info("Database initialized")

	// 初始化服务层
	githubSvc := service.NewGitHubService(cfg.GitHub, logger)
	llmSvc := service.NewLLMService(cfg.LLM, logger)
	repoSvc := service.NewRepoService(db, logger)
	feedbackSvc := service.NewFeedbackService(db, githubSvc, logger)
	// 构建默认配置用于仓库级覆盖
	defaultLLMCfg := service.LLMConfig{
		Provider:  cfg.LLM.Provider,
		APIKey:    cfg.LLM.APIKey,
		Model:     cfg.LLM.Model,
		BaseURL:   cfg.LLM.BaseURL,
		Timeout:   cfg.LLM.Timeout,
		MaxTokens: cfg.LLM.MaxTokens,
	}
	defaultGHCfg := service.GitHubConfig{
		Token:   cfg.GitHub.Token,
		BaseURL: cfg.GitHub.BaseURL,
	}
	analyzerSvc := service.NewAnalyzerService(githubSvc, llmSvc, db, logger, defaultLLMCfg, defaultGHCfg)

	// 初始化 Handler
	h := handler.NewHandler(analyzerSvc, repoSvc, feedbackSvc, db, cfg, logger)

	// 设置路由
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ginLogger(logger))

	// 健康检查
	router.GET("/health", h.Health)
	router.GET("/ready", h.Ready)

	// Webhook
	router.POST("/webhook/github", h.HandleGitHubWebhook)

	// 静态文件服务（前端）
	router.Static("/assets", "./web/dist/assets")
	router.StaticFile("/vite.svg", "./web/dist/vite.svg")
	router.NoRoute(func(c *gin.Context) {
		// API 路由返回 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
			return
		}
		// 其他路由返回 index.html（SPA）
		c.File("./web/dist/index.html")
	})

	// API
	api := router.Group("/api/v1")
	{
		// 仓库管理
		api.GET("/repos", h.ListRepos)
		api.POST("/repos", h.CreateRepo)
		api.GET("/repos/:id", h.GetRepo)
		api.PUT("/repos/:id", h.UpdateRepo)
		api.DELETE("/repos/:id", h.DeleteRepo)
		api.PUT("/repos/:id/toggle", h.ToggleRepo)

		// 审查记录
		api.GET("/reviews", h.ListReviews)
		api.GET("/reviews/:id", h.GetReview)

		// 反馈管理
		api.GET("/feedbacks", h.ListFeedbacks)
		api.POST("/feedbacks", h.CreateFeedback)
		api.GET("/feedbacks/stats", h.GetFeedbackStats)

		// 配置模板
		api.GET("/config-templates", h.GetConfigTemplates)

		// 全局配置
		api.GET("/configs", h.ListConfigs)
		api.PUT("/configs/:key", h.UpdateConfig)
	}

	// 启动服务器
	srv := &http.Server{
		Addr:    cfg.Server.Addr(),
		Handler: router,
	}

	go func() {
		sugar.Infof("Server starting on %s", cfg.Server.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		sugar.Fatalf("Server forced to shutdown: %v", err)
	}

	sugar.Info("Server exited")
}

func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}
