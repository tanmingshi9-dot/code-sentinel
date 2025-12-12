# 🛡️ Code-Sentinel

企业级智能代码审查平台 - AI 驱动的 PR 自动审查工具

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)

## 项目状态

✅ **MVP 阶段已完成** - 核心审查功能已就绪，可投入生产使用

## 功能特性

### 核心能力

- 🤖 **AI 智能审查**：基于大语言模型的深度代码分析
- ⚡ **增量审查**：仅分析变更代码，显著降低 Token 消耗
- 🌍 **多语言支持**：Go、Java、Python、JavaScript、TypeScript 等
- 🔗 **GitHub 深度集成**：自动在 PR 评论中发布审查报告
- 📊 **审查记录**：完整的历史记录和统计分析

### LLM 支持

- ✅ OpenAI (GPT-4, GPT-3.5)
- ✅ 阿里通义千问 (qwen-turbo, qwen-plus, qwen-max)
- ✅ Azure OpenAI
- ✅ 本地模型 (Ollama)

### 审查维度

- 🔴 **严重问题**：Bug、逻辑错误、安全漏洞（SQL注入、XSS等）
- 🟡 **警告**：性能问题、潜在风险、边界条件处理
- 🟢 **建议**：代码风格、最佳实践、可维护性优化

## 快速开始

### 1. 安装依赖

```bash
make deps
```

### 2. 配置

```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml`，配置必要参数：

```yaml
github:
  token: ghp_your_github_token        # GitHub Personal Access Token
  webhook_secret: your_webhook_secret  # Webhook 密钥

llm:
  provider: openai                     # openai / azure / ollama
  api_key: your_api_key               
  model: qwen-plus                     # 模型名称
  base_url: https://dashscope.aliyuncs.com/compatible-mode/v1  # 通义千问
  # base_url: https://api.openai.com/v1  # OpenAI
```

### 3. 运行

```bash
# 直接运行
make run

# 或使用 Docker
make docker-up
```

## 配置说明

### 环境变量

```bash
export GITHUB_TOKEN=ghp_xxx           # GitHub Token
export GITHUB_WEBHOOK_SECRET=xxx      # Webhook 密钥
export OPENAI_API_KEY=sk-xxx          # OpenAI API Key
```

### GitHub Webhook 配置

1. 进入 GitHub 仓库 **Settings** → **Webhooks** → **Add webhook**
2. 填写配置：
   - **Payload URL**: `http://your-server-ip:8080/webhook/github`
   - **Content type**: `application/json`
   - **Secret**: 与 `config.yaml` 中的 `webhook_secret` 一致
   - **Which events**: 选择 **Let me select individual events** → 勾选 **Pull requests**
3. 点击 **Add webhook**
4. 测试：创建一个 PR，查看是否收到 AI 审查评论

### 通义千问 API Key 获取

1. 访问 [阿里云百炼平台](https://bailian.console.aliyun.com/)
2. 登录并进入 **API-KEY 管理**
3. 创建 API Key 并复制到配置文件

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/ready` | 就绪检查 |
| POST | `/webhook/github` | GitHub Webhook |
| GET | `/api/v1/repos` | 获取仓库列表 |
| POST | `/api/v1/repos` | 添加仓库 |
| GET | `/api/v1/reviews` | 获取审查记录 |

## 项目结构

```
code-sentinel/
├── cmd/server/          # 程序入口
├── internal/
│   ├── config/         # 配置管理
│   ├── handler/        # HTTP 处理
│   ├── service/        # 业务逻辑
│   ├── model/          # 数据模型
│   └── store/          # 数据存储
├── pkg/
│   ├── diff/           # Diff 解析
│   ├── prompt/         # Prompt 模板
│   └── signature/      # 签名验证
├── configs/            # 配置文件
└── docs/               # 文档
```

## 演示效果

当你创建或更新 PR 时，Code-Sentinel 会自动：

1. 接收 GitHub Webhook 事件
2. 获取 PR 的代码变更（diff）
3. 调用 AI 进行代码审查
4. 在 PR 中发布审查报告评论

审查报告包含：
- 📊 审查元数据（模型、Token消耗、耗时等）
- 🔴 严重问题列表
- 🟡 警告列表
- 🟢 优化建议
- 📝 总结与评分

## 开发

```bash
# 格式化代码
make fmt

# 运行测试
make test

# 构建二进制
make build

# 清理
make clean
```

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin (HTTP), Viper (配置)
- **数据库**: SQLite (MVP) → PostgreSQL (Phase 2)
- **AI**: OpenAI SDK, 通义千问兼容接口
- **部署**: Docker, Docker Compose

## 路线图

### ✅ Phase 1: MVP (已完成)
- [x] GitHub Webhook 接收
- [x] 增量代码审查
- [x] AI 评论回写
- [x] 基础配置管理
- [x] 审查记录持久化

### 🚧 Phase 2: Web管理后台与配置能力 (开发中)
- [ ] Web 管理后台（仓库管理、配置页面、审查历史、统计看板）
- [ ] 优化 Prompt 模板（JSON 结构化输出、严重程度分级）
- [ ] 误报反馈机制（`/false` 命令、数据记录、统计分析）
- [ ] 多仓库独立配置（LLM 选择、自定义 Prompt、忽略规则）

### 📋 Phase 3: 知识库
- [ ] 代码向量化与 RAG
- [ ] 智能问答助手
- [ ] 相似代码检测

### 📋 Phase 4: 企业级
- [ ] Kafka 消息队列
- [ ] 微服务架构 (Go-Zero)
- [ ] K8s 部署
- [ ] 效能大盘

## 贡献

欢迎贡献代码、提出问题或建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 常见问题

**Q: 支持哪些编程语言？**  
A: 当前支持 Go, Java, Python, JavaScript, TypeScript, Rust, C/C++, Ruby, PHP 等主流语言。

**Q: 如何切换到 OpenAI？**  
A: 修改 `config.yaml` 中的 `llm.base_url` 为 `https://api.openai.com/v1`，并填入 OpenAI API Key。

**Q: Token 消耗如何？**  
A: 增量审查仅分析变更代码，单次 PR 一般消耗 500-2000 tokens，成本在 $0.01-0.05 之间。

**Q: 是否支持私有部署？**  
A: 支持，可配置本地 LLM（Ollama），数据不出内网。

## 相关文档

- [MVP 需求文档](./docs/MVP需求文档.md)
- [完整需求文档](./docs/全部需求文档.md)

## License

MIT © 2024
