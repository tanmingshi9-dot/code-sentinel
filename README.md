# 🛡️ Code-Sentinel

企业级智能代码审查平台 - AI 驱动的 PR 自动审查工具
12312312312
## 功能特性

- **自动代码审查**：GitHub PR 触发自动 AI 审查
- **增量审查**：仅审查变更代码，节省 Token
- **多语言支持**：Go、Java、Python
- **GitHub 集成**：自动在 PR 上发布审查评论
- **可配置规则**：支持自定义忽略文件和审查规则

## 快速开始

### 1. 安装依赖

```bash
make deps
```

### 2. 配置

```bash
cp configs/config.example.yaml configs/config.yaml
# 编辑 config.yaml，填入你的配置
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

1. 进入仓库 Settings → Webhooks → Add webhook
2. Payload URL: `https://your-domain.com/webhook/github`
3. Content type: `application/json`
4. Secret: 与配置文件中的 `webhook_secret` 一致
5. Events: 选择 `Pull requests`

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

## 开发

```bash
# 格式化代码
make fmt

# 运行测试
make test

# 构建
make build
```

## License

MIT
