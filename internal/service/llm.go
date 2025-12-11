package service

import (
	"context"
	"fmt"
	"time"

	"code-sentinel/internal/config"
	"code-sentinel/internal/model"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type LLMService struct {
	client *resty.Client
	config config.LLMConfig
	logger *zap.Logger
}

func NewLLMService(cfg config.LLMConfig, logger *zap.Logger) *LLMService {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Duration(timeout) * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(2 * time.Second)

	if cfg.APIKey != "" {
		client.SetHeader("Authorization", "Bearer "+cfg.APIKey)
	}

	return &LLMService{
		client: client,
		config: cfg,
		logger: logger,
	}
}

func (s *LLMService) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, int, error) {
	s.logger.Info("Calling LLM API",
		zap.String("model", s.config.Model),
		zap.Int("max_tokens", s.config.MaxTokens),
	)

	req := model.ChatRequest{
		Model: s.config.Model,
		Messages: []model.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   s.config.MaxTokens,
		Temperature: 0.3,
	}

	var resp model.ChatResponse
	httpResp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/chat/completions")

	if err != nil {
		return "", 0, fmt.Errorf("LLM API request failed: %w", err)
	}

	if httpResp.StatusCode() != 200 {
		return "", 0, fmt.Errorf("LLM API error: %d %s", httpResp.StatusCode(), httpResp.String())
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("LLM returned empty response")
	}

	s.logger.Info("LLM API response received",
		zap.Int("prompt_tokens", resp.Usage.PromptTokens),
		zap.Int("completion_tokens", resp.Usage.CompletionTokens),
		zap.Int("total_tokens", resp.Usage.TotalTokens),
	)

	return resp.Choices[0].Message.Content, resp.Usage.TotalTokens, nil
}

func (s *LLMService) GetModel() string {
	return s.config.Model
}
