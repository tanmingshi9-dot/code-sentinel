package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"code-sentinel/internal/model"
	"code-sentinel/internal/store"

	"go.uber.org/zap"
)

// RepoService 仓库管理服务
type RepoService struct {
	store  store.Store
	logger *zap.Logger
}

// NewRepoService 创建 RepoService 实例
func NewRepoService(store store.Store, logger *zap.Logger) *RepoService {
	return &RepoService{
		store:  store,
		logger: logger,
	}
}

// CreateRepoRequest 创建仓库请求
type CreateRepoRequest struct {
	FullName      string              `json:"full_name" binding:"required"`
	WebhookSecret string              `json:"webhook_secret"`
	Enabled       bool                `json:"enabled"`
	Config        *model.ReviewConfig `json:"config"`
}

// UpdateRepoRequest 更新仓库请求
type UpdateRepoRequest struct {
	WebhookSecret *string             `json:"webhook_secret"`
	Enabled       *bool               `json:"enabled"`
	Config        *model.ReviewConfig `json:"config"`
}

// repoNameRegex 仓库名格式校验
var repoNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)

// ListRepos 获取仓库列表
func (s *RepoService) ListRepos(ctx context.Context, page, pageSize int, search string) ([]model.Repo, int64, error) {
	return s.store.ListRepos(ctx, page, pageSize, search)
}

// CreateRepo 创建仓库
func (s *RepoService) CreateRepo(ctx context.Context, req *CreateRepoRequest) (*model.Repo, error) {
	// 1. 验证仓库名格式
	if !repoNameRegex.MatchString(req.FullName) {
		return nil, errors.New("invalid repo name format, expected: owner/repo")
	}

	// 2. 检查是否已存在
	existing, err := s.store.GetRepoByFullName(ctx, req.FullName)
	if err == nil && existing != nil {
		return nil, errors.New("repo already exists")
	}

	// 3. 解析 owner 和 name
	owner, name := splitRepoFullName(req.FullName)

	// 4. 序列化配置
	var configJSON string
	if req.Config != nil {
		configBytes, err := json.Marshal(req.Config)
		if err != nil {
			return nil, err
		}
		configJSON = string(configBytes)
	} else {
		// 使用默认配置
		defaultConfig := s.GetDefaultConfig()
		configBytes, _ := json.Marshal(defaultConfig)
		configJSON = string(configBytes)
	}

	// 5. 创建仓库记录
	repo := &model.Repo{
		FullName:      req.FullName,
		Owner:         owner,
		Name:          name,
		WebhookSecret: req.WebhookSecret,
		Enabled:       req.Enabled,
		Config:        configJSON,
	}

	if err := s.store.CreateRepo(ctx, repo); err != nil {
		s.logger.Error("Failed to create repo", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Repo created", zap.String("full_name", req.FullName))
	return repo, nil
}

// GetRepo 获取仓库详情
func (s *RepoService) GetRepo(ctx context.Context, id uint) (*model.Repo, error) {
	return s.store.GetRepo(ctx, id)
}

// GetRepoByFullName 根据全名获取仓库
func (s *RepoService) GetRepoByFullName(ctx context.Context, fullName string) (*model.Repo, error) {
	return s.store.GetRepoByFullName(ctx, fullName)
}

// UpdateRepo 更新仓库
func (s *RepoService) UpdateRepo(ctx context.Context, id uint, req *UpdateRepoRequest) (*model.Repo, error) {
	repo, err := s.store.GetRepo(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.WebhookSecret != nil {
		repo.WebhookSecret = *req.WebhookSecret
	}
	if req.Enabled != nil {
		repo.Enabled = *req.Enabled
	}
	if req.Config != nil {
		configBytes, err := json.Marshal(req.Config)
		if err != nil {
			return nil, err
		}
		repo.Config = string(configBytes)
	}

	if err := s.store.UpdateRepo(ctx, repo); err != nil {
		s.logger.Error("Failed to update repo", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Repo updated", zap.Uint("id", id))
	return repo, nil
}

// DeleteRepo 删除仓库
func (s *RepoService) DeleteRepo(ctx context.Context, id uint) error {
	if err := s.store.DeleteRepo(ctx, id); err != nil {
		s.logger.Error("Failed to delete repo", zap.Error(err))
		return err
	}
	s.logger.Info("Repo deleted", zap.Uint("id", id))
	return nil
}

// ToggleRepo 切换仓库状态
func (s *RepoService) ToggleRepo(ctx context.Context, id uint, enabled bool) (*model.Repo, error) {
	repo, err := s.store.GetRepo(ctx, id)
	if err != nil {
		return nil, err
	}

	repo.Enabled = enabled
	if err := s.store.UpdateRepo(ctx, repo); err != nil {
		return nil, err
	}

	s.logger.Info("Repo toggled", zap.Uint("id", id), zap.Bool("enabled", enabled))
	return repo, nil
}

// GetRepoConfig 获取仓库配置
func (s *RepoService) GetRepoConfig(ctx context.Context, fullName string) (*model.ReviewConfig, error) {
	repo, err := s.store.GetRepoByFullName(ctx, fullName)
	if err != nil {
		return nil, err
	}

	if !repo.Enabled {
		return nil, errors.New("repo is disabled")
	}

	if repo.Config == "" {
		return s.GetDefaultConfig(), nil
	}

	var config model.ReviewConfig
	if err := json.Unmarshal([]byte(repo.Config), &config); err != nil {
		s.logger.Warn("Failed to parse repo config, using default",
			zap.String("repo", fullName),
			zap.Error(err),
		)
		return s.GetDefaultConfig(), nil
	}

	return &config, nil
}

// GetDefaultConfig 获取默认配置
func (s *RepoService) GetDefaultConfig() *model.ReviewConfig {
	return &model.ReviewConfig{
		LLMProvider:  "openai",
		Model:        "gpt-4-turbo",
		MaxTokens:    4096,
		ReviewFocus:  []string{"security", "performance", "logic"},
		MinSeverity:  "P1",
		Languages:    []string{"go", "python", "javascript", "java"},
		IgnoreFiles:  []string{"*.test.go", "vendor/*", "node_modules/*"},
		MaxDiffLines: 1000,
		AutoReview:   true,
	}
}

// GetConfigTemplates 获取配置模板列表
func (s *RepoService) GetConfigTemplates() []ConfigTemplate {
	return []ConfigTemplate{
		{
			Name:        "默认配置",
			Description: "适用于大多数项目",
			Config: &model.ReviewConfig{
				LLMProvider:  "openai",
				Model:        "gpt-4-turbo",
				ReviewFocus:  []string{"security", "performance", "logic"},
				Languages:    []string{"go", "python", "javascript"},
				MinSeverity:  "P1",
				MaxDiffLines: 1000,
				AutoReview:   true,
			},
		},
		{
			Name:        "前端项目",
			Description: "React/Vue 项目",
			Config: &model.ReviewConfig{
				LLMProvider:  "openai",
				Model:        "gpt-4-turbo",
				ReviewFocus:  []string{"security", "performance"},
				Languages:    []string{"javascript", "typescript"},
				IgnoreFiles:  []string{"dist/*", "build/*", "*.test.tsx"},
				MinSeverity:  "P1",
				MaxDiffLines: 1000,
				AutoReview:   true,
			},
		},
		{
			Name:        "后端项目",
			Description: "Go/Java 项目",
			Config: &model.ReviewConfig{
				LLMProvider:  "openai",
				Model:        "gpt-4-turbo",
				ReviewFocus:  []string{"security", "performance", "logic"},
				Languages:    []string{"go", "java"},
				IgnoreFiles:  []string{"vendor/*", "*.test.go"},
				MinSeverity:  "P0",
				MaxDiffLines: 1000,
				AutoReview:   true,
			},
		},
	}
}

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Config      *model.ReviewConfig `json:"config"`
}

// splitRepoFullName 分割仓库全名
func splitRepoFullName(fullName string) (owner, name string) {
	for i, c := range fullName {
		if c == '/' {
			return fullName[:i], fullName[i+1:]
		}
	}
	return fullName, ""
}
