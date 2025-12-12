package store

import (
	"context"

	"code-sentinel/internal/model"
)

// ReviewFilter 审查列表筛选条件
type ReviewFilter struct {
	RepoFullName string
	Status       string
	PRNumber     int
	StartDate    string // YYYY-MM-DD
	EndDate      string // YYYY-MM-DD
}

// FeedbackFilter 反馈列表筛选条件
type FeedbackFilter struct {
	RepoFullName string
	Category     string
	Severity     string
	StartDate    string
	EndDate      string
}

type Store interface {
	// Repo
	CreateRepo(ctx context.Context, repo *model.Repo) error
	GetRepo(ctx context.Context, id uint) (*model.Repo, error)
	GetRepoByFullName(ctx context.Context, fullName string) (*model.Repo, error)
	ListRepos(ctx context.Context, page, pageSize int, search string) ([]model.Repo, int64, error)
	UpdateRepo(ctx context.Context, repo *model.Repo) error
	DeleteRepo(ctx context.Context, id uint) error

	// Config
	GetConfig(ctx context.Context, key string) (*model.Config, error)
	SetConfig(ctx context.Context, key, value, description string) error
	ListConfigs(ctx context.Context) ([]model.Config, error)

	// Review
	CreateReview(ctx context.Context, review *model.Review) error
	GetReview(ctx context.Context, id uint) (*model.Review, error)
	UpdateReview(ctx context.Context, review *model.Review) error
	ListReviews(ctx context.Context, filter *ReviewFilter, page, pageSize int) ([]model.Review, int64, error)
	GetReviewByPR(ctx context.Context, repoFullName string, prNumber int) (*model.Review, error)

	// Feedback
	CreateFeedback(ctx context.Context, feedback *model.Feedback) error
	GetFeedback(ctx context.Context, id uint) (*model.Feedback, error)
	ListFeedbacks(ctx context.Context, filter *FeedbackFilter, page, pageSize int) ([]model.Feedback, int64, error)
	GetFeedbackStats(ctx context.Context, repoFullName string, startDate, endDate string) (*FeedbackStats, error)

	// Health
	Ping(ctx context.Context) error
}

// FeedbackStats 反馈统计
type FeedbackStats struct {
	TotalFeedbacks    int            `json:"total_feedbacks"`
	FalsePositiveRate float64        `json:"false_positive_rate"`
	ByCategory        map[string]int `json:"by_category"`
	BySeverity        map[string]int `json:"by_severity"`
}
