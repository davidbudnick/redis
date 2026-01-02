package cmd

import (
	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

// Favorites

func LoadFavoritesCmd(connID int64) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.FavoritesLoadedMsg{Err: nil}
		}
		favorites := Config.ListFavorites(connID)
		return types.FavoritesLoadedMsg{Favorites: favorites, Err: nil}
	}
}

func AddFavoriteCmd(connID int64, key, label string) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.FavoriteAddedMsg{Err: nil}
		}
		fav, err := Config.AddFavorite(connID, key, label)
		return types.FavoriteAddedMsg{Favorite: fav, Err: err}
	}
}

func RemoveFavoriteCmd(connID int64, key string) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.FavoriteRemovedMsg{Err: nil}
		}
		err := Config.RemoveFavorite(connID, key)
		return types.FavoriteRemovedMsg{Key: key, Err: err}
	}
}

// Recent keys

func LoadRecentKeysCmd(connID int64) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.RecentKeysLoadedMsg{Err: nil}
		}
		keys := Config.ListRecentKeys(connID)
		return types.RecentKeysLoadedMsg{Keys: keys, Err: nil}
	}
}

func AddRecentKeyCmd(connID int64, key string, keyType types.KeyType) tea.Cmd {
	return func() tea.Msg {
		if Config != nil {
			Config.AddRecentKey(connID, key, keyType)
		}
		return nil
	}
}

// Templates

func LoadTemplatesCmd() tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.TemplatesLoadedMsg{Err: nil}
		}
		templates := Config.ListTemplates()
		return types.TemplatesLoadedMsg{Templates: templates, Err: nil}
	}
}
