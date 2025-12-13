package prompt

import (
	"bytes"
	"text/template"

	"code-sentinel/pkg/diff"
)

const ReviewPromptTemplate = `请审查以下代码变更：

## 变更概览
- **文件数量**：{{.FileCount}}
- **新增行数**：{{.AdditionCount}}
- **删除行数**：{{.DeletionCount}}
- **主要语言**：{{.MainLanguage}}

## 变更详情
{{.DiffContent}}

请按要求输出 JSON 格式的审查结果。`

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
	return &Builder{
		template: tpl,
	}
}

// BuildUserPrompt 构建用户提示词
func (b *Builder) BuildUserPrompt(changes []diff.FileChange) (string, error) {
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
		return "", err
	}

	return buf.String(), nil
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
