package handler

import (
	"net/http"
	"strconv"

	"code-sentinel/internal/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Repo handlers

func (h *Handler) ListRepos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	repos, total, err := h.store.ListRepos(c.Request.Context(), page, pageSize)
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
	repoFullName := c.Query("repo")

	reviews, total, err := h.store.ListReviews(c.Request.Context(), repoFullName, page, pageSize)
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

// Helper functions

func splitFullName(fullName string) []string {
	for i, c := range fullName {
		if c == '/' {
			return []string{fullName[:i], fullName[i+1:]}
		}
	}
	return []string{fullName}
}
