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

	if err := db.AutoMigrate(&model.Repo{}, &model.Config{}, &model.Review{}, &model.Feedback{}); err != nil {
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

func (s *SQLiteStore) ListRepos(ctx context.Context, page, pageSize int, search string) ([]model.Repo, int64, error) {
	var repos []model.Repo
	var total int64

	query := s.db.WithContext(ctx).Model(&model.Repo{})
	if search != "" {
		query = query.Where("full_name LIKE ?", "%"+search+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("last_review_at DESC NULLS LAST, created_at DESC").Offset(offset).Limit(pageSize).Find(&repos).Error; err != nil {
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

func (s *SQLiteStore) ListReviews(ctx context.Context, filter *ReviewFilter, page, pageSize int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64

	query := s.db.WithContext(ctx).Model(&model.Review{})
	if filter != nil {
		if filter.RepoFullName != "" {
			query = query.Where("repo_full_name = ?", filter.RepoFullName)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.PRNumber > 0 {
			query = query.Where("pr_number = ?", filter.PRNumber)
		}
		if filter.StartDate != "" {
			query = query.Where("created_at >= ?", filter.StartDate+" 00:00:00")
		}
		if filter.EndDate != "" {
			query = query.Where("created_at <= ?", filter.EndDate+" 23:59:59")
		}
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

func (s *SQLiteStore) GetReviewByPR(ctx context.Context, repoFullName string, prNumber int) (*model.Review, error) {
	var review model.Review
	if err := s.db.WithContext(ctx).Where("repo_full_name = ? AND pr_number = ?", repoFullName, prNumber).Order("created_at DESC").First(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// Feedback methods

func (s *SQLiteStore) CreateFeedback(ctx context.Context, feedback *model.Feedback) error {
	return s.db.WithContext(ctx).Create(feedback).Error
}

func (s *SQLiteStore) GetFeedback(ctx context.Context, id uint) (*model.Feedback, error) {
	var feedback model.Feedback
	if err := s.db.WithContext(ctx).First(&feedback, id).Error; err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (s *SQLiteStore) ListFeedbacks(ctx context.Context, filter *FeedbackFilter, page, pageSize int) ([]model.Feedback, int64, error) {
	var feedbacks []model.Feedback
	var total int64

	query := s.db.WithContext(ctx).Model(&model.Feedback{})
	if filter != nil {
		if filter.RepoFullName != "" {
			query = query.Where("repo_full_name = ?", filter.RepoFullName)
		}
		if filter.Category != "" {
			query = query.Where("category = ?", filter.Category)
		}
		if filter.Severity != "" {
			query = query.Where("severity = ?", filter.Severity)
		}
		if filter.StartDate != "" {
			query = query.Where("created_at >= ?", filter.StartDate+" 00:00:00")
		}
		if filter.EndDate != "" {
			query = query.Where("created_at <= ?", filter.EndDate+" 23:59:59")
		}
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&feedbacks).Error; err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

func (s *SQLiteStore) GetFeedbackStats(ctx context.Context, repoFullName string, startDate, endDate string) (*FeedbackStats, error) {
	stats := &FeedbackStats{
		ByCategory: make(map[string]int),
		BySeverity: make(map[string]int),
	}

	query := s.db.WithContext(ctx).Model(&model.Feedback{})
	if repoFullName != "" {
		query = query.Where("repo_full_name = ?", repoFullName)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	// 总数
	var total int64
	query.Count(&total)
	stats.TotalFeedbacks = int(total)

	// 按 category 统计
	type categoryCount struct {
		Category string
		Count    int
	}
	var categoryStats []categoryCount
	query.Select("category, COUNT(*) as count").Group("category").Scan(&categoryStats)
	for _, c := range categoryStats {
		stats.ByCategory[c.Category] = c.Count
	}

	// 按 severity 统计
	type severityCount struct {
		Severity string
		Count    int
	}
	var severityStats []severityCount
	query.Select("severity, COUNT(*) as count").Group("severity").Scan(&severityStats)
	for _, s := range severityStats {
		stats.BySeverity[s.Severity] = s.Count
	}

	// 误报率（这里简化处理，假设所有反馈都是误报）
	if stats.TotalFeedbacks > 0 {
		stats.FalsePositiveRate = 1.0
	}

	return stats, nil
}
