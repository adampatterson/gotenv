package cmd

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/adampatterson/gotenv/pkg/gotenv"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A371F7")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#58A6FF")).
			Padding(0, 1)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E6EDF3")).
			Padding(0, 1)

	commentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B949E")).
			Italic(true).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2EA043"))
)

// cleanCommentForDisplay removes leading comment markers and normalizes whitespace for table display.
func cleanCommentForDisplay(rawComment string) string {
	rawComment = strings.TrimSpace(rawComment)
	if rawComment == "" {
		return ""
	}
	lines := strings.Split(rawComment, "\n")
	var cleanedLines []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		trimmed = strings.TrimLeft(trimmed, "#")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}
	return strings.Join(cleanedLines, " ")
}

// RenderTable constructs a stylized Lip Gloss table with the provided headers and rows.
func RenderTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			switch col {
			case 0:
				return keyStyle
			case 1:
				return valueStyle
			case 2:
				return commentStyle
			default:
				return lipgloss.NewStyle().Padding(0, 1)
			}
		})

	return t.Render()
}

// RenderListTable formats a list of environment variables and their comments as a stylized table.
func RenderListTable(envMap map[string]string, envText string) string {
	keys := make([]string, 0, len(envMap))
	for k := range envMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hasAnyComment := false
	comments := make(map[string]string, len(keys))
	for _, k := range keys {
		c := cleanCommentForDisplay(gotenv.GetEnvVarComment(envText, k))
		if c != "" {
			hasAnyComment = true
		}
		comments[k] = c
	}

	var headers []string
	if hasAnyComment {
		headers = []string{"KEY", "VALUE", "COMMENT"}
	} else {
		headers = []string{"KEY", "VALUE"}
	}

	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		val := envMap[k]
		if hasAnyComment {
			rows = append(rows, []string{k, val, comments[k]})
		} else {
			rows = append(rows, []string{k, val})
		}
	}

	return RenderTable(headers, rows)
}

// RenderGetWithCommentsTable renders a single-row stylized table for a retrieved environment variable.
func RenderGetWithCommentsTable(key string, val string, comment string) string {
	cleanComment := cleanCommentForDisplay(comment)
	var headers []string
	var row []string

	if cleanComment != "" {
		headers = []string{"KEY", "VALUE", "COMMENT"}
		row = []string{key, val, cleanComment}
	} else {
		headers = []string{"KEY", "VALUE"}
		row = []string{key, val}
	}

	return RenderTable(headers, [][]string{row})
}

// RenderSetPreview formats the updated/created variables preview in a stylized table.
func RenderSetPreview(filePath string, keys []string, envMap map[string]string, commentsMap map[string]string) string {
	sort.Strings(keys)

	hasAnyComment := false
	for _, k := range keys {
		if commentsMap != nil && strings.TrimSpace(commentsMap[k]) != "" {
			hasAnyComment = true
			break
		}
	}

	var headers []string
	if hasAnyComment {
		headers = []string{"KEY", "VALUE", "COMMENT"}
	} else {
		headers = []string{"KEY", "VALUE"}
	}

	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		val := envMap[k]
		if gotenv.IsSecret(k, val) {
			val = gotenv.ObfuscateValue(val)
		}
		escapedVal := gotenv.EscapeEnvFileValue(val)
		comment := ""
		if commentsMap != nil {
			comment = cleanCommentForDisplay(commentsMap[k])
		}

		if hasAnyComment {
			rows = append(rows, []string{k, escapedVal, comment})
		} else {
			rows = append(rows, []string{k, escapedVal})
		}
	}

	title := titleStyle.Render(fmt.Sprintf("✓ Updated %s with:", filePath))
	tableStr := RenderTable(headers, rows)

	return fmt.Sprintf("%s\n\n%s\n", title, tableStr)
}
