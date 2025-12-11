package prompt

import (
	"bytes"
	"text/template"

	"code-sentinel/pkg/diff"
)

const SystemPrompt = `你是一个资深的代码审查专家，拥有 10 年以上的软件开发经验。
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

---

请按以下格式输出审查结果：

## 🔴 严重问题
（如果没有则写"无"）

## 🟡 警告
（如果没有则写"无"）

## 🟢 建议
（如果没有则写"无"）

## 📝 总结
（简要总结代码质量，给出 1-10 分的评分）`

type PromptData struct {
	FileCount     int
	AdditionCount int
	DeletionCount int
	MainLanguage  string
	DiffContent   string
}

type Builder struct {
	template *template.Template
}

func NewBuilder() *Builder {
	tpl := template.Must(template.New("review").Parse(ReviewPromptTemplate))
	return &Builder{template: tpl}
}

func (b *Builder) Build(changes []diff.FileChange) (string, string, error) {
	data := PromptData{
		FileCount:    len(changes),
		MainLanguage: detectMainLanguage(changes),
		DiffContent:  diff.FormatChangesForPrompt(changes),
	}

	for _, c := range changes {
		data.AdditionCount += len(c.Additions)
		data.DeletionCount += len(c.Deletions)
	}

	var buf bytes.Buffer
	if err := b.template.Execute(&buf, data); err != nil {
		return "", "", err
	}

	return SystemPrompt, buf.String(), nil
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
