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
	store     store.Store
	githubSvc *GitHubService
	logger    *zap.Logger
}

// NewFeedbackService 创建 FeedbackService 实例
func NewFeedbackService(store store.Store, githubSvc *GitHubService, logger *zap.Logger) *FeedbackService {
	return &FeedbackService{
		store:     store,
		githubSvc: githubSvc,
		logger:    logger,
	}
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

	// 为每个 issue 创建反馈记录
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

	s.logger.Info("Feedback recorded",
		zap.String("repo", event.Repository.FullName),
		zap.Int("pr_number", event.Issue.Number),
		zap.Int("issues_count", len(reviewResult.Issues)),
		zap.String("reporter", event.Comment.User.Login),
	)

	// 回复确认评论
	reply := "✅ 已记录反馈，感谢您的反馈！我们会持续改进审查质量。"
	if err := s.githubSvc.CreatePRComment(ctx, event.Repository.FullName, event.Issue.Number, reply); err != nil {
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
