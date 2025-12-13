package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"code-sentinel/internal/model"
	"code-sentinel/internal/store"

	"go.uber.org/zap"
)

// FeedbackService 误报反馈服务
type FeedbackService struct {
	store        store.Store
	githubSvc    *GitHubService
	logger       *zap.Logger
	defaultGHCfg GitHubConfig
}

// NewFeedbackService 创建 FeedbackService 实例
func NewFeedbackService(store store.Store, githubSvc *GitHubService, logger *zap.Logger, defaultGHCfg GitHubConfig) *FeedbackService {
	return &FeedbackService{
		store:        store,
		githubSvc:    githubSvc,
		logger:       logger,
		defaultGHCfg: defaultGHCfg,
	}
}

// getGitHubService 获取 GitHub 服务（优先使用仓库级配置）
func (s *FeedbackService) getGitHubService(ctx context.Context, repoFullName string) *GitHubService {
	repo, err := s.store.GetRepoByFullName(ctx, repoFullName)
	if err != nil {
		return s.githubSvc
	}

	if repo.Config == "" {
		return s.githubSvc
	}

	var config model.ReviewConfig
	if err := json.Unmarshal([]byte(repo.Config), &config); err != nil {
		return s.githubSvc
	}

	if config.GitHubToken != "" {
		cfg := s.defaultGHCfg
		cfg.Token = config.GitHubToken
		return NewGitHubServiceWithConfig(cfg, s.logger)
	}

	return s.githubSvc
}

// HandleFalseCommand 处理 /false 命令
func (s *FeedbackService) HandleFalseCommand(ctx context.Context, event *model.IssueCommentEvent) error {
	content := event.Comment.Body
	if !strings.Contains(content, "/false") {
		return nil
	}

	// 提取原因
	reason := s.extractReason(content)

	// 查找关联的审查记录
	review, err := s.store.GetReviewByPR(ctx, event.Repository.FullName, event.Issue.Number)
	if err != nil {
		s.logger.Warn("No review found for PR",
			zap.String("repo", event.Repository.FullName),
			zap.Int("pr_number", event.Issue.Number),
			zap.Error(err),
		)
		return err
	}

	// 解析审查结果
	var reviewResult model.ReviewResult
	if err := json.Unmarshal([]byte(review.Result), &reviewResult); err != nil {
		s.logger.Warn("Failed to parse review result",
			zap.Uint("review_id", review.ID),
			zap.Error(err),
		)
		return err
	}

	// 创建反馈记录
	if len(reviewResult.Issues) > 0 {
		// 误报：AI 报告了问题，但用户认为不是问题
		for i, issue := range reviewResult.Issues {
			feedback := &model.Feedback{
				ReviewID:        review.ID,
				RepoFullName:    review.RepoFullName,
				PRNumber:        review.PRNumber,
				File:            issue.File,
				Line:            issue.Line,
				IssueIndex:      i,
				Severity:        issue.Severity,
				Category:        issue.Category,
				Title:           issue.Title,
				AIContent:       issue.Description,
				IsFalsePositive: true,
				Reason:          reason,
				Reporter:        event.Comment.User.Login,
			}

			if err := s.store.CreateFeedback(ctx, feedback); err != nil {
				s.logger.Error("Failed to create feedback",
					zap.Uint("review_id", review.ID),
					zap.Int("issue_index", i),
					zap.Error(err),
				)
			}
		}
	} else {
		// 漏报：AI 没报告问题，但用户认为有问题
		feedback := &model.Feedback{
			ReviewID:        review.ID,
			RepoFullName:    review.RepoFullName,
			PRNumber:        review.PRNumber,
			File:            "",
			Line:            0,
			IssueIndex:      -1,
			Severity:        "",
			Category:        "missed",
			Title:           "AI 漏报",
			AIContent:       reviewResult.Summary,
			IsFalsePositive: false, // 这是漏报，不是误报
			Reason:          reason,
			Reporter:        event.Comment.User.Login,
		}

		if err := s.store.CreateFeedback(ctx, feedback); err != nil {
			s.logger.Error("Failed to create feedback for missed issue",
				zap.Uint("review_id", review.ID),
				zap.Error(err),
			)
		}
	}

	feedbackType := "误报"
	if len(reviewResult.Issues) == 0 {
		feedbackType = "漏报"
	}

	s.logger.Info("Feedback recorded",
		zap.String("repo", event.Repository.FullName),
		zap.Int("pr_number", event.Issue.Number),
		zap.String("type", feedbackType),
		zap.Int("issues_count", len(reviewResult.Issues)),
		zap.String("reporter", event.Comment.User.Login),
	)

	// 回复确认评论（使用仓库级 GitHub Token）
	githubSvc := s.getGitHubService(ctx, event.Repository.FullName)
	reply := "✅ 已记录反馈，感谢您的反馈！我们会持续改进审查质量。"
	if err := githubSvc.CreatePRComment(ctx, event.Repository.FullName, event.Issue.Number, reply); err != nil {
		s.logger.Warn("Failed to reply feedback confirmation",
			zap.String("repo", event.Repository.FullName),
			zap.Int("pr_number", event.Issue.Number),
			zap.Error(err),
		)
	}

	return nil
}

// extractReason 提取反馈原因
func (s *FeedbackService) extractReason(content string) string {
	// 匹配 /false 后面的文本
	re := regexp.MustCompile(`/false\s+(.+)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return "未提供原因"
}

// ListFeedbacks 获取反馈列表
func (s *FeedbackService) ListFeedbacks(ctx context.Context, filter *store.FeedbackFilter, page, pageSize int) ([]model.Feedback, int64, error) {
	return s.store.ListFeedbacks(ctx, filter, page, pageSize)
}

// GetFeedback 获取反馈详情
func (s *FeedbackService) GetFeedback(ctx context.Context, id uint) (*model.Feedback, error) {
	return s.store.GetFeedback(ctx, id)
}

// GetFeedbackStats 获取反馈统计
func (s *FeedbackService) GetFeedbackStats(ctx context.Context, repoFullName, startDate, endDate string) (*store.FeedbackStats, error) {
	return s.store.GetFeedbackStats(ctx, repoFullName, startDate, endDate)
}

// CreateFeedback 创建反馈（API 调用）
func (s *FeedbackService) CreateFeedback(ctx context.Context, feedback *model.Feedback) error {
	return s.store.CreateFeedback(ctx, feedback)
}
