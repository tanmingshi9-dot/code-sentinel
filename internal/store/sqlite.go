package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"code-sentinel/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteStore struct {
	db *gorm.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.AutoMigrate(&model.Repo{}, &model.Config{}, &model.Review{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Repo methods

func (s *SQLiteStore) CreateRepo(ctx context.Context, repo *model.Repo) error {
	return s.db.WithContext(ctx).Create(repo).Error
}

func (s *SQLiteStore) GetRepo(ctx context.Context, id uint) (*model.Repo, error) {
	var repo model.Repo
	if err := s.db.WithContext(ctx).First(&repo, id).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (s *SQLiteStore) GetRepoByFullName(ctx context.Context, fullName string) (*model.Repo, error) {
	var repo model.Repo
	if err := s.db.WithContext(ctx).Where("full_name = ?", fullName).First(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (s *SQLiteStore) ListRepos(ctx context.Context, page, pageSize int) ([]model.Repo, int64, error) {
	var repos []model.Repo
	var total int64

	s.db.WithContext(ctx).Model(&model.Repo{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&repos).Error; err != nil {
		return nil, 0, err
	}

	return repos, total, nil
}

func (s *SQLiteStore) UpdateRepo(ctx context.Context, repo *model.Repo) error {
	return s.db.WithContext(ctx).Save(repo).Error
}

func (s *SQLiteStore) DeleteRepo(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&model.Repo{}, id).Error
}

// Config methods

func (s *SQLiteStore) GetConfig(ctx context.Context, key string) (*model.Config, error) {
	var config model.Config
	if err := s.db.WithContext(ctx).Where("key = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *SQLiteStore) SetConfig(ctx context.Context, key, value, description string) error {
	var config model.Config
	result := s.db.WithContext(ctx).Where("key = ?", key).First(&config)

	if result.Error == gorm.ErrRecordNotFound {
		config = model.Config{
			Key:         key,
			Value:       value,
			Description: description,
		}
		return s.db.WithContext(ctx).Create(&config).Error
	}

	config.Value = value
	if description != "" {
		config.Description = description
	}
	return s.db.WithContext(ctx).Save(&config).Error
}

func (s *SQLiteStore) ListConfigs(ctx context.Context) ([]model.Config, error) {
	var configs []model.Config
	if err := s.db.WithContext(ctx).Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// Review methods

func (s *SQLiteStore) CreateReview(ctx context.Context, review *model.Review) error {
	return s.db.WithContext(ctx).Create(review).Error
}

func (s *SQLiteStore) GetReview(ctx context.Context, id uint) (*model.Review, error) {
	var review model.Review
	if err := s.db.WithContext(ctx).First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (s *SQLiteStore) UpdateReview(ctx context.Context, review *model.Review) error {
	return s.db.WithContext(ctx).Save(review).Error
}

func (s *SQLiteStore) ListReviews(ctx context.Context, repoFullName string, page, pageSize int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64

	query := s.db.WithContext(ctx).Model(&model.Review{})
	if repoFullName != "" {
		query = query.Where("repo_full_name = ?", repoFullName)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}
