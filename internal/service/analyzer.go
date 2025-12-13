package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"code-sentinel/internal/model"
	"code-sentinel/internal/store"
	"code-sentinel/pkg/diff"
	"code-sentinel/pkg/prompt"

	"go.uber.org/zap"
)

type AnalyzerService struct {
	githubSvc     *GitHubService
	llmSvc        *LLMService
	store         store.Store
	logger        *zap.Logger
	builder       *prompt.Builder
	defaultLLMCfg LLMConfig
	defaultGHCfg  GitHubConfig
}

// LLMConfig 用于创建仓库级 LLM 客户端
type LLMConfig struct {
	Provider  string
	APIKey    string
	Model     string
	BaseURL   string
	Timeout   int
	MaxTokens int
}

// GitHubConfig 用于创建仓库级 GitHub 客户端
type GitHubConfig struct {
	Token   string
	BaseURL string
}

func NewAnalyzerService(githubSvc *GitHubService, llmSvc *LLMService, store store.Store, logger *zap.Logger, defaultLLMCfg LLMConfig, defaultGHCfg GitHubConfig) *AnalyzerService {
	return &AnalyzerService{
		githubSvc:     githubSvc,
		llmSvc:        llmSvc,
		store:         store,
		logger:        logger,
		builder:       prompt.NewBuilder(),
		defaultLLMCfg: defaultLLMCfg,
		defaultGHCfg:  defaultGHCfg,
	}
}

func (s *AnalyzerService) AnalyzePR(ctx context.Context, event *model.PullRequestEvent) error {
	repoFullName := event.Repository.FullName
	prNumber := event.Number

	s.logger.Info("Starting PR analysis",
		zap.String("repo", repoFullName),
		zap.Int("pr_number", prNumber),
		zap.String("action", event.Action),
	)

	// 1. 加载仓库配置
	config := s.loadRepoConfig(ctx, repoFullName)

	// 2. 检查是否启用自动审查
	if !config.AutoReview {
		s.logger.Info("Auto review disabled for repo", zap.String("repo", repoFullName))
		return nil
	}

	// 获取仓库级的 LLM 和 GitHub 服务（如果有自定义配置）
	llmSvc := s.getLLMService(config)
	githubSvc := s.getGitHubService(config)

	// 3. 创建审查记录
	review := &model.Review{
		RepoFullName: repoFullName,
		PRNumber:     prNumber,
		PRTitle:      event.PullRequest.Title,
		PRAuthor:     event.PullRequest.User.Login,
		CommitSHA:    event.PullRequest.Head.SHA,
		Status:       model.ReviewStatusPending,
	}

	if err := s.store.CreateReview(ctx, review); err != nil {
		s.logger.Error("Failed to create review record", zap.Error(err))
		return err
	}

	review.Status = model.ReviewStatusRunning
	s.store.UpdateReview(ctx, review)

	startTime := time.Now()

	// 4. 获取 PR Diff
	diffContent, err := githubSvc.GetPRDiff(ctx, repoFullName, prNumber)
	if err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	changes, err := diff.ParseDiff(diffContent)
	if err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	// 5. 应用过滤规则
	changes = s.applyFilters(changes, config)

	if len(changes) == 0 {
		s.logger.Info("No reviewable changes after filtering",
			zap.String("repo", repoFullName),
			zap.Int("pr_number", prNumber),
		)
		review.Status = model.ReviewStatusSkipped
		review.Result = "No reviewable changes after filtering"
		s.store.UpdateReview(ctx, review)
		return nil
	}

	// 6. 检查 Diff 行数限制
	totalLines := s.countDiffLines(changes)
	if config.MaxDiffLines > 0 && totalLines > config.MaxDiffLines {
		s.logger.Info("Diff too large, skipping",
			zap.String("repo", repoFullName),
			zap.Int("pr_number", prNumber),
			zap.Int("total_lines", totalLines),
			zap.Int("max_lines", config.MaxDiffLines),
		)
		review.Status = model.ReviewStatusSkipped
		review.Result = fmt.Sprintf("Diff too large: %d lines (max %d)", totalLines, config.MaxDiffLines)
		s.store.UpdateReview(ctx, review)
		return nil
	}

	// 7. 构建提示词（使用配置）
	promptConfig := &prompt.ReviewConfig{
		Languages:    config.Languages,
		ReviewFocus:  config.ReviewFocus,
		CustomPrompt: config.SystemPrompt,
	}
	systemPrompt, userPrompt, err := s.builder.BuildWithConfig(changes, promptConfig)
	if err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	// 8. 调用 LLM
	result, tokenUsed, err := llmSvc.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	duration := time.Since(startTime)

	// 9. 解析 JSON 响应
	reviewResult := s.parseReviewResult(result)
	reviewResult.Model = llmSvc.GetModel()
	reviewResult.Duration = duration.Milliseconds()

	// 10. 按最小严重程度过滤
	reviewResult.Issues = s.filterBySeverity(reviewResult.Issues, config.MinSeverity)

	// 11. 格式化评论
	comment := s.formatCommentFromResult(reviewResult, tokenUsed, duration, len(changes))
	if err := githubSvc.CreatePRComment(ctx, repoFullName, prNumber, comment); err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	// 12. 更新审查记录
	review.Status = model.ReviewStatusCompleted
	review.TokenUsed = tokenUsed
	review.DurationMs = duration.Milliseconds()

	resultJSON, _ := json.Marshal(reviewResult)
	review.Result = string(resultJSON)

	s.store.UpdateReview(ctx, review)

	// 13. 更新仓库统计
	s.updateRepoStats(ctx, repoFullName)

	s.logger.Info("PR analysis completed",
		zap.String("repo", repoFullName),
		zap.Int("pr_number", prNumber),
		zap.Int("token_used", tokenUsed),
		zap.Duration("duration", duration),
		zap.Int("issues_count", len(reviewResult.Issues)),
	)

	return nil
}

// getLLMService 获取 LLM 服务（优先使用仓库级配置）
func (s *AnalyzerService) getLLMService(config *model.ReviewConfig) *LLMService {
	// 如果仓库有自定义 LLM 配置，创建新的 LLM 客户端
	if config.LLMAPIKey != "" {
		cfg := s.defaultLLMCfg
		cfg.APIKey = config.LLMAPIKey
		if config.LLMBaseURL != "" {
			cfg.BaseURL = config.LLMBaseURL
		}
		if config.LLMProvider != "" {
			cfg.Provider = config.LLMProvider
		}
		if config.Model != "" {
			cfg.Model = config.Model
		}
		if config.MaxTokens > 0 {
			cfg.MaxTokens = config.MaxTokens
		}

		s.logger.Info("Using repo-level LLM config",
			zap.String("provider", cfg.Provider),
			zap.String("model", cfg.Model),
		)

		return NewLLMServiceWithConfig(cfg, s.logger)
	}

	// 使用默认 LLM 服务
	return s.llmSvc
}

// getGitHubService 获取 GitHub 服务（优先使用仓库级配置）
func (s *AnalyzerService) getGitHubService(config *model.ReviewConfig) *GitHubService {
	// 如果仓库有自定义 GitHub Token，创建新的 GitHub 客户端
	if config.GitHubToken != "" {
		cfg := s.defaultGHCfg
		cfg.Token = config.GitHubToken

		s.logger.Info("Using repo-level GitHub token")

		return NewGitHubServiceWithConfig(cfg, s.logger)
	}

	// 使用默认 GitHub 服务
	return s.githubSvc
}

// loadRepoConfig 加载仓库配置
func (s *AnalyzerService) loadRepoConfig(ctx context.Context, repoFullName string) *model.ReviewConfig {
	repo, err := s.store.GetRepoByFullName(ctx, repoFullName)
	if err != nil {
		s.logger.Debug("Repo not found, using default config", zap.String("repo", repoFullName))
		return s.getDefaultConfig()
	}

	if !repo.Enabled {
		return &model.ReviewConfig{AutoReview: false}
	}

	if repo.Config == "" {
		return s.getDefaultConfig()
	}

	var config model.ReviewConfig
	if err := json.Unmarshal([]byte(repo.Config), &config); err != nil {
		s.logger.Warn("Failed to parse repo config, using default",
			zap.String("repo", repoFullName),
			zap.Error(err),
		)
		return s.getDefaultConfig()
	}

	return &config
}

// getDefaultConfig 获取默认配置
func (s *AnalyzerService) getDefaultConfig() *model.ReviewConfig {
	return &model.ReviewConfig{
		LLMProvider:  "openai",
		Model:        "gpt-4-turbo",
		ReviewFocus:  []string{"security", "performance", "logic"},
		MinSeverity:  "P2",
		Languages:    []string{"go", "python", "javascript"},
		MaxDiffLines: 10000,
		AutoReview:   true,
	}
}

// applyFilters 应用过滤规则
func (s *AnalyzerService) applyFilters(changes []diff.FileChange, config *model.ReviewConfig) []diff.FileChange {
	var filtered []diff.FileChange

	for _, change := range changes {
		// 检查文件是否被忽略
		if s.shouldIgnoreFile(change.Filename, config.IgnoreFiles) {
			continue
		}

		// 检查语言是否支持
		if len(config.Languages) > 0 && !s.isLanguageSupported(change.Language, config.Languages) {
			continue
		}

		filtered = append(filtered, change)
	}

	return filtered
}

// shouldIgnoreFile 检查文件是否应被忽略
func (s *AnalyzerService) shouldIgnoreFile(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchGlob(filename, pattern) {
			return true
		}
	}
	return false
}

// isLanguageSupported 检查语言是否支持
func (s *AnalyzerService) isLanguageSupported(lang string, languages []string) bool {
	for _, l := range languages {
		if l == lang {
			return true
		}
	}
	return false
}

// countDiffLines 统计 Diff 行数
func (s *AnalyzerService) countDiffLines(changes []diff.FileChange) int {
	total := 0
	for _, c := range changes {
		total += len(c.Additions) + len(c.Deletions)
	}
	return total
}

// parseReviewResult 解析 JSON 响应
func (s *AnalyzerService) parseReviewResult(result string) *model.ReviewResult {
	// 清理可能的 markdown 包裹
	cleaned := result
	if len(cleaned) > 7 && cleaned[:7] == "```json" {
		cleaned = cleaned[7:]
	}
	if len(cleaned) > 3 && cleaned[:3] == "```" {
		cleaned = cleaned[3:]
	}
	if len(cleaned) > 3 && cleaned[len(cleaned)-3:] == "```" {
		cleaned = cleaned[:len(cleaned)-3]
	}

	var reviewResult model.ReviewResult
	if err := json.Unmarshal([]byte(cleaned), &reviewResult); err != nil {
		// JSON 解析失败，降级为纯文本
		s.logger.Debug("Failed to parse JSON result, using raw text", zap.Error(err))
		return &model.ReviewResult{
			Summary: result,
			Issues:  []model.ReviewIssue{},
		}
	}

	return &reviewResult
}

// filterBySeverity 按最小严重程度过滤
func (s *AnalyzerService) filterBySeverity(issues []model.ReviewIssue, minSeverity string) []model.ReviewIssue {
	if minSeverity == "" || minSeverity == "P2" {
		return issues
	}

	severityOrder := map[string]int{"P0": 0, "P1": 1, "P2": 2}
	minOrder, ok := severityOrder[minSeverity]
	if !ok {
		return issues
	}

	var filtered []model.ReviewIssue
	for _, issue := range issues {
		if order, ok := severityOrder[issue.Severity]; ok && order <= minOrder {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

// updateRepoStats 更新仓库统计
func (s *AnalyzerService) updateRepoStats(ctx context.Context, repoFullName string) {
	repo, err := s.store.GetRepoByFullName(ctx, repoFullName)
	if err != nil {
		return
	}

	now := time.Now()
	repo.LastReviewAt = &now
	repo.ReviewCount++

	s.store.UpdateRepo(ctx, repo)
}

// formatCommentFromResult 从结构化结果格式化评论
func (s *AnalyzerService) formatCommentFromResult(result *model.ReviewResult, tokenUsed int, duration time.Duration, fileCount int) string {
	var issuesText string
	if len(result.Issues) == 0 {
		issuesText = "✅ " + result.Summary
	} else {
		issuesText = fmt.Sprintf("**总结**：%s\n\n", result.Summary)
		for _, issue := range result.Issues {
			icon := "🟢"
			if issue.Severity == "P0" {
				icon = "🔴"
			} else if issue.Severity == "P1" {
				icon = "🟡"
			}
			issuesText += fmt.Sprintf("### %s [%s] %s\n", icon, issue.Severity, issue.Title)
			issuesText += fmt.Sprintf("**文件**：`%s:%d`\n", issue.File, issue.Line)
			issuesText += fmt.Sprintf("**问题**：%s\n", issue.Description)
			issuesText += fmt.Sprintf("**建议**：%s\n\n", issue.Suggestion)
		}
	}

	return fmt.Sprintf(`## 🤖 Code-Sentinel 代码审查报告

**审查时间**：%s
**审查模型**：%s
**变更文件**：%d 个文件
**Token 消耗**：%d
**耗时**：%.2f 秒

---

%s

---

> 💡 如有误报，请回复 `+"`/false`"+` 标记
> 📚 Powered by [Code-Sentinel](https://github.com/code-sentinel)
`,
		time.Now().Format("2006-01-02 15:04:05"),
		result.Model,
		fileCount,
		tokenUsed,
		duration.Seconds(),
		issuesText,
	)
}

// matchGlob 简单的 glob 匹配
func matchGlob(name, pattern string) bool {
	// 简单实现：支持 * 通配符
	if pattern == "*" {
		return true
	}
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(name) >= len(prefix) && name[:len(prefix)] == prefix
	}
	return name == pattern
}

func (s *AnalyzerService) updateReviewFailed(ctx context.Context, review *model.Review, err error) {
	review.Status = model.ReviewStatusFailed
	review.ErrorMsg = err.Error()
	s.store.UpdateReview(ctx, review)
	s.logger.Error("PR analysis failed",
		zap.String("repo", review.RepoFullName),
		zap.Int("pr_number", review.PRNumber),
		zap.Error(err),
	)
}

func (s *AnalyzerService) formatComment(result string, tokenUsed int, duration time.Duration, fileCount int) string {
	return fmt.Sprintf(`## 🤖 Code-Sentinel 代码审查报告

**审查时间**：%s
**审查模型**：%s
**变更文件**：%d 个文件
**Token 消耗**：%d
**耗时**：%.2f 秒

---

%s

---

> 💡 如有误报，请回复 `+"`/false`"+` 标记
> 📚 Powered by [Code-Sentinel](https://github.com/code-sentinel)
`,
		time.Now().Format("2006-01-02 15:04:05"),
		s.llmSvc.GetModel(),
		fileCount,
		tokenUsed,
		duration.Seconds(),
		result,
	)
}

func (s *AnalyzerService) GetStore() store.Store {
	return s.store
}
