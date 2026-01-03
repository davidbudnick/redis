package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidbudnick/redis-tui/internal/types"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).MarginBottom(1)
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("39")).Foreground(lipgloss.Color("0"))
	keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	descStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func getTypeColor(keyType types.KeyType) lipgloss.Color {
	switch keyType {
	case types.KeyTypeString:
		return lipgloss.Color("2") // Green
	case types.KeyTypeList:
		return lipgloss.Color("3") // Yellow
	case types.KeyTypeSet:
		return lipgloss.Color("4") // Blue
	case types.KeyTypeZSet:
		return lipgloss.Color("5") // Magenta
	case types.KeyTypeHash:
		return lipgloss.Color("6") // Cyan
	case types.KeyTypeStream:
		return lipgloss.Color("13") // Bright Magenta
	default:
		return lipgloss.Color("15") // White
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// sanitizeBinaryString checks if a string contains binary/non-printable data
// and returns a safe representation for terminal display
func sanitizeBinaryString(s string) (string, bool) {
	// Check for common binary data signatures
	if strings.HasPrefix(s, "HYLL") { // HyperLogLog
		return fmt.Sprintf("(HyperLogLog data, %d bytes)", len(s)), true
	}

	// Count non-printable characters
	nonPrintable := 0
	for _, r := range s {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			nonPrintable++
		}
		if r > 126 && r < 160 {
			nonPrintable++
		}
	}

	// If more than 10% non-printable, treat as binary
	if len(s) > 0 && float64(nonPrintable)/float64(len(s)) > 0.1 {
		return fmt.Sprintf("(binary data, %d bytes)", len(s)), true
	}

	// Replace any remaining problematic characters
	var result strings.Builder
	for _, r := range s {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			result.WriteString(fmt.Sprintf("\\x%02x", r))
		} else if r > 126 && r < 160 {
			result.WriteString(fmt.Sprintf("\\x%02x", r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String(), false
}

func formatPossibleJSON(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}

	// First, check if this is binary data
	sanitized, isBinary := sanitizeBinaryString(s)
	if isBinary {
		return sanitized
	}
	s = sanitized

	// Check if it looks like JSON (starts with { or [)
	if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, []byte(s), "", "  "); err == nil {
			return colorizeJSON(prettyJSON.String())
		}
	}
	return s
}

// colorizeJSON adds jq-style syntax highlighting to JSON
func colorizeJSON(s string) string {
	var result strings.Builder
	inString := false
	escaped := false
	isKey := false
	afterColon := false

	// jq-style colors
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))     // Blue for keys
	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // Green for string values
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))  // Yellow for numbers
	boolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))    // Magenta for booleans
	nullStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("90"))    // Gray for null
	bracketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White for brackets

	i := 0
	for i < len(s) {
		c := s[i]

		if escaped {
			result.WriteByte(c)
			escaped = false
			i++
			continue
		}

		if c == '\\' && inString {
			result.WriteByte(c)
			escaped = true
			i++
			continue
		}

		if c == '"' {
			if !inString {
				inString = true
				// Determine if this is a key or value
				isKey = !afterColon && !isAfterArrayStart(s, i)
				afterColon = false
				// Find the end of the string
				end := findStringEnd(s, i+1)
				if end > i {
					str := s[i : end+1]
					if isKey {
						result.WriteString(keyStyle.Render(str))
					} else {
						result.WriteString(stringStyle.Render(str))
					}
					i = end + 1
					inString = false
					continue
				}
			}
			i++
			continue
		}

		if !inString {
			// Handle structural characters
			if c == ':' {
				result.WriteByte(c)
				afterColon = true
				i++
				continue
			}
			if c == ',' || c == '\n' {
				result.WriteByte(c)
				afterColon = false
				i++
				continue
			}
			if c == '{' || c == '}' || c == '[' || c == ']' {
				result.WriteString(bracketStyle.Render(string(c)))
				if c == '[' || c == '{' {
					afterColon = false
				}
				i++
				continue
			}
			// Handle numbers
			if (c >= '0' && c <= '9') || c == '-' {
				end := i
				for end < len(s) && (s[end] >= '0' && s[end] <= '9' || s[end] == '.' || s[end] == '-' || s[end] == 'e' || s[end] == 'E' || s[end] == '+') {
					end++
				}
				result.WriteString(numberStyle.Render(s[i:end]))
				i = end
				afterColon = false
				continue
			}
			// Handle true/false
			if strings.HasPrefix(s[i:], "true") {
				result.WriteString(boolStyle.Render("true"))
				i += 4
				afterColon = false
				continue
			}
			if strings.HasPrefix(s[i:], "false") {
				result.WriteString(boolStyle.Render("false"))
				i += 5
				afterColon = false
				continue
			}
			// Handle null
			if strings.HasPrefix(s[i:], "null") {
				result.WriteString(nullStyle.Render("null"))
				i += 4
				afterColon = false
				continue
			}
		}

		result.WriteByte(c)
		i++
	}

	return result.String()
}

// findStringEnd finds the closing quote of a JSON string
func findStringEnd(s string, start int) int {
	for i := start; i < len(s); i++ {
		if s[i] == '\\' {
			i++ // Skip escaped character
			continue
		}
		if s[i] == '"' {
			return i
		}
	}
	return -1
}

// isAfterArrayStart checks if we're inside an array (value context)
func isAfterArrayStart(s string, pos int) bool {
	// Look backwards for the last structural character
	for i := pos - 1; i >= 0; i-- {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		if c == '[' || c == ',' {
			// Check if we're in an array context
			return isInArrayContext(s, i)
		}
		return false
	}
	return false
}

// isInArrayContext checks if position is within an array
func isInArrayContext(s string, pos int) bool {
	bracketCount := 0
	braceCount := 0
	for i := pos; i >= 0; i-- {
		c := s[i]
		switch c {
		case ']':
			bracketCount++
		case '[':
			if bracketCount > 0 {
				bracketCount--
			} else {
				return true
			}
		case '}':
			braceCount++
		case '{':
			if braceCount > 0 {
				braceCount--
			} else {
				return false
			}
		}
	}
	return false
}

func (m Model) renderModal(content string) string {
	modalWidth := 60
	if m.Width-10 < 60 {
		modalWidth = m.Width - 10
	}
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modalStyle.Render(content))
}

func (m Model) renderModalWide(content string) string {
	modalWidth := 90
	if m.Width-10 < 90 {
		modalWidth = m.Width - 10
	}
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modalStyle.Render(content))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

type logEntry struct {
	Time  string
	Level string
	Msg   string
}

func parseLogEntry(logLine string) logEntry {
	entry := logEntry{}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(logLine), &data); err != nil {
		entry.Msg = logLine
		return entry
	}

	if t, ok := data["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, t); err == nil {
			entry.Time = parsed.Format("15:04:05")
		} else {
			entry.Time = t
		}
	}
	if l, ok := data["level"].(string); ok {
		entry.Level = strings.ToUpper(l)
	}
	if m, ok := data["msg"].(string); ok {
		entry.Msg = m
	}

	return entry
}
