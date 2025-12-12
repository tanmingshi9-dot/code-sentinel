# 📘 Code-Sentinel Phase 2 需求文档

| 文档版本 | V1.0 |
| :--- | :--- |
| **阶段名称** | Phase 2: Web管理后台与配置能力 |
| **核心目标** | 让产品真正可用，支持多项目接入和可视化管理 |
| **预计工期** | 3-4 周 |
| **技术栈** | React, TailwindCSS, shadcn/ui, Go Gin, PostgreSQL (可选) |

---

## 1. 背景与目标

### 1.1 Phase 1 现状
MVP 阶段已完成：
- ✅ GitHub Webhook 接收
- ✅ AI 代码审查（增量审查）
- ✅ GitHub PR 评论回写
- ✅ SQLite 数据持久化

### 1.2 Phase 1 痛点
- **配置不灵活**：只能修改配置文件 + 重启服务
- **无法多项目接入**：所有仓库共用一套配置
- **无数据可视化**：不知道审查了多少 PR，发现了哪些问题
- **无反馈机制**：AI 误报了也无法告知系统

### 1.3 Phase 2 目标
1. **产品化**：从"能用"到"好用"，支持多项目生产使用
2. **可配置**：不同项目独立配置，灵活控制审查策略
3. **可视化**：审查历史查询、问题统计、数据分析
4. **可改进**：收集误报反馈，持续优化审查质量

---

## 2. 核心功能需求

### 2.1 功能一：Web 管理后台 `[P0]`

#### 2.1.1 仓库管理

**功能描述**：
提供可视化界面管理接入的 GitHub 仓库。

**核心页面**：

##### (1) 仓库列表页 `/repos`
展示所有已接入的仓库，支持增删改查。

**页面元素**：
- 仓库列表表格
  - 列：仓库全名、启用状态、最后审查时间、审查次数、操作
  - 操作按钮：编辑配置、删除、启用/禁用开关
- 新建仓库按钮
- 搜索框（按仓库名搜索）

**交互逻辑**：
- 点击"新建仓库" → 跳转到新建页面
- 点击"编辑配置" → 跳转到配置页面
- 点击"启用/禁用"开关 → 即时生效
- 点击"删除" → 弹窗确认 → 删除仓库及相关数据

##### (2) 新建仓库页 `/repos/new`
添加新的 GitHub 仓库到系统。

**表单字段**：
- 仓库全名（必填）：`owner/repo` 格式，如 `tanmingshi9/ai-agent`
- 是否启用（默认启用）
- 配置模板选择（可选）：
  - 默认配置
  - 前端项目模板
  - 后端项目模板
  - 自定义

**验证逻辑**：
- 检查仓库名格式是否正确
- 检查仓库是否已存在
- （可选）验证 GitHub 仓库是否可访问

##### (3) 仓库配置页 `/repos/:id/config`
详细配置每个仓库的审查策略。

**配置分组**：

**A. 基础配置**
- 仓库全名（只读显示）
- 是否启用审查
- LLM 提供商：OpenAI / 阿里通义千问 / Azure / Ollama
- LLM 模型：gpt-4 / gpt-4-turbo / qwen-max / qwen-plus 等
- 最大 Token 数：限制单次审查的 Token 消耗

**B. 审查配置**
- 自定义系统提示词（Textarea，支持 Markdown）
  - 默认值：系统预置的通用 Prompt
  - 用户可自定义审查重点、风格
- 审查重点（多选框）：
  - [ ] 安全问题（SQL 注入、XSS、硬编码密钥等）
  - [ ] 性能问题（循环内查库、N+1 查询等）
  - [ ] 代码风格（命名规范、注释质量等）
  - [ ] 逻辑错误（空指针、边界条件等）
- 最小报告级别（下拉框）：
  - 仅 P0（严重问题）
  - P0 + P1（严重 + 重要）
  - 全部（P0 + P1 + P2）

**C. 过滤规则**
- 支持的语言（多选）：Go、Java、Python、JavaScript、TypeScript、Rust 等
- 忽略文件规则（列表，支持 glob 模式）：
  - 示例：`*.test.go`、`vendor/*`、`node_modules/*`、`docs/*`
  - 支持添加/删除规则
- 最大 Diff 行数：超过此行数的 PR 不自动审查（防止超大 PR 消耗大量 Token）

**D. 高级选项**
- 是否自动审查（默认开启）：关闭后仅手动触发
- Webhook Secret：用于验证 GitHub Webhook 签名
- 审查触发条件（多选）：
  - [ ] PR 打开时（opened）
  - [ ] PR 同步时（synchronize，即新 commit 推送）
  - [ ] PR 重新打开时（reopened）

**操作按钮**：
- 保存配置
- 重置为默认
- 测试配置（发送测试请求验证配置是否正确）

---

#### 2.1.2 审查历史查询

**功能描述**：
查询某个仓库或全局的审查记录，支持筛选和搜索。

**核心页面**：

##### 审查历史页 `/reviews`

**筛选条件**：
- 仓库选择（下拉框，支持全部仓库）
- 时间范围（日期选择器）
- 审查状态：全部 / 成功 / 失败 / 跳过
- PR 编号搜索

**列表展示**：
表格展示审查记录，列包括：
- 仓库名称
- PR 编号
- PR 标题
- 审查状态（成功/失败/跳过）
- 发现问题数量（P0/P1/P2）
- Token 消耗
- 审查时长
- 审查时间
- 操作（查看详情）

**详情弹窗**：
点击"查看详情"弹出弹窗，展示：
- PR 基本信息（作者、分支、Commit SHA）
- AI 审查结果（完整的评论内容）
- 问题列表（按严重程度分组）
- 元数据（使用的模型、Token 消耗、耗时）

---

### 2.2 功能二：优化 Prompt 模板 `[P0]`

#### 2.2.1 背景
当前 Prompt 输出格式不统一，导致：
- 前端难以解析和展示
- 无法统计问题类型
- 严重程度不明确

#### 2.2.2 优化目标
1. **结构化输出**：统一格式，方便解析
2. **明确严重程度**：P0/P1/P2 分级
3. **提供修复建议**：不仅指出问题，还给出解决方案
4. **可扩展性**：支持不同语言定制 Prompt

#### 2.2.3 Prompt 模板设计

**系统提示词（System Prompt）**：
```markdown
你是资深代码审查专家，精通 {languages} 开发。

你的任务是审查代码变更，识别潜在问题，并提供详细的修复建议。

## 审查重点
{review_focus}

## 严重程度定义
- P0（严重）：安全漏洞、会导致系统崩溃或数据泄露的问题
- P1（重要）：性能问题、明显的逻辑错误、潜在的 Bug
- P2（建议）：代码风格、注释质量、可读性改进

## 输出格式要求
请严格按照以下 JSON 格式输出，不要添加任何额外内容：

```json
{
  "summary": "本次审查总体评价（1-2句话）",
  "issues": [
    {
      "severity": "P0|P1|P2",
      "category": "security|performance|logic|style",
      "file": "文件路径",
      "line": 行号,
      "title": "问题标题（简短）",
      "description": "问题详细描述",
      "suggestion": "修复建议",
      "code_fix": "修复后的代码片段（可选）"
    }
  ],
  "stats": {
    "p0_count": 0,
    "p1_count": 0,
    "p2_count": 0
  }
}
```

## 注意事项
- 如果代码没有问题，issues 返回空数组
- code_fix 字段仅在能提供具体修复代码时填写
- 保持客观和专业，避免主观判断
```

**用户提示词（User Prompt）**：
```markdown
请审查以下代码变更：

## 变更文件
{file_changes}

请按要求输出 JSON 格式的审查结果。
```

#### 2.2.4 语言定制
针对不同语言，在 `{review_focus}` 中注入特定关注点：

**Go 语言**：
- goroutine 泄露和并发安全
- error 处理是否完整
- defer 使用是否正确

**Python 语言**：
- 类型提示（Type Hints）
- 异常处理
- 资源管理（with 语句）

**Java 语言**：
- 空指针异常（NullPointerException）
- 资源关闭（try-with-resources）
- 线程安全

#### 2.2.5 验收标准
- AI 输出 100% 符合 JSON 格式
- 严重程度标注准确率 > 90%
- 每个问题都包含修复建议

---

### 2.3 功能三：误报反馈机制 `[P0]`

#### 2.3.1 背景
AI 会产生误报，必须有反馈渠道，用于：
- 让开发者标记误报
- 收集数据，后续优化 Prompt
- 为 Phase 3 的"错题本"功能提供数据基础

#### 2.3.2 功能设计

##### (1) GitHub 评论反馈
在 AI 的 PR 评论末尾添加提示：
```markdown
---
💡 **反馈机制**
如果发现误报，请回复：
- `/false` - 标记为误报
- `/false [原因]` - 标记误报并说明原因

示例：`/false 这是测试代码，不需要修复`
```

**实现逻辑**：
- 监听 PR 评论事件（`issue_comment` webhook）
- 检测评论内容是否包含 `/false`
- 提取误报原因（可选）
- 关联到对应的审查记录
- 存入数据库

##### (2) 误报数据存储

**数据表设计**：
```go
type Feedback struct {
    ID        uint      `gorm:"primaryKey"`
    ReviewID  uint      `gorm:"index"` // 关联的审查记录
    IssueIndex int      // 问题在审查结果中的索引
    
    // 问题信息（快照，防止审查记录被删）
    File      string    // 文件路径
    Line      int       // 行号
    Severity  string    // P0/P1/P2
    Category  string    // security/performance/logic/style
    Title     string    // 问题标题
    AIContent string    `gorm:"type:text"` // AI 原始判断
    
    // 反馈信息
    IsFalsePositive bool   // 是否误报
    Reason          string // 用户提供的原因
    Reporter        string // 反馈人（GitHub 用户名）
    
    CreatedAt time.Time
}
```

##### (3) Web 后台误报管理

**误报列表页** `/feedbacks`

**筛选条件**：
- 仓库选择
- 问题类型（安全/性能/逻辑/风格）
- 严重程度（P0/P1/P2）
- 时间范围

**列表展示**：
- 仓库名称
- PR 编号
- 文件:行号
- 问题标题
- AI 判断（摘要）
- 误报原因
- 反馈人
- 反馈时间

**操作**：
- 查看详情（弹窗展示完整信息）
- 导出 CSV（用于后续分析）

##### (4) 误报统计

在统计看板中增加：
- 误报率：误报数 / 总问题数
- 误报趋势图（按周统计）
- 误报高发类型（哪类问题误报最多）

#### 2.3.3 后续规划（Phase 3）
基于收集的误报数据，实现"错题本"功能：
- 将误报案例向量化存入 pgvector
- 审查时检索相似场景
- 在 Prompt 中注入："上次在类似场景你误报了，请注意..."

---

## 3. 非功能需求

### 3.1 性能要求
- Web 页面加载时间 < 2s
- 仓库列表查询响应时间 < 500ms
- 审查历史查询（100 条）< 1s

### 3.2 可用性要求
- 界面响应式，支持桌面和平板
- 关键操作有确认提示（删除仓库等）
- 错误提示友好，有明确的解决建议

### 3.3 扩展性要求
- 支持后续添加更多 LLM 提供商
- Prompt 模板支持版本管理
- 数据库设计预留扩展字段

---

## 4. 技术方案

### 4.1 前端技术栈
- **框架**：React 18 + TypeScript
- **路由**：React Router v6
- **UI 组件**：shadcn/ui（基于 Radix UI）
- **样式**：TailwindCSS
- **状态管理**：React Query（服务端状态）+ Zustand（客户端状态）
- **图表**：Recharts
- **HTTP 客户端**：Axios
- **构建工具**：Vite

### 4.2 后端技术栈
- **框架**：Go Gin（复用现有）
- **数据库**：SQLite（Phase 2 继续使用）或 PostgreSQL（可选升级）
- **ORM**：GORM
- **API 设计**：RESTful

### 4.3 部署方案
- 前端构建产物放入 `web/dist`
- 后端 Gin 服务 Static 文件服务
- Docker 镜像包含前后端
- 单个容器部署（简化运维）

---

## 5. 接口设计

### 5.1 仓库管理接口
```
GET    /api/repos              # 获取仓库列表
POST   /api/repos              # 新建仓库
GET    /api/repos/:id          # 获取仓库详情
PUT    /api/repos/:id          # 更新仓库配置
DELETE /api/repos/:id          # 删除仓库
PUT    /api/repos/:id/toggle   # 启用/禁用仓库
```

### 5.2 审查记录接口
```
GET    /api/reviews            # 获取审查列表（支持筛选）
GET    /api/reviews/:id        # 获取审查详情
```

### 5.3 反馈接口
```
GET    /api/feedbacks          # 获取误报列表
POST   /api/feedbacks          # 创建反馈（Webhook 调用）
GET    /api/feedbacks/stats    # 误报统计
```

### 5.4 配置模板接口
```
GET    /api/config-templates   # 获取预置配置模板
```

---

## 6. 开发计划

### Week 1: 基础架构
- [ ] 前端项目初始化（React + Vite + TailwindCSS）
- [ ] 集成 shadcn/ui 组件库
- [ ] 后端新增 API 路由框架
- [ ] 数据库表设计和迁移脚本

### Week 2: 仓库管理功能
- [ ] 仓库列表页（前端 + 后端）
- [ ] 新建仓库页
- [ ] 仓库配置页（基础配置 + 审查配置）
- [ ] 配置保存和加载逻辑

### Week 3: 审查历史与 Prompt 优化
- [ ] 审查历史列表页
- [ ] 审查详情弹窗
- [ ] 优化 Prompt 模板（JSON 格式）
- [ ] Prompt 模板测试和验证

### Week 4: 反馈机制与联调
- [ ] GitHub 评论监听（`/false` 命令）
- [ ] 误报数据存储
- [ ] 误报列表页
- [ ] 前后端联调
- [ ] 部署脚本优化

---

## 7. 验收标准

### 7.1 功能验收
- [ ] 可通过 Web 界面新建和配置仓库
- [ ] 不同仓库使用独立的配置审查 PR
- [ ] 审查历史可查询，支持筛选
- [ ] 统计看板展示正确的数据和图表
- [ ] `/false` 命令能成功标记误报
- [ ] 误报数据正确存储到数据库

### 7.2 质量验收
- [ ] 前端页面响应速度 < 2s
- [ ] API 响应时间符合性能要求
- [ ] 界面在 Chrome/Firefox/Safari 正常展示
- [ ] 关键操作有错误提示和确认

### 7.3 文档验收
- [ ] README 更新部署说明
- [ ] API 文档完整（Swagger 或 Markdown）
- [ ] 用户使用手册（如何配置仓库）

---

## 8. 风险与应对

### 8.1 技术风险
**风险**：Prompt 输出的 JSON 格式不稳定，AI 偶尔不按格式输出

**应对**：
- 在 Prompt 中强调格式要求
- 后端添加 JSON 解析兜底逻辑
- 解析失败时记录日志，使用降级逻辑（纯文本展示）

### 8.2 数据风险
**风险**：SQLite 并发性能不足，多仓库同时审查可能卡顿

**应对**：
- Phase 2 优先使用 SQLite，出现瓶颈再迁移 PostgreSQL
- 设计时保持数据库抽象层，便于后续切换

### 8.3 用户体验风险
**风险**：配置项过多，用户不知道怎么配

**应对**：
- 提供合理的默认配置
- 提供配置模板（前端/后端项目）
- 关键配置项添加帮助提示（Tooltip）

---

## 9. 后续规划（Phase 3 方向）

基于 Phase 2 的数据积累，Phase 3 可以做：
1. **错题本增强**：基于误报数据，RAG 检索相似场景
2. **智能问答**：代码向量化，支持"某个功能在哪里实现"
3. **多模型对比**：同一 PR 用不同模型审查，比较效果
4. **自动修复**：高置信度的问题直接生成修复 PR

---

## 附录 A：配置示例

### 默认配置
```yaml
llm_provider: openai
model: gpt-4-turbo
system_prompt: |
  你是资深代码审查专家，请识别代码中的安全漏洞、性能问题和逻辑错误。
review_focus:
  - security
  - performance
  - logic
languages:
  - go
  - python
  - javascript
  - typescript
ignore_files:
  - "*.test.go"
  - "vendor/*"
  - "node_modules/*"
min_severity: P1
max_diff_lines: 1000
auto_review: true
```

### 前端项目模板
```yaml
llm_provider: openai
model: gpt-4-turbo
system_prompt: |
  你是资深前端工程师，重点关注 React/Vue 最佳实践、性能优化、可访问性。
review_focus:
  - security  # XSS, CSRF
  - performance  # 不必要的重渲染
  - style  # 组件命名规范
languages:
  - javascript
  - typescript
  - jsx
  - tsx
ignore_files:
  - "dist/*"
  - "build/*"
  - "*.test.tsx"
min_severity: P1
```

---

## 附录 B：Prompt 示例

### 输入（User Prompt）
```markdown
请审查以下代码变更：

## 文件: internal/service/user.go
```diff
+ func (s *UserService) DeleteUser(userId string) error {
+     query := "DELETE FROM users WHERE id = " + userId
+     return s.db.Exec(query).Error
+ }
```

请按要求输出 JSON 格式的审查结果。
```

### 期望输出
```json
{
  "summary": "发现 1 个严重的 SQL 注入漏洞，必须立即修复。",
  "issues": [
    {
      "severity": "P0",
      "category": "security",
      "file": "internal/service/user.go",
      "line": 2,
      "title": "SQL 注入漏洞",
      "description": "直接拼接用户输入到 SQL 语句中，存在 SQL 注入风险。攻击者可通过构造特殊的 userId 删除任意数据。",
      "suggestion": "使用参数化查询（Prepared Statement）防止 SQL 注入",
      "code_fix": "query := \"DELETE FROM users WHERE id = ?\"\nreturn s.db.Exec(query, userId).Error"
    }
  ],
  "stats": {
    "p0_count": 1,
    "p1_count": 0,
    "p2_count": 0
  }
}
```
