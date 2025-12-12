package handler

import (
	"net/http"
	"strconv"

	"code-sentinel/internal/model"
	"code-sentinel/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Repo handlers

func (h *Handler) ListRepos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	repos, total, err := h.store.ListRepos(c.Request.Context(), page, pageSize, search)
	if err != nil {
		h.logger.Error("Failed to list repos", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list repos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     repos,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *Handler) CreateRepo(c *gin.Context) {
	var req struct {
		FullName      string `json:"full_name" binding:"required"`
		WebhookSecret string `json:"webhook_secret"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repo := &model.Repo{
		FullName:      req.FullName,
		WebhookSecret: req.WebhookSecret,
		Enabled:       true,
	}

	// Parse owner and name from full_name
	parts := splitFullName(req.FullName)
	if len(parts) == 2 {
		repo.Owner = parts[0]
		repo.Name = parts[1]
	}

	if err := h.store.CreateRepo(c.Request.Context(), repo); err != nil {
		h.logger.Error("Failed to create repo", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create repo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    repo,
	})
}

func (h *Handler) GetRepo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	repo, err := h.store.GetRepo(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "repo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    repo,
	})
}

func (h *Handler) UpdateRepo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	repo, err := h.store.GetRepo(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "repo not found"})
		return
	}

	var req struct {
		WebhookSecret *string `json:"webhook_secret"`
		Enabled       *bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.WebhookSecret != nil {
		repo.WebhookSecret = *req.WebhookSecret
	}
	if req.Enabled != nil {
		repo.Enabled = *req.Enabled
	}

	if err := h.store.UpdateRepo(c.Request.Context(), repo); err != nil {
		h.logger.Error("Failed to update repo", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update repo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    repo,
	})
}

func (h *Handler) DeleteRepo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.store.DeleteRepo(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to delete repo", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete repo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// Review handlers

func (h *Handler) ListReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	prNumber, _ := strconv.Atoi(c.Query("pr_number"))

	filter := &store.ReviewFilter{
		RepoFullName: c.Query("repo"),
		Status:       c.Query("status"),
		PRNumber:     prNumber,
		StartDate:    c.Query("start_date"),
		EndDate:      c.Query("end_date"),
	}

	reviews, total, err := h.store.ListReviews(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list reviews", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     reviews,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *Handler) GetReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	review, err := h.store.GetReview(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    review,
	})
}

// Config handlers

func (h *Handler) ListConfigs(c *gin.Context) {
	configs, err := h.store.ListConfigs(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    configs,
	})
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.store.SetConfig(c.Request.Context(), key, req.Value, req.Description); err != nil {
		h.logger.Error("Failed to update config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ToggleRepo 切换仓库状态
func (h *Handler) ToggleRepo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	repo, err := h.repoSvc.ToggleRepo(c.Request.Context(), uint(id), req.Enabled)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "repo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    repo,
	})
}

// Feedback handlers

func (h *Handler) ListFeedbacks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := &store.FeedbackFilter{
		RepoFullName: c.Query("repo"),
		Category:     c.Query("category"),
		Severity:     c.Query("severity"),
		StartDate:    c.Query("start_date"),
		EndDate:      c.Query("end_date"),
	}

	feedbacks, total, err := h.feedbackSvc.ListFeedbacks(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list feedbacks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to list feedbacks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     feedbacks,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *Handler) CreateFeedback(c *gin.Context) {
	var feedback model.Feedback
	if err := c.ShouldBindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := h.feedbackSvc.CreateFeedback(c.Request.Context(), &feedback); err != nil {
		h.logger.Error("Failed to create feedback", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to create feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":         feedback.ID,
			"created_at": feedback.CreatedAt,
		},
	})
}

func (h *Handler) GetFeedbackStats(c *gin.Context) {
	repoFullName := c.Query("repo")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	stats, err := h.feedbackSvc.GetFeedbackStats(c.Request.Context(), repoFullName, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get feedback stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get feedback stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetConfigTemplates 获取配置模板
func (h *Handler) GetConfigTemplates(c *gin.Context) {
	templates := h.repoSvc.GetConfigTemplates()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"templates": templates,
		},
	})
}

// Helper functions

func splitFullName(fullName string) []string {
	for i, c := range fullName {
		if c == '/' {
			return []string{fullName[:i], fullName[i+1:]}
		}
	}
	return []string{fullName}
}
