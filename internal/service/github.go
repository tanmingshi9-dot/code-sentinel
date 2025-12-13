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

type GitHubService struct {
	client *resty.Client
	config config.GitHubConfig
	logger *zap.Logger
}

func NewGitHubService(cfg config.GitHubConfig, logger *zap.Logger) *GitHubService {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Accept", "application/vnd.github.v3+json").
		SetHeader("User-Agent", "Code-Sentinel/1.0").
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second)

	if cfg.Token != "" {
		client.SetHeader("Authorization", "Bearer "+cfg.Token)
	}

	return &GitHubService{
		client: client,
		config: cfg,
		logger: logger,
	}
}

func (s *GitHubService) GetPRDiff(ctx context.Context, repoFullName string, prNumber int) (string, error) {
	s.logger.Info("Fetching PR diff",
		zap.String("repo", repoFullName),
		zap.Int("pr_number", prNumber),
	)

	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/vnd.github.v3.diff").
		Get(fmt.Sprintf("/repos/%s/pulls/%d", repoFullName, prNumber))

	if err != nil {
		return "", fmt.Errorf("failed to get PR diff: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("GitHub API error: %d %s", resp.StatusCode(), resp.String())
	}

	return resp.String(), nil
}

func (s *GitHubService) GetPRFiles(ctx context.Context, repoFullName string, prNumber int) ([]model.PRFile, error) {
	s.logger.Info("Fetching PR files",
		zap.String("repo", repoFullName),
		zap.Int("pr_number", prNumber),
	)

	var files []model.PRFile
	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&files).
		Get(fmt.Sprintf("/repos/%s/pulls/%d/files", repoFullName, prNumber))

	if err != nil {
		return nil, fmt.Errorf("failed to get PR files: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("GitHub API error: %d %s", resp.StatusCode(), resp.String())
	}

	return files, nil
}

func (s *GitHubService) CreatePRComment(ctx context.Context, repoFullName string, prNumber int, body string) error {
	s.logger.Info("Creating PR comment",
		zap.String("repo", repoFullName),
		zap.Int("pr_number", prNumber),
	)

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"body": body}).
		Post(fmt.Sprintf("/repos/%s/issues/%d/comments", repoFullName, prNumber))

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	if resp.StatusCode() != 201 {
		return fmt.Errorf("GitHub API error: %d %s", resp.StatusCode(), resp.String())
	}

	s.logger.Info("PR comment created successfully",
		zap.String("repo", repoFullName),
		zap.Int("pr_number", prNumber),
	)

	return nil
}

func (s *GitHubService) GetWebhookSecret() string {
	return s.config.WebhookSecret
}

// NewGitHubServiceWithConfig 使用简化配置创建 GitHub 服务（用于仓库级配置）
func NewGitHubServiceWithConfig(cfg GitHubConfig, logger *zap.Logger) *GitHubService {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Accept", "application/vnd.github.v3+json").
		SetHeader("User-Agent", "Code-Sentinel/1.0").
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second)

	if cfg.Token != "" {
		client.SetHeader("Authorization", "Bearer "+cfg.Token)
	}

	return &GitHubService{
		client: client,
		config: config.GitHubConfig{
			Token:   cfg.Token,
			BaseURL: baseURL,
		},
		logger: logger,
	}
}
