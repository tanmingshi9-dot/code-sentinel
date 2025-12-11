# 📘 Code-Sentinel MVP 需求文档

| 文档版本 | V1.0 (MVP) |
| :--- | :--- |
| **项目代号** | Code-Sentinel (代码哨兵) |
| **MVP 目标** | 跑通 GitHub PR → AI 审查 → 评论回写 的完整闭环 |
| **预计工期** | 2-3 周 |
| **技术栈** | Go Gin, SQLite, Docker |

---

## 1. MVP 核心价值

**一句话描述**：当开发者提交 PR 时，AI 自动审查代码并在 GitHub 上留下评论。

### 1.1 解决的核心问题
- Senior 工程师没时间 Review 所有代码 → AI 先过一遍，筛出明显问题

### 1.2 MVP 不做什么
- ❌ 不做 RAG 知识库
- ❌ 不做效能报表
- ❌ 不做多租户
- ❌ 不做 Kafka 消息队列
- ❌ 不做 K8s 部署

---

## 2. 功能需求

### 2.1 GitHub Webhook 接收 `[必须]`

**用户故事**：作为开发者，当我创建/更新 PR 时，系统能自动触发代码审查。

**功能点**：
- 接收 GitHub `pull_request` 事件（opened, synchronize）
- 解析 PR 元数据（仓库、分支、PR 号、作者）
- 获取 PR 的 diff 内容（变更的代码）

**接口设计**：
```
POST /webhook/github
Headers: X-GitHub-Event, X-Hub-Signature-256
Body: GitHub Webhook Payload
```

**验收标准**：
- [ ] 能正确接收 GitHub Webhook 请求
- [ ] 能验证 Webhook 签名（安全）
- [ ] 能解析出 PR 的 diff 内容

---

### 2.2 增量代码审查 `[必须]`

**用户故事**：作为开发者，我希望 AI 只审查我修改的代码，而不是整个文件。

**功能点**：
- 仅提取 PR 中**变更的代码行**（+/-）
- 调用 LLM API 进行代码审查
- 支持 Go/Java/Python 三种语言识别

**Prompt 模板**：
```
你是一个资深代码审查专家。请审查以下代码变更，指出：
1. 潜在的 Bug 或逻辑错误
2. 性能问题（如循环内查库、N+1 查询）
3. 安全隐患（如 SQL 注入、XSS）
4. 代码风格问题

代码语言：{language}
变更内容：
{diff_content}

请用中文回复，格式如下：
## 问题列表
- **[严重程度]** 文件名:行号 - 问题描述

## 改进建议
- 具体的修复建议
```

**验收标准**：
- [ ] 能正确提取 diff 中的变更代码
- [ ] 能识别代码语言
- [ ] 能调用 LLM API 并获取审查结果

---

### 2.3 GitHub 评论回写 `[必须]`

**用户故事**：作为开发者，我希望在 PR 页面直接看到 AI 的审查意见。

**功能点**：
- 将 AI 审查结果作为 PR Comment 发布
- 支持 Markdown 格式
- 包含审查时间戳和模型信息

**评论格式**：
```markdown
## 🤖 Code-Sentinel 代码审查报告

**审查时间**：2024-01-15 10:30:00
**审查模型**：GPT-4
**变更文件**：3 个文件，+120/-45 行

---

### 🔴 严重问题 (1)
- **main.go:42** - 存在 SQL 注入风险，建议使用参数化查询

### 🟡 建议优化 (2)
- **utils/helper.go:15** - 循环内重复创建对象，建议提取到循环外
- **service/user.go:88** - 缺少错误处理

### 🟢 代码风格 (1)
- **config/config.go:20** - 变量命名不符合 Go 规范，建议使用驼峰命名

---
> 💡 如有误报，请回复 `/false` 标记
```

**验收标准**：
- [ ] 能在 PR 上发布评论
- [ ] 评论格式清晰、易读
- [ ] 包含问题严重程度分类

---

### 2.4 基础配置管理 `[必须]`

**用户故事**：作为管理员，我需要配置 GitHub Token 和 LLM API Key。

**功能点**：
- SQLite 存储配置信息
- 支持配置：
  - GitHub App 凭证（App ID, Private Key）
  - LLM API 配置（API Key, Model, Base URL）
  - 审查规则（启用/禁用的检查项）

**数据模型**：
```sql
-- 仓库配置表
CREATE TABLE repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner VARCHAR(100) NOT NULL,        -- 仓库所有者
    name VARCHAR(100) NOT NULL,         -- 仓库名称
    webhook_secret VARCHAR(255),        -- Webhook 密钥
    enabled BOOLEAN DEFAULT true,       -- 是否启用
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(owner, name)
);

-- 系统配置表
CREATE TABLE configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key VARCHAR(100) UNIQUE NOT NULL,   -- 配置键
    value TEXT NOT NULL,                -- 配置值（加密存储敏感信息）
    description VARCHAR(255),           -- 配置说明
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 审查记录表
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id INTEGER NOT NULL,           -- 关联仓库
    pr_number INTEGER NOT NULL,         -- PR 编号
    commit_sha VARCHAR(40) NOT NULL,    -- Commit SHA
    status VARCHAR(20) NOT NULL,        -- pending/completed/failed
    result TEXT,                        -- AI 审查结果（JSON）
    token_used INTEGER,                 -- Token 消耗
    duration_ms INTEGER,                -- 耗时（毫秒）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (repo_id) REFERENCES repos(id)
);
```

**验收标准**：
- [ ] 能持久化存储配置
- [ ] 敏感信息加密存储
- [ ] 能记录审查历史

---

### 2.5 简易 Web 管理界面 `[可选-建议]`

**用户故事**：作为管理员，我希望有一个简单的界面来管理配置和查看审查记录。

**功能点**：
- 配置管理页面（CRUD）
- 审查记录列表
- 简单的统计（今日审查数、问题数）

**技术方案**：
- 使用 Go 内嵌静态文件
- 前端使用简单的 HTML + Alpine.js 或纯 HTML

**验收标准**：
- [ ] 能查看和修改配置
- [ ] 能查看审查历史记录

---

## 3. 技术架构

### 3.1 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        GitHub                                │
│  ┌─────────┐    Webhook     ┌─────────┐    API              │
│  │   PR    │ ──────────────→│         │←─────────────────── │
│  └─────────┘                │         │     Comment         │
└─────────────────────────────┼─────────┼─────────────────────┘
                              │         │
                              ▼         │
┌─────────────────────────────────────────────────────────────┐
│                    Code-Sentinel MVP                         │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Webhook    │───→│   Analyzer   │───→│   Commenter  │  │
│  │   Handler    │    │   Service    │    │   Service    │  │
│  └──────────────┘    └──────┬───────┘    └──────────────┘  │
│                             │                               │
│                             ▼                               │
│                      ┌──────────────┐                       │
│                      │  LLM Client  │                       │
│                      │ (OpenAI/等)  │                       │
│                      └──────────────┘                       │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                     SQLite                            │  │
│  │  repos | configs | reviews                            │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 目录结构

```
code-sentinel/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载
│   ├── handler/
│   │   ├── webhook.go           # Webhook 处理
│   │   └── api.go               # API 接口
│   ├── service/
│   │   ├── analyzer.go          # 代码分析服务
│   │   ├── github.go            # GitHub API 封装
│   │   └── llm.go               # LLM 调用封装
│   ├── model/
│   │   └── model.go             # 数据模型
│   └── store/
│       └── sqlite.go            # SQLite 存储
├── pkg/
│   ├── diff/
│   │   └── parser.go            # Diff 解析
│   └── prompt/
│       └── template.go          # Prompt 模板
├── web/                         # 静态前端文件（可选）
├── configs/
│   └── config.yaml              # 配置文件
├── Dockerfile
├── docker-compose.yaml
├── go.mod
├── go.sum
└── README.md
```

### 3.3 核心流程

```
1. GitHub 发送 Webhook
       │
       ▼
2. 验证签名 & 解析 Payload
       │
       ▼
3. 获取 PR Diff（调用 GitHub API）
       │
       ▼
4. 提取变更代码 & 识别语言
       │
       ▼
5. 组装 Prompt & 调用 LLM
       │
       ▼
6. 解析 AI 响应 & 格式化
       │
       ▼
7. 发布 PR Comment（调用 GitHub API）
       │
       ▼
8. 记录审查结果到 SQLite
```

---

## 4. 接口设计

### 4.1 Webhook 接口

```yaml
POST /webhook/github
Content-Type: application/json
X-GitHub-Event: pull_request
X-Hub-Signature-256: sha256=xxx

Response:
  200: { "status": "ok", "review_id": 123 }
  400: { "error": "invalid payload" }
  401: { "error": "invalid signature" }
```

### 4.2 配置接口

```yaml
# 获取配置列表
GET /api/configs
Response: [{ "key": "llm_api_key", "value": "***", "description": "..." }]

# 更新配置
PUT /api/configs/:key
Body: { "value": "new_value" }

# 获取仓库列表
GET /api/repos
Response: [{ "id": 1, "owner": "xxx", "name": "yyy", "enabled": true }]

# 添加仓库
POST /api/repos
Body: { "owner": "xxx", "name": "yyy", "webhook_secret": "xxx" }
```

### 4.3 审查记录接口

```yaml
# 获取审查记录
GET /api/reviews?repo_id=1&page=1&size=20
Response: {
  "total": 100,
  "items": [{
    "id": 1,
    "pr_number": 42,
    "status": "completed",
    "result": {...},
    "created_at": "2024-01-15T10:30:00Z"
  }]
}
```

---

## 5. 部署方案

### 5.1 Docker 部署

```yaml
# docker-compose.yaml
version: '3.8'
services:
  code-sentinel:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data          # SQLite 数据持久化
      - ./configs:/app/configs    # 配置文件
    environment:
      - GIN_MODE=release
      - CONFIG_PATH=/app/configs/config.yaml
```

### 5.2 配置文件

```yaml
# configs/config.yaml
server:
  port: 8080
  mode: release

database:
  path: ./data/sentinel.db

github:
  app_id: ${GITHUB_APP_ID}
  private_key_path: ./configs/github-app.pem
  # 或使用 Personal Access Token
  # token: ${GITHUB_TOKEN}

llm:
  provider: openai          # openai / azure / ollama
  api_key: ${OPENAI_API_KEY}
  model: gpt-4
  base_url: https://api.openai.com/v1
  timeout: 60s
  max_tokens: 4096

review:
  enabled_languages:
    - go
    - java
    - python
  max_diff_lines: 500       # 超过则截断
  ignore_patterns:          # 忽略的文件
    - "*.md"
    - "*.json"
    - "vendor/*"
    - "node_modules/*"
```

---

## 6. 开发计划

### Week 1: 基础框架
| 任务 | 预计耗时 | 产出 |
|------|----------|------|
| 项目初始化 & 目录结构 | 2h | 可运行的空项目 |
| SQLite 存储层 | 4h | 数据模型 & CRUD |
| 配置管理模块 | 2h | 配置加载 & 热更新 |
| GitHub Webhook 接收 | 4h | 能接收并验证 Webhook |
| Diff 解析器 | 4h | 能提取变更代码 |

### Week 2: 核心功能
| 任务 | 预计耗时 | 产出 |
|------|----------|------|
| LLM 客户端封装 | 4h | 支持 OpenAI API |
| Prompt 模板设计 | 2h | 审查 Prompt |
| 代码分析服务 | 6h | 完整的分析流程 |
| GitHub 评论回写 | 4h | 能发布 PR Comment |
| 错误处理 & 日志 | 2h | 完善的错误处理 |

### Week 3: 完善 & 部署
| 任务 | 预计耗时 | 产出 |
|------|----------|------|
| 简易 Web 界面 | 6h | 配置管理页面 |
| Docker 打包 | 2h | Dockerfile |
| 文档编写 | 4h | README & 部署文档 |
| 测试 & Bug 修复 | 4h | 稳定可用的 MVP |

---

## 7. 验收标准

### 7.1 功能验收
- [ ] 能接收 GitHub Webhook 并验证签名
- [ ] 能获取 PR 的 diff 内容
- [ ] 能调用 LLM 进行代码审查
- [ ] 能在 PR 上发布格式化的评论
- [ ] 能持久化存储配置和审查记录

### 7.2 性能指标
- 单次审查响应时间 < 30s（取决于 LLM）
- 支持同时处理 5 个 PR 审查请求

### 7.3 可用性
- Docker 一键部署
- 配置文件清晰易懂
- 有基本的错误日志

---

## 8. 后续迭代方向

MVP 完成后，可按以下优先级迭代：

1. **Phase 2**: 安全拦截（Pre-LLM 敏感信息检测）
2. **Phase 3**: 误报反馈（`/false` 命令 + 错题本）
3. **Phase 4**: 自定义规则（`.sentinel.yaml`）
4. **Phase 5**: RAG 知识库
5. **Phase 6**: 企业级架构（Kafka + K8s）

---

## 附录 A: GitHub App 配置指南

1. 访问 https://github.com/settings/apps/new
2. 填写 App 名称：`Code-Sentinel`
3. Homepage URL：你的服务地址
4. Webhook URL：`https://your-domain.com/webhook/github`
5. Webhook Secret：生成一个随机字符串
6. 权限设置：
   - Pull requests: Read & Write
   - Contents: Read
   - Metadata: Read
7. 订阅事件：
   - Pull request
8. 生成 Private Key 并下载

---

## 附录 B: 环境变量

```bash
# 必需
export GITHUB_APP_ID=123456
export GITHUB_PRIVATE_KEY_PATH=./github-app.pem
export OPENAI_API_KEY=sk-xxx

# 可选
export GIN_MODE=release
export CONFIG_PATH=./configs/config.yaml
```
