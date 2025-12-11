package diff

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type FileChange struct {
	Filename  string
	Language  string
	OldPath   string
	NewPath   string
	Additions []Line
	Deletions []Line
	Hunks     []Hunk
}

type Line struct {
	Number  int
	Content string
}

type Hunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Content  string
}

var hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

func ParseDiff(diffContent string) ([]FileChange, error) {
	var changes []FileChange

	fileDiffs := splitByFile(diffContent)

	for _, fileDiff := range fileDiffs {
		if fileDiff == "" {
			continue
		}

		change := parseFileDiff(fileDiff)
		if change.Filename != "" {
			changes = append(changes, change)
		}
	}

	return changes, nil
}

func splitByFile(diffContent string) []string {
	var files []string
	var current strings.Builder

	lines := strings.Split(diffContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if current.Len() > 0 {
				files = append(files, current.String())
				current.Reset()
			}
		}
		current.WriteString(line)
		current.WriteString("\n")
	}

	if current.Len() > 0 {
		files = append(files, current.String())
	}

	return files
}

func parseFileDiff(fileDiff string) FileChange {
	change := FileChange{}
	lines := strings.Split(fileDiff, "\n")

	var inHunk bool
	var currentHunk strings.Builder
	var hunkNewLine int

	for _, line := range lines {
		if strings.HasPrefix(line, "--- a/") {
			change.OldPath = strings.TrimPrefix(line, "--- a/")
		} else if strings.HasPrefix(line, "+++ b/") {
			change.NewPath = strings.TrimPrefix(line, "+++ b/")
			change.Filename = change.NewPath
			change.Language = detectLanguage(change.Filename)
		} else if strings.HasPrefix(line, "@@") {
			if inHunk && currentHunk.Len() > 0 {
				// Save previous hunk
			}
			inHunk = true
			currentHunk.Reset()

			matches := hunkHeaderRegex.FindStringSubmatch(line)
			if len(matches) >= 4 {
				newStart, _ := strconv.Atoi(matches[3])
				hunkNewLine = newStart
			}
			currentHunk.WriteString(line)
			currentHunk.WriteString("\n")
		} else if inHunk {
			currentHunk.WriteString(line)
			currentHunk.WriteString("\n")

			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				change.Additions = append(change.Additions, Line{
					Number:  hunkNewLine,
					Content: strings.TrimPrefix(line, "+"),
				})
				hunkNewLine++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				change.Deletions = append(change.Deletions, Line{
					Number:  hunkNewLine,
					Content: strings.TrimPrefix(line, "-"),
				})
			} else if !strings.HasPrefix(line, "\\") {
				hunkNewLine++
			}
		}
	}

	return change
}

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
	case ".jsx", ".tsx":
		return "react"
	case ".rs":
		return "rust"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".cs":
		return "csharp"
	case ".sh", ".bash":
		return "shell"
	case ".sql":
		return "sql"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".md":
		return "markdown"
	default:
		return "unknown"
	}
}

func ShouldIgnore(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filepath.Base(filename))
		if err == nil && matched {
			return true
		}

		if strings.Contains(pattern, "/") {
			matched, err = filepath.Match(pattern, filename)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

func FormatChangesForPrompt(changes []FileChange) string {
	var sb strings.Builder

	for _, change := range changes {
		sb.WriteString("### File: ")
		sb.WriteString(change.Filename)
		sb.WriteString(" (")
		sb.WriteString(change.Language)
		sb.WriteString(")\n\n")

		if len(change.Additions) > 0 {
			sb.WriteString("**Added lines:**\n```")
			sb.WriteString(change.Language)
			sb.WriteString("\n")
			for _, line := range change.Additions {
				sb.WriteString("// Line ")
				sb.WriteString(strconv.Itoa(line.Number))
				sb.WriteString("\n")
				sb.WriteString(line.Content)
				sb.WriteString("\n")
			}
			sb.WriteString("```\n\n")
		}

		if len(change.Deletions) > 0 {
			sb.WriteString("**Deleted lines:**\n```")
			sb.WriteString(change.Language)
			sb.WriteString("\n")
			for _, line := range change.Deletions {
				sb.WriteString(line.Content)
				sb.WriteString("\n")
			}
			sb.WriteString("```\n\n")
		}
	}

	return sb.String()
}
