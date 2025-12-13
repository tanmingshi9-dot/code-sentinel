package model

import (
	"time"

	"gorm.io/gorm"
)

type Repo struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	FullName      string `gorm:"uniqueIndex;size:200" json:"full_name"`
	Owner         string `gorm:"size:100" json:"owner"`
	Name          string `gorm:"size:100" json:"name"`
	WebhookSecret string `gorm:"size:255" json:"-"`
	Enabled       bool   `gorm:"default:true" json:"enabled"`
	// Phase 2: 新增字段
	Config       string         `gorm:"type:text" json:"config"`       // ReviewConfig JSON
	LastReviewAt *time.Time     `gorm:"index" json:"last_review_at"`   // 最后审查时间
	ReviewCount  int            `gorm:"default:0" json:"review_count"` // 累计审查次数
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// ReviewConfig 仓库审查配置（JSON 存储）
type ReviewConfig struct {
	LLMProvider  string   `json:"llm_provider"`   // openai/qwen/azure/ollama
	Model        string   `json:"model"`          // gpt-4-turbo/qwen-max
	MaxTokens    int      `json:"max_tokens"`     // 单次最大 Token
	SystemPrompt string   `json:"system_prompt"`  // 自定义系统提示词
	ReviewFocus  []string `json:"review_focus"`   // 审查重点: security/performance/logic/style
	MinSeverity  string   `json:"min_severity"`   // 最小报告级别: P0/P1/P2
	Languages    []string `json:"languages"`      // 支持语言: go/java/python
	IgnoreFiles  []string `json:"ignore_files"`   // 忽略文件: *.test.go
	MaxDiffLines int      `json:"max_diff_lines"` // 最大 Diff 行数
	AutoReview   bool     `json:"auto_review"`    // 是否自动审查

	// 仓库级 LLM 配置（可选，覆盖全局配置）
	LLMAPIKey  string `json:"llm_api_key,omitempty"`  // LLM API Key
	LLMBaseURL string `json:"llm_base_url,omitempty"` // LLM API Base URL

	// 仓库级 GitHub 配置（可选，覆盖全局配置）
	GitHubToken string `json:"github_token,omitempty"` // GitHub Personal Access Token
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
	Stats    ReviewStats   `json:"stats"`
	Score    int           `json:"score"`
	Model    string        `json:"model"`
	Duration int64         `json:"duration_ms"`
}

type ReviewIssue struct {
	Severity    string `json:"severity"` // P0/P1/P2
	Category    string `json:"category"` // security/performance/logic/style
	File        string `json:"file"`
	Line        int    `json:"line"`
	Title       string `json:"title"` // 问题标题
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
	CodeFix     string `json:"code_fix,omitempty"` // 修复代码
}

// ReviewStats 审查统计
type ReviewStats struct {
	P0Count int `json:"p0_count"`
	P1Count int `json:"p1_count"`
	P2Count int `json:"p2_count"`
}

// Feedback 误报反馈记录
type Feedback struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ReviewID        uint      `gorm:"index" json:"review_id"` // 关联的审查记录
	RepoFullName    string    `gorm:"index;size:200" json:"repo_full_name"`
	PRNumber        int       `json:"pr_number"`
	File            string    `gorm:"size:500" json:"file"`        // 文件路径
	Line            int       `json:"line"`                        // 行号
	IssueIndex      int       `json:"issue_index"`                 // 问题索引
	Severity        string    `gorm:"size:10" json:"severity"`     // P0/P1/P2
	Category        string    `gorm:"size:20" json:"category"`     // security/performance/logic/style
	Title           string    `gorm:"size:255" json:"title"`       // 问题标题
	AIContent       string    `gorm:"type:text" json:"ai_content"` // AI 原始判断
	IsFalsePositive bool      `gorm:"index;default:true" json:"is_false_positive"`
	Reason          string    `gorm:"type:text" json:"reason"`  // 用户提供的原因
	Reporter        string    `gorm:"size:100" json:"reporter"` // 反馈人
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
}
