package prompt

import (
	"bytes"
	"strings"
	"text/template"

	"code-sentinel/pkg/diff"
)

// 审查重点映射
var reviewFocusMap = map[string]string{
	"security":    "- 安全问题：SQL 注入、XSS、硬编码密钥、敏感信息泄露、不安全的加密",
	"performance": "- 性能问题：循环内查库、N+1 查询、不必要的重复计算、内存泄漏",
	"logic":       "- 逻辑错误：空指针、边界条件、异常处理不当、死循环、竞态条件",
	"style":       "- 代码风格：命名规范、注释质量、代码可读性、过长函数",
}

// SystemPromptTemplate JSON 结构化输出的系统提示词模板
const SystemPromptTemplate = `你是资深代码审查专家，精通 {{.Languages}} 开发。

你的任务是审查代码变更，识别潜在问题，并提供详细的修复建议。

## 审查重点
{{.ReviewFocus}}

## 严重程度定义
- P0（严重）：安全漏洞、会导致系统崩溃或数据泄露的问题
- P1（重要）：性能问题、明显的逻辑错误、潜在的 Bug
- P2（建议）：代码风格、注释质量、可读性改进

## 输出格式要求
请严格按照以下 JSON 格式输出，不要添加任何额外内容：

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

## 注意事项
- 如果代码没有问题，issues 返回空数组，summary 写 "代码质量良好，未发现明显问题"
- code_fix 字段仅在能提供具体修复代码时填写
- 保持客观和专业，避免主观判断
- 确保输出的是合法的 JSON，不要包含注释或额外文本
{{if .CustomPrompt}}

## 额外要求
{{.CustomPrompt}}
{{end}}`

// LegacySystemPrompt 向后兼容的旧版系统提示词
const LegacySystemPrompt = `你是一个资深的代码审查专家，拥有 10 年以上的软件开发经验。
你的任务是审查代码变更，找出潜在的问题并提供改进建议。

审查重点：
1. **Bug 和逻辑错误**：空指针、数组越界、逻辑漏洞、边界条件
2. **性能问题**：循环内查库、N+1 查询、不必要的内存分配、算法复杂度
3. **安全隐患**：SQL 注入、XSS、敏感信息泄露、不安全的加密
4. **代码质量**：命名规范、代码重复、过长函数、复杂度过高

输出要求：
- 使用中文回复
- 按严重程度分类（🔴 严重 / 🟡 警告 / 🟢 建议）
- 指出具体的文件名和行号
- 提供具体的修复建议
- 如果代码质量良好，也请给出肯定`

const ReviewPromptTemplate = `请审查以下代码变更：

## 变更概览
- **文件数量**：{{.FileCount}}
- **新增行数**：{{.AdditionCount}}
- **删除行数**：{{.DeletionCount}}
- **主要语言**：{{.MainLanguage}}

## 变更详情
{{.DiffContent}}

请按要求输出 JSON 格式的审查结果。`

// SystemPromptData 系统提示词数据
type SystemPromptData struct {
	Languages    string
	ReviewFocus  string
	CustomPrompt string
}

type PromptData struct {
	FileCount     int
	AdditionCount int
	DeletionCount int
	MainLanguage  string
	DiffContent   string
}

// ReviewConfig Prompt 配置
type ReviewConfig struct {
	Languages    []string
	ReviewFocus  []string
	CustomPrompt string
}

type Builder struct {
	template       *template.Template
	systemTemplate *template.Template
}

func NewBuilder() *Builder {
	tpl := template.Must(template.New("review").Parse(ReviewPromptTemplate))
	sysTpl := template.Must(template.New("system").Parse(SystemPromptTemplate))
	return &Builder{
		template:       tpl,
		systemTemplate: sysTpl,
	}
}

// Build 构建提示词（向后兼容）
func (b *Builder) Build(changes []diff.FileChange) (string, string, error) {
	return b.BuildWithConfig(changes, nil)
}

// BuildWithConfig 使用配置构建提示词
func (b *Builder) BuildWithConfig(changes []diff.FileChange, config *ReviewConfig) (string, string, error) {
	// 构建用户提示词
	data := PromptData{
		FileCount:    len(changes),
		MainLanguage: detectMainLanguage(changes),
		DiffContent:  diff.FormatChangesForPrompt(changes),
	}

	for _, c := range changes {
		data.AdditionCount += len(c.Additions)
		data.DeletionCount += len(c.Deletions)
	}

	var userBuf bytes.Buffer
	if err := b.template.Execute(&userBuf, data); err != nil {
		return "", "", err
	}

	// 构建系统提示词
	systemPrompt := b.buildSystemPrompt(config)

	return systemPrompt, userBuf.String(), nil
}

// buildSystemPrompt 构建系统提示词
func (b *Builder) buildSystemPrompt(config *ReviewConfig) string {
	if config == nil {
		return LegacySystemPrompt
	}

	// 解析语言列表
	languages := "Go、Python、JavaScript"
	if len(config.Languages) > 0 {
		languages = strings.Join(config.Languages, "、")
	}

	// 解析审查重点
	var focusItems []string
	if len(config.ReviewFocus) > 0 {
		for _, focus := range config.ReviewFocus {
			if desc, ok := reviewFocusMap[focus]; ok {
				focusItems = append(focusItems, desc)
			}
		}
	} else {
		// 默认所有重点
		for _, desc := range reviewFocusMap {
			focusItems = append(focusItems, desc)
		}
	}
	reviewFocus := strings.Join(focusItems, "\n")

	sysData := SystemPromptData{
		Languages:    languages,
		ReviewFocus:  reviewFocus,
		CustomPrompt: config.CustomPrompt,
	}

	var sysBuf bytes.Buffer
	if err := b.systemTemplate.Execute(&sysBuf, sysData); err != nil {
		return LegacySystemPrompt
	}

	return sysBuf.String()
}

func detectMainLanguage(changes []diff.FileChange) string {
	langCount := make(map[string]int)

	for _, c := range changes {
		if c.Language != "unknown" {
			langCount[c.Language] += len(c.Additions) + len(c.Deletions)
		}
	}

	var mainLang string
	var maxCount int
	for lang, count := range langCount {
		if count > maxCount {
			maxCount = count
			mainLang = lang
		}
	}

	if mainLang == "" {
		return "unknown"
	}
	return mainLang
}
