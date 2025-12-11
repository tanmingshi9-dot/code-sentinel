# 📐 Code-Sentinel MVP 技术设计文档

| 文档版本 | V1.0 |
| :--- | :--- |
| **对应需求** | MVP需求文档 V1.0 |
| **技术栈** | Go 1.21+, Gin, SQLite, Docker |

---

## 1. 技术选型

### 1.1 核心技术栈

| 组件 | 技术选型 | 选型理由 |
|------|----------|----------|
| **语言** | Go 1.21+ | 高性能、部署简单、单二进制 |
| **Web 框架** | Gin | 轻量、高性能、生态成熟 |
| **数据库** | SQLite | 零配置、单文件、MVP 足够 |
| **ORM** | GORM | Go 生态主流、支持 SQLite |
| **配置** | Viper | 支持多格式、环境变量 |
| **日志** | Zap | 高性能结构化日志 |
| **HTTP 客户端** | Resty | 简洁易用、支持重试 |

### 1.2 外部依赖

| 服务 | 用途 | 备注 |
|------|------|------|
| **GitHub API** | 获取 PR diff、发布评论 | 需要 GitHub App 或 PAT |
| **OpenAI API** | LLM 代码审查 | 支持兼容接口（Azure、Ollama） |

---

## 2. 系统架构

### 2.1 整体架构

```
                                    ┌─────────────────┐
                                    │     GitHub      │
                                    │   (Webhook)     │
                                    └────────┬────────┘
                                             │ POST /webhook/github
                                             ▼
┌────────────────────────────────────────────────────────────────────┐
│                         Code-Sentinel MVP                          │
│                                                                    │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐         │
│  │   Router    │────→│  Handlers   │────→│  Services   │         │
│  │   (Gin)     │     │             │     │             │         │
│  └─────────────┘     └─────────────┘     └──────┬──────┘         │
│                                                  │                 │
│         ┌────────────────────────────────────────┼────────────┐   │
│         │                                        │            │   │
│         ▼                                        ▼            ▼   │
│  ┌─────────────┐                          ┌───────────┐ ┌───────┐│
│  │   Store     │                          │ GitHub    │ │  LLM  ││
│  │  (SQLite)   │                          │ Client    │ │Client ││
│  └─────────────┘                          └───────────┘ └───────┘│
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

### 2.2 分层架构

```
┌──────────────────────────────────────────────────────────────┐
│                      Presentation Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   Webhook    │  │   API        │  │   Web UI     │       │
│  │   Handler    │  │   Handler    │  │   (可选)     │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
├──────────────────────────────────────────────────────────────┤
│                       Service Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   Analyzer   │  │   Reviewer   │  │   Config     │       │
│  │   Service    │  │   Service    │  │   Service    │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
├──────────────────────────────────────────────────────────────┤
│                      Integration Layer                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   GitHub     │  │   LLM        │  │   Diff       │       │
│  │   Client     │  │   Client     │  │   Parser     │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
├──────────────────────────────────────────────────────────────┤
│                      Persistence Layer                        │
│  ┌──────────────────────────────────────────────────┐       │
│  │                    SQLite Store                   │       │
│  │   repos | configs | reviews                       │       │
│  └──────────────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────────────┘
```

---

## 3. 核心模块设计

### 3.1 Webhook Handler

**职责**：接收 GitHub Webhook，验证签名，分发事件

```go
// internal/handler/webhook.go

type WebhookHandler struct {
    analyzerSvc *service.AnalyzerService
    store       store.Store
}

// HandleGitHubWebhook 处理 GitHub Webhook
func (h *WebhookHandler) HandleGitHubWebhook(c *gin.Context) {
    // 1. 验证签名
    signature := c.GetHeader("X-Hub-Signature-256")
    if !h.verifySignature(c.Request.Body, signature) {
        c.JSON(401, gin.H{"error": "invalid signature"})
        return
    }
    
    // 2. 解析事件类型
    eventType := c.GetHeader("X-GitHub-Event")
    if eventType != "pull_request" {
        c.JSON(200, gin.H{"status": "ignored", "event": eventType})
        return
    }
    
    // 3. 解析 Payload
    var payload PullRequestEvent
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(400, gin.H{"error": "invalid payload"})
        return
    }
    
    // 4. 过滤 Action（只处理 opened 和 synchronize）
    if payload.Action != "opened" && payload.Action != "synchronize" {
        c.JSON(200, gin.H{"status": "ignored", "action": payload.Action})
        return
    }
    
    // 5. 异步处理审查（避免 Webhook 超时）
    go h.analyzerSvc.AnalyzePR(context.Background(), &payload)
    
    c.JSON(200, gin.H{"status": "processing"})
}

// verifySignature 验证 GitHub Webhook 签名
func (h *WebhookHandler) verifySignature(body []byte, signature string) bool {
    mac := hmac.New(sha256.New, []byte(h.webhookSecret))
    mac.Write(body)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

**数据结构**：

```go
// internal/model/github.go

type PullRequestEvent struct {
    Action      string      `json:"action"`
    Number      int         `json:"number"`
    PullRequest PullRequest `json:"pull_request"`
    Repository  Repository  `json:"repository"`
}

type PullRequest struct {
    Number    int    `json:"number"`
    Title     string `json:"title"`
    Body      string `json:"body"`
    State     string `json:"state"`
    DiffURL   string `json:"diff_url"`
    User      User   `json:"user"`
    Head      Ref    `json:"head"`
    Base      Ref    `json:"base"`
}

type Repository struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    FullName string `json:"full_name"`
    Owner    User   `json:"owner"`
}
```

---

### 3.2 Analyzer Service

**职责**：协调整个审查流程

```go
// internal/service/analyzer.go

type AnalyzerService struct {
    githubClient *github.Client
    llmClient    *llm.Client
    store        store.Store
    promptTpl    *template.Template
}

// AnalyzePR 分析 PR
func (s *AnalyzerService) AnalyzePR(ctx context.Context, event *model.PullRequestEvent) error {
    // 1. 创建审查记录
    review := &model.Review{
        RepoFullName: event.Repository.FullName,
        PRNumber:     event.Number,
        CommitSHA:    event.PullRequest.Head.SHA,
        Status:       model.ReviewStatusPending,
    }
    if err := s.store.CreateReview(ctx, review); err != nil {
        return err
    }
    
    // 2. 获取 PR Diff
    diff, err := s.githubClient.GetPRDiff(ctx, event.Repository.FullName, event.Number)
    if err != nil {
        s.updateReviewStatus(ctx, review.ID, model.ReviewStatusFailed, err.Error())
        return err
    }
    
    // 3. 解析 Diff，提取变更
    changes, err := s.parseDiff(diff)
    if err != nil {
        s.updateReviewStatus(ctx, review.ID, model.ReviewStatusFailed, err.Error())
        return err
    }
    
    // 4. 过滤忽略的文件
    changes = s.filterIgnoredFiles(changes)
    if len(changes) == 0 {
        s.updateReviewStatus(ctx, review.ID, model.ReviewStatusSkipped, "no reviewable changes")
        return nil
    }
    
    // 5. 组装 Prompt
    prompt := s.buildPrompt(changes)
    
    // 6. 调用 LLM
    startTime := time.Now()
    result, tokenUsed, err := s.llmClient.Chat(ctx, prompt)
    duration := time.Since(startTime)
    
    if err != nil {
        s.updateReviewStatus(ctx, review.ID, model.ReviewStatusFailed, err.Error())
        return err
    }
    
    // 7. 解析 AI 响应
    reviewResult := s.parseAIResponse(result)
    
    // 8. 发布 GitHub 评论
    comment := s.formatComment(reviewResult, tokenUsed, duration)
    if err := s.githubClient.CreatePRComment(ctx, event.Repository.FullName, event.Number, comment); err != nil {
        s.updateReviewStatus(ctx, review.ID, model.ReviewStatusFailed, err.Error())
        return err
    }
    
    // 9. 更新审查记录
    review.Status = model.ReviewStatusCompleted
    review.Result = reviewResult
    review.TokenUsed = tokenUsed
    review.DurationMs = duration.Milliseconds()
    s.store.UpdateReview(ctx, review)
    
    return nil
}
```

---

### 3.3 GitHub Client

**职责**：封装 GitHub API 调用

```go
// internal/integration/github/client.go

type Client struct {
    httpClient *resty.Client
    token      string
    baseURL    string
}

func NewClient(token string) *Client {
    client := resty.New().
        SetBaseURL("https://api.github.com").
        SetHeader("Authorization", "Bearer "+token).
        SetHeader("Accept", "application/vnd.github.v3+json").
        SetTimeout(30 * time.Second).
        SetRetryCount(3)
    
    return &Client{
        httpClient: client,
        token:      token,
        baseURL:    "https://api.github.com",
    }
}

// GetPRDiff 获取 PR 的 diff 内容
func (c *Client) GetPRDiff(ctx context.Context, repoFullName string, prNumber int) (string, error) {
    resp, err := c.httpClient.R().
        SetContext(ctx).
        SetHeader("Accept", "application/vnd.github.v3.diff").
        Get(fmt.Sprintf("/repos/%s/pulls/%d", repoFullName, prNumber))
    
    if err != nil {
        return "", fmt.Errorf("failed to get PR diff: %w", err)
    }
    
    if resp.StatusCode() != 200 {
        return "", fmt.Errorf("GitHub API error: %d %s", resp.StatusCode(), resp.String())
    }
    
    return resp.String(), nil
}

// CreatePRComment 在 PR 上创建评论
func (c *Client) CreatePRComment(ctx context.Context, repoFullName string, prNumber int, body string) error {
    resp, err := c.httpClient.R().
        SetContext(ctx).
        SetBody(map[string]string{"body": body}).
        Post(fmt.Sprintf("/repos/%s/issues/%d/comments", repoFullName, prNumber))
    
    if err != nil {
        return fmt.Errorf("failed to create comment: %w", err)
    }
    
    if resp.StatusCode() != 201 {
        return fmt.Errorf("GitHub API error: %d %s", resp.StatusCode(), resp.String())
    }
    
    return nil
}
```

---

### 3.4 LLM Client

**职责**：封装 LLM API 调用，支持多种后端

```go
// internal/integration/llm/client.go

type Client struct {
    httpClient *resty.Client
    config     Config
}

type Config struct {
    Provider  string // openai, azure, ollama
    APIKey    string
    Model     string
    BaseURL   string
    Timeout   time.Duration
    MaxTokens int
}

func NewClient(cfg Config) *Client {
    client := resty.New().
        SetBaseURL(cfg.BaseURL).
        SetHeader("Authorization", "Bearer "+cfg.APIKey).
        SetHeader("Content-Type", "application/json").
        SetTimeout(cfg.Timeout)
    
    return &Client{
        httpClient: client,
        config:     cfg,
    }
}

// Chat 发送聊天请求
func (c *Client) Chat(ctx context.Context, prompt string) (string, int, error) {
    req := ChatRequest{
        Model: c.config.Model,
        Messages: []Message{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: prompt},
        },
        MaxTokens:   c.config.MaxTokens,
        Temperature: 0.3, // 降低随机性，提高一致性
    }
    
    var resp ChatResponse
    _, err := c.httpClient.R().
        SetContext(ctx).
        SetBody(req).
        SetResult(&resp).
        Post("/chat/completions")
    
    if err != nil {
        return "", 0, fmt.Errorf("LLM API error: %w", err)
    }
    
    if len(resp.Choices) == 0 {
        return "", 0, fmt.Errorf("LLM returned empty response")
    }
    
    return resp.Choices[0].Message.Content, resp.Usage.TotalTokens, nil
}

type ChatRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatResponse struct {
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

type Choice struct {
    Message Message `json:"message"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

---

### 3.5 Diff Parser

**职责**：解析 Git Diff，提取变更内容

```go
// pkg/diff/parser.go

type FileChange struct {
    Filename    string
    Language    string
    Additions   []Line
    Deletions   []Line
    OldPath     string
    NewPath     string
}

type Line struct {
    Number  int
    Content string
}

// ParseDiff 解析 unified diff 格式
func ParseDiff(diffContent string) ([]FileChange, error) {
    var changes []FileChange
    
    // 按文件分割
    fileDiffs := splitByFile(diffContent)
    
    for _, fileDiff := range fileDiffs {
        change := FileChange{}
        
        // 解析文件名
        change.Filename = extractFilename(fileDiff)
        change.Language = detectLanguage(change.Filename)
        
        // 解析变更行
        lines := strings.Split(fileDiff, "\n")
        currentLine := 0
        
        for _, line := range lines {
            if strings.HasPrefix(line, "@@") {
                // 解析 hunk header: @@ -start,count +start,count @@
                currentLine = parseHunkHeader(line)
                continue
            }
            
            if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
                change.Additions = append(change.Additions, Line{
                    Number:  currentLine,
                    Content: strings.TrimPrefix(line, "+"),
                })
                currentLine++
            } else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
                change.Deletions = append(change.Deletions, Line{
                    Number:  currentLine,
                    Content: strings.TrimPrefix(line, "-"),
                })
            } else if !strings.HasPrefix(line, "\\") {
                currentLine++
            }
        }
        
        changes = append(changes, change)
    }
    
    return changes, nil
}

// detectLanguage 根据文件扩展名检测语言
func detectLanguage(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".go":
        return "go"
    case ".java":
        return "java"
    case ".py":
        return "python"
    case ".js":
        return "javascript"
    case ".ts":
        return "typescript"
    case ".rs":
        return "rust"
    case ".c", ".h":
        return "c"
    case ".cpp", ".cc", ".hpp":
        return "cpp"
    default:
        return "unknown"
    }
}
```

---

### 3.6 Prompt Template

**职责**：组装 LLM Prompt

```go
// pkg/prompt/template.go

const SystemPrompt = `你是一个资深的代码审查专家，拥有 10 年以上的软件开发经验。
你的任务是审查代码变更，找出潜在的问题并提供改进建议。

审查重点：
1. **Bug 和逻辑错误**：空指针、数组越界、逻辑漏洞
2. **性能问题**：循环内查库、N+1 查询、不必要的内存分配
3. **安全隐患**：SQL 注入、XSS、敏感信息泄露
4. **代码质量**：命名规范、代码重复、过长函数

输出格式要求：
- 使用中文回复
- 按严重程度分类（严重/警告/建议）
- 指出具体的文件名和行号
- 提供具体的修复建议`

const ReviewPromptTemplate = `请审查以下代码变更：

## 变更概览
- 文件数量：{{.FileCount}}
- 新增行数：{{.AdditionCount}}
- 删除行数：{{.DeletionCount}}

## 变更详情
{{range .Changes}}
### 文件：{{.Filename}} ({{.Language}})

**新增代码：**
{{range .Additions}}
第 {{.Number}} 行：{{.Content}}
{{end}}

**删除代码：**
{{range .Deletions}}
第 {{.Number}} 行：{{.Content}}
{{end}}
---
{{end}}

请按以下格式输出审查结果：

## 🔴 严重问题
（如果没有则写"无"）

## 🟡 警告
（如果没有则写"无"）

## 🟢 建议
（如果没有则写"无"）

## 📝 总结
（简要总结代码质量和主要问题）`

type PromptBuilder struct {
    template *template.Template
}

func NewPromptBuilder() *PromptBuilder {
    tpl := template.Must(template.New("review").Parse(ReviewPromptTemplate))
    return &PromptBuilder{template: tpl}
}

func (b *PromptBuilder) Build(changes []diff.FileChange) (string, error) {
    data := struct {
        FileCount     int
        AdditionCount int
        DeletionCount int
        Changes       []diff.FileChange
    }{
        FileCount: len(changes),
        Changes:   changes,
    }
    
    for _, c := range changes {
        data.AdditionCount += len(c.Additions)
        data.DeletionCount += len(c.Deletions)
    }
    
    var buf bytes.Buffer
    if err := b.template.Execute(&buf, data); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}
```

---

### 3.7 Store (SQLite)

**职责**：数据持久化

```go
// internal/store/sqlite.go

type SQLiteStore struct {
    db *gorm.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, err
    }
    
    // 自动迁移
    if err := db.AutoMigrate(&model.Repo{}, &model.Config{}, &model.Review{}); err != nil {
        return nil, err
    }
    
    return &SQLiteStore{db: db}, nil
}

// Repo CRUD
func (s *SQLiteStore) CreateRepo(ctx context.Context, repo *model.Repo) error {
    return s.db.WithContext(ctx).Create(repo).Error
}

func (s *SQLiteStore) GetRepoByFullName(ctx context.Context, fullName string) (*model.Repo, error) {
    var repo model.Repo
    err := s.db.WithContext(ctx).Where("full_name = ?", fullName).First(&repo).Error
    if err != nil {
        return nil, err
    }
    return &repo, nil
}

// Config CRUD
func (s *SQLiteStore) GetConfig(ctx context.Context, key string) (string, error) {
    var config model.Config
    err := s.db.WithContext(ctx).Where("key = ?", key).First(&config).Error
    if err != nil {
        return "", err
    }
    return config.Value, nil
}

func (s *SQLiteStore) SetConfig(ctx context.Context, key, value string) error {
    return s.db.WithContext(ctx).Save(&model.Config{Key: key, Value: value}).Error
}

// Review CRUD
func (s *SQLiteStore) CreateReview(ctx context.Context, review *model.Review) error {
    return s.db.WithContext(ctx).Create(review).Error
}

func (s *SQLiteStore) UpdateReview(ctx context.Context, review *model.Review) error {
    return s.db.WithContext(ctx).Save(review).Error
}

func (s *SQLiteStore) ListReviews(ctx context.Context, repoFullName string, page, size int) ([]model.Review, int64, error) {
    var reviews []model.Review
    var total int64
    
    query := s.db.WithContext(ctx).Model(&model.Review{})
    if repoFullName != "" {
        query = query.Where("repo_full_name = ?", repoFullName)
    }
    
    query.Count(&total)
    err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&reviews).Error
    
    return reviews, total, err
}
```

---

## 4. 数据模型

```go
// internal/model/model.go

type Repo struct {
    ID            uint      `gorm:"primaryKey"`
    FullName      string    `gorm:"uniqueIndex;size:200"` // owner/repo
    WebhookSecret string    `gorm:"size:255"`
    Enabled       bool      `gorm:"default:true"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type Config struct {
    ID          uint   `gorm:"primaryKey"`
    Key         string `gorm:"uniqueIndex;size:100"`
    Value       string `gorm:"type:text"`
    Description string `gorm:"size:255"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Review struct {
    ID           uint         `gorm:"primaryKey"`
    RepoFullName string       `gorm:"index;size:200"`
    PRNumber     int          `gorm:"index"`
    CommitSHA    string       `gorm:"size:40"`
    Status       ReviewStatus `gorm:"size:20"`
    Result       string       `gorm:"type:text"` // JSON
    TokenUsed    int
    DurationMs   int64
    ErrorMsg     string `gorm:"type:text"`
    CreatedAt    time.Time
}

type ReviewStatus string

const (
    ReviewStatusPending   ReviewStatus = "pending"
    ReviewStatusCompleted ReviewStatus = "completed"
    ReviewStatusFailed    ReviewStatus = "failed"
    ReviewStatusSkipped   ReviewStatus = "skipped"
)
```

---

## 5. 配置设计

```yaml
# configs/config.yaml

server:
  host: 0.0.0.0
  port: 8080
  mode: release  # debug / release

database:
  driver: sqlite
  path: ./data/sentinel.db

github:
  # 方式一：GitHub App（推荐）
  app_id: 123456
  installation_id: 789012
  private_key_path: ./configs/github-app.pem
  
  # 方式二：Personal Access Token
  # token: ghp_xxxxxxxxxxxx
  
  webhook_secret: your-webhook-secret

llm:
  provider: openai  # openai / azure / ollama
  api_key: sk-xxxxxxxxxxxx
  model: gpt-4
  base_url: https://api.openai.com/v1
  timeout: 60s
  max_tokens: 4096

review:
  # 支持的语言
  languages:
    - go
    - java
    - python
  
  # 最大 diff 行数（超过则截断）
  max_diff_lines: 500
  
  # 忽略的文件模式
  ignore_patterns:
    - "*.md"
    - "*.json"
    - "*.yaml"
    - "*.yml"
    - "go.mod"
    - "go.sum"
    - "vendor/*"
    - "node_modules/*"
    - "*_test.go"  # 可选：是否审查测试文件

log:
  level: info  # debug / info / warn / error
  format: json # json / console
  output: stdout
```

---

## 6. 错误处理

### 6.1 错误码定义

```go
// pkg/errors/errors.go

type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
}

var (
    ErrInvalidSignature  = &AppError{Code: 40101, Message: "invalid webhook signature"}
    ErrInvalidPayload    = &AppError{Code: 40001, Message: "invalid request payload"}
    ErrRepoNotFound      = &AppError{Code: 40401, Message: "repository not found"}
    ErrGitHubAPIFailed   = &AppError{Code: 50201, Message: "GitHub API request failed"}
    ErrLLMAPIFailed      = &AppError{Code: 50202, Message: "LLM API request failed"}
    ErrDatabaseFailed    = &AppError{Code: 50301, Message: "database operation failed"}
)
```

### 6.2 重试策略

```go
// LLM 调用重试
retryConfig := retry.Config{
    MaxAttempts: 3,
    InitialDelay: 1 * time.Second,
    MaxDelay: 10 * time.Second,
    Multiplier: 2,
    RetryOn: []int{429, 500, 502, 503, 504},
}

// GitHub API 重试
githubRetryConfig := retry.Config{
    MaxAttempts: 3,
    InitialDelay: 500 * time.Millisecond,
    MaxDelay: 5 * time.Second,
    Multiplier: 2,
    RetryOn: []int{500, 502, 503, 504},
}
```

---

## 7. 安全设计

### 7.1 Webhook 签名验证

```go
func verifyGitHubSignature(payload []byte, signature, secret string) bool {
    if !strings.HasPrefix(signature, "sha256=") {
        return false
    }
    
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

### 7.2 敏感配置加密

```go
// 使用 AES-256-GCM 加密敏感配置
func encryptConfig(plaintext, key string) (string, error) {
    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}
```

---

## 8. 部署架构

### 8.1 Docker 部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /sentinel ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app
COPY --from=builder /sentinel .
COPY configs/config.yaml ./configs/

EXPOSE 8080
CMD ["./sentinel"]
```

### 8.2 Docker Compose

```yaml
# docker-compose.yaml
version: '3.8'

services:
  code-sentinel:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./configs:/app/configs:ro
    environment:
      - GIN_MODE=release
      - GITHUB_APP_ID=${GITHUB_APP_ID}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

---

## 9. 监控与日志

### 9.1 健康检查接口

```go
// GET /health
func (h *Handler) Health(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "ok",
        "time":   time.Now().Format(time.RFC3339),
    })
}

// GET /ready
func (h *Handler) Ready(c *gin.Context) {
    // 检查数据库连接
    if err := h.store.Ping(); err != nil {
        c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "ready"})
}
```

### 9.2 结构化日志

```go
// 使用 Zap 记录结构化日志
logger.Info("PR review completed",
    zap.String("repo", repoFullName),
    zap.Int("pr_number", prNumber),
    zap.Int("token_used", tokenUsed),
    zap.Duration("duration", duration),
    zap.Int("issues_found", len(issues)),
)
```

---

## 10. 测试策略

### 10.1 单元测试

```go
// internal/service/analyzer_test.go

func TestParseDiff(t *testing.T) {
    diffContent := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -10,6 +10,8 @@ func main() {
     fmt.Println("Hello")
+    // New comment
+    fmt.Println("World")
 }
`
    changes, err := diff.ParseDiff(diffContent)
    assert.NoError(t, err)
    assert.Len(t, changes, 1)
    assert.Equal(t, "main.go", changes[0].Filename)
    assert.Len(t, changes[0].Additions, 2)
}
```

### 10.2 集成测试

```go
// 使用 httptest 测试 Webhook Handler
func TestWebhookHandler(t *testing.T) {
    router := setupTestRouter()
    
    payload := `{"action": "opened", "number": 1, ...}`
    signature := computeSignature(payload, "test-secret")
    
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/webhook/github", strings.NewReader(payload))
    req.Header.Set("X-GitHub-Event", "pull_request")
    req.Header.Set("X-Hub-Signature-256", signature)
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

---

## 附录：依赖清单

```go
// go.mod
module github.com/yourname/code-sentinel

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/go-resty/resty/v2 v2.10.0
    github.com/spf13/viper v1.17.0
    go.uber.org/zap v1.26.0
    gorm.io/driver/sqlite v1.5.4
    gorm.io/gorm v1.25.5
)
```
