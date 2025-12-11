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
	githubSvc *GitHubService
	llmSvc    *LLMService
	store     store.Store
	logger    *zap.Logger
	builder   *prompt.Builder
}

func NewAnalyzerService(githubSvc *GitHubService, llmSvc *LLMService, store store.Store, logger *zap.Logger) *AnalyzerService {
	return &AnalyzerService{
		githubSvc: githubSvc,
		llmSvc:    llmSvc,
		store:     store,
		logger:    logger,
		builder:   prompt.NewBuilder(),
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

	diffContent, err := s.githubSvc.GetPRDiff(ctx, repoFullName, prNumber)
	if err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	changes, err := diff.ParseDiff(diffContent)
	if err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	if len(changes) == 0 {
		s.logger.Info("No reviewable changes found",
			zap.String("repo", repoFullName),
			zap.Int("pr_number", prNumber),
		)
		review.Status = model.ReviewStatusSkipped
		review.Result = "No reviewable changes"
		s.store.UpdateReview(ctx, review)
		return nil
	}

	systemPrompt, userPrompt, err := s.builder.Build(changes)
	if err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	result, tokenUsed, err := s.llmSvc.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	duration := time.Since(startTime)

	comment := s.formatComment(result, tokenUsed, duration, len(changes))
	if err := s.githubSvc.CreatePRComment(ctx, repoFullName, prNumber, comment); err != nil {
		s.updateReviewFailed(ctx, review, err)
		return err
	}

	review.Status = model.ReviewStatusCompleted
	review.TokenUsed = tokenUsed
	review.DurationMs = duration.Milliseconds()

	resultData := model.ReviewResult{
		Summary:  result,
		Model:    s.llmSvc.GetModel(),
		Duration: duration.Milliseconds(),
	}
	resultJSON, _ := json.Marshal(resultData)
	review.Result = string(resultJSON)

	s.store.UpdateReview(ctx, review)

	s.logger.Info("PR analysis completed",
		zap.String("repo", repoFullName),
		zap.Int("pr_number", prNumber),
		zap.Int("token_used", tokenUsed),
		zap.Duration("duration", duration),
	)

	return nil
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
