package store

import (
	"context"

	"code-sentinel/internal/model"
)

type Store interface {
	// Repo
	CreateRepo(ctx context.Context, repo *model.Repo) error
	GetRepo(ctx context.Context, id uint) (*model.Repo, error)
	GetRepoByFullName(ctx context.Context, fullName string) (*model.Repo, error)
	ListRepos(ctx context.Context, page, pageSize int) ([]model.Repo, int64, error)
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
	ListReviews(ctx context.Context, repoFullName string, page, pageSize int) ([]model.Review, int64, error)

	// Health
	Ping(ctx context.Context) error
}
