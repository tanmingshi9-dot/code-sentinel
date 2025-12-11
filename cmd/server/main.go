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
	analyzerSvc := service.NewAnalyzerService(githubSvc, llmSvc, db, logger)

	// 初始化 Handler
	h := handler.NewHandler(analyzerSvc, db, cfg, logger)

	// 设置路由
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ginLogger(logger))

	// 健康检查
	router.GET("/health", h.Health)
	router.GET("/ready", h.Ready)

	// Webhook
	router.POST("/webhook/github", h.HandleGitHubWebhook)

	// API
	api := router.Group("/api/v1")
	{
		api.GET("/repos", h.ListRepos)
		api.POST("/repos", h.CreateRepo)
		api.GET("/repos/:id", h.GetRepo)
		api.PUT("/repos/:id", h.UpdateRepo)
		api.DELETE("/repos/:id", h.DeleteRepo)

		api.GET("/reviews", h.ListReviews)
		api.GET("/reviews/:id", h.GetReview)

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
