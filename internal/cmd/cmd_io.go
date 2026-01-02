package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func ExportKeysCmd(pattern, filename string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ExportCompleteMsg{Filename: filename, Err: nil}
		}
		data, err := RedisClient.ExportKeys(pattern)
		if err != nil {
			return types.ExportCompleteMsg{Filename: filename, Err: err}
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return types.ExportCompleteMsg{Filename: filename, Err: err}
		}

		err = os.WriteFile(filename, jsonData, 0600)
		return types.ExportCompleteMsg{Filename: filename, KeyCount: len(data), Err: err}
	}
}

func ImportKeysCmd(filename string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ImportCompleteMsg{Filename: filename, Err: nil}
		}

		// Clean the file path to prevent directory traversal
		cleanPath := filepath.Clean(filename)
		jsonData, err := os.ReadFile(cleanPath) // #nosec G304 - user-provided import path is intentional
		if err != nil {
			return types.ImportCompleteMsg{Filename: filename, Err: err}
		}

		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return types.ImportCompleteMsg{Filename: filename, Err: err}
		}

		count, err := RedisClient.ImportKeys(data)
		return types.ImportCompleteMsg{Filename: filename, KeyCount: count, Err: err}
	}
}

func CopyToClipboardCmd(content string) tea.Cmd {
	return func() tea.Msg {
		// Use pbcopy on macOS
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(content)
		err := cmd.Run()
		return types.ClipboardCopiedMsg{Content: content, Err: err}
	}
}
