package handler

import (
	"code-sentinel/internal/config"
	"code-sentinel/internal/service"
	"code-sentinel/internal/store"

	"go.uber.org/zap"
)

type Handler struct {
	analyzerSvc *service.AnalyzerService
	repoSvc     *service.RepoService
	feedbackSvc *service.FeedbackService
	store       store.Store
	config      *config.Config
	logger      *zap.Logger
}

func NewHandler(
	analyzerSvc *service.AnalyzerService,
	repoSvc *service.RepoService,
	feedbackSvc *service.FeedbackService,
	store store.Store,
	cfg *config.Config,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		analyzerSvc: analyzerSvc,
		repoSvc:     repoSvc,
		feedbackSvc: feedbackSvc,
		store:       store,
		config:      cfg,
		logger:      logger,
	}
}
