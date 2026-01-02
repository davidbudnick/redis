package cmd

import (
	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func SearchByValueCmd(pattern, valueSearch string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeysLoadedMsg{Err: nil}
		}
		keys, err := RedisClient.SearchByValue(pattern, valueSearch, maxKeys)
		return types.KeysLoadedMsg{Keys: keys, Cursor: 0, Err: err}
	}
}

func RegexSearchCmd(pattern string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.RegexSearchResultMsg{Err: nil}
		}
		keys, err := RedisClient.ScanKeysWithRegex(pattern, maxKeys)
		return types.RegexSearchResultMsg{Keys: keys, Err: err}
	}
}

func FuzzySearchCmd(term string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.FuzzySearchResultMsg{Err: nil}
		}
		keys, err := RedisClient.FuzzySearchKeys(term, maxKeys)
		return types.FuzzySearchResultMsg{Keys: keys, Err: err}
	}
}

func CompareKeysCmd(key1, key2 string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.CompareKeysResultMsg{Err: nil}
		}
		val1, val2, err := RedisClient.CompareKeys(key1, key2)
		return types.CompareKeysResultMsg{Key1Value: val1, Key2Value: val2, Err: err}
	}
}
