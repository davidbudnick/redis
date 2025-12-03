package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidbudnick/redis/internal/types"
)

// getScreenView returns the view function for a given screen
func (m Model) getScreenView() string {
	viewMap := map[types.Screen]func() string{
		types.ScreenConnections:          m.viewConnections,
		types.ScreenAddConnection:        m.viewAddConnection,
		types.ScreenEditConnection:       m.viewEditConnection,
		types.ScreenKeys:                 m.viewKeys,
		types.ScreenKeyDetail:            m.viewKeyDetail,
		types.ScreenAddKey:               m.viewAddKey,
		types.ScreenHelp:                 m.viewHelp,
		types.ScreenConfirmDelete:        m.viewConfirmDelete,
		types.ScreenServerInfo:           m.viewServerInfo,
		types.ScreenTTLEditor:            m.viewTTLEditor,
		types.ScreenEditValue:            m.viewEditValue,
		types.ScreenAddToCollection:      m.viewAddToCollection,
		types.ScreenRemoveFromCollection: m.viewRemoveFromCollection,
		types.ScreenRenameKey:            m.viewRenameKey,
		types.ScreenCopyKey:              m.viewCopyKey,
		types.ScreenPubSub:               m.viewPubSub,
		types.ScreenPublishMessage:       m.viewPubSub,
		types.ScreenSwitchDB:             m.viewSwitchDB,
		types.ScreenSearchValues:         m.viewSearchValues,
		types.ScreenExport:               m.viewExport,
		types.ScreenImport:               m.viewImport,
		types.ScreenSlowLog:              m.viewSlowLog,
		types.ScreenLuaScript:            m.viewLuaScript,
		types.ScreenTestConnection:       m.viewTestConnection,
		types.ScreenLogs:                 m.viewLogs,
		types.ScreenBulkDelete:           m.viewBulkDelete,
		types.ScreenBatchTTL:             m.viewBatchTTL,
		types.ScreenFavorites:            m.viewFavorites,
		types.ScreenRecentKeys:           m.viewRecentKeys,
		types.ScreenTreeView:             m.viewTreeView,
		types.ScreenRegexSearch:          m.viewRegexSearch,
		types.ScreenFuzzySearch:          m.viewFuzzySearch,
		types.ScreenClientList:           m.viewClientList,
		types.ScreenMemoryStats:          m.viewMemoryStats,
		types.ScreenClusterInfo:          m.viewClusterInfo,
		types.ScreenCompareKeys:          m.viewCompareKeys,
		types.ScreenTemplates:            m.viewTemplates,
		types.ScreenValueHistory:         m.viewValueHistory,
		types.ScreenKeyspaceEvents:       m.viewKeyspaceEvents,
		types.ScreenJSONPath:             m.viewJSONPath,
		types.ScreenExpiringKeys:         m.viewExpiringKeys,
		types.ScreenLiveMetrics:          m.viewLiveMetrics,
	}

	if viewFunc, ok := viewMap[m.Screen]; ok {
		return viewFunc()
	}
	return ""
}

func (m Model) View() string {
	if m.Width < 50 || m.Height < 15 {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
			"Terminal too small.\nResize to at least 50x15.")
	}

	content := m.getScreenView()

	// Status bar
	status := m.getStatusBar()

	fullContent := content + "\n\n" + status

	// Use PlaceHorizontal and PlaceVertical with whitespace to ensure full screen clear
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, fullContent,
		lipgloss.WithWhitespaceChars(" "))
}

func (m Model) getStatusBar() string {
	if m.Loading {
		return dimStyle.Render("Loading...")
	}
	if m.StatusMsg != "" {
		if strings.HasPrefix(m.StatusMsg, "Error") {
			return errorStyle.Render(m.StatusMsg)
		}
		return successStyle.Render(m.StatusMsg)
	}
	return ""
}
