package model

import (
	"time"

	"gorm.io/gorm"
)

type Repo struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	FullName      string         `gorm:"uniqueIndex;size:200" json:"full_name"`
	Owner         string         `gorm:"size:100" json:"owner"`
	Name          string         `gorm:"size:100" json:"name"`
	WebhookSecret string         `gorm:"size:255" json:"-"`
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Config struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"uniqueIndex;size:100" json:"key"`
	Value       string    `gorm:"type:text" json:"value"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Review struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	RepoID       uint         `gorm:"index" json:"repo_id"`
	RepoFullName string       `gorm:"index;size:200" json:"repo_full_name"`
	PRNumber     int          `gorm:"index" json:"pr_number"`
	PRTitle      string       `gorm:"size:500" json:"pr_title"`
	PRAuthor     string       `gorm:"size:100" json:"pr_author"`
	CommitSHA    string       `gorm:"size:40" json:"commit_sha"`
	Status       ReviewStatus `gorm:"size:20;index" json:"status"`
	Result       string       `gorm:"type:text" json:"result"`
	TokenUsed    int          `json:"token_used"`
	DurationMs   int64        `json:"duration_ms"`
	ErrorMsg     string       `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

type ReviewStatus string

const (
	ReviewStatusPending   ReviewStatus = "pending"
	ReviewStatusRunning   ReviewStatus = "running"
	ReviewStatusCompleted ReviewStatus = "completed"
	ReviewStatusFailed    ReviewStatus = "failed"
	ReviewStatusSkipped   ReviewStatus = "skipped"
)

type ReviewResult struct {
	Summary  string        `json:"summary"`
	Issues   []ReviewIssue `json:"issues"`
	Score    int           `json:"score"`
	Model    string        `json:"model"`
	Duration int64         `json:"duration_ms"`
}

type ReviewIssue struct {
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
}
