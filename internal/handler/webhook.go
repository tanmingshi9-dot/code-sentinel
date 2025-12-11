package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"code-sentinel/internal/model"
	"code-sentinel/pkg/signature"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) HandleGitHubWebhook(c *gin.Context) {
	eventType := c.GetHeader("X-GitHub-Event")
	sig := c.GetHeader("X-Hub-Signature-256")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	webhookSecret := h.config.GitHub.WebhookSecret
	if !signature.VerifyGitHubSignature(body, sig, webhookSecret) {
		h.logger.Warn("Invalid webhook signature",
			zap.String("event", eventType),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	h.logger.Info("Received GitHub webhook",
		zap.String("event", eventType),
	)

	switch eventType {
	case "pull_request":
		h.handlePullRequest(c, body)
	case "ping":
		c.JSON(http.StatusOK, gin.H{"status": "pong"})
	default:
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "event": eventType})
	}
}

func (h *Handler) handlePullRequest(c *gin.Context, body []byte) {
	var event model.PullRequestEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		if err := bindJSON(body, &event); err != nil {
			h.logger.Error("Failed to parse PR event", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
	}

	if event.Action != "opened" && event.Action != "synchronize" {
		h.logger.Info("Ignoring PR action",
			zap.String("action", event.Action),
			zap.Int("pr_number", event.Number),
		)
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "action": event.Action})
		return
	}

	go func() {
		ctx := context.Background()
		if err := h.analyzerSvc.AnalyzePR(ctx, &event); err != nil {
			h.logger.Error("Failed to analyze PR",
				zap.String("repo", event.Repository.FullName),
				zap.Int("pr_number", event.Number),
				zap.Error(err),
			)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"status":    "processing",
		"repo":      event.Repository.FullName,
		"pr_number": event.Number,
	})
}

func bindJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
