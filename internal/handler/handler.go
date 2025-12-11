package handler

import (
	"code-sentinel/internal/config"
	"code-sentinel/internal/service"
	"code-sentinel/internal/store"

	"go.uber.org/zap"
)

type Handler struct {
	analyzerSvc *service.AnalyzerService
	store       store.Store
	config      *config.Config
	logger      *zap.Logger
}

func NewHandler(analyzerSvc *service.AnalyzerService, store store.Store, cfg *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		analyzerSvc: analyzerSvc,
		store:       store,
		config:      cfg,
		logger:      logger,
	}
}
