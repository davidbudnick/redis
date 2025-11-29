package types

// Screen represents the current view in the application
type Screen int

const (
	ScreenConnections Screen = iota
	ScreenAddConnection
	ScreenEditConnection
	ScreenKeys
	ScreenKeyDetail
	ScreenAddKey
	ScreenHelp
	ScreenConfirmDelete
	ScreenServerInfo
	ScreenTTLEditor
	ScreenEditValue
	ScreenAddToCollection
	ScreenRemoveFromCollection
	ScreenRenameKey
	ScreenCopyKey
	ScreenPubSub
	ScreenPublishMessage
	ScreenSwitchDB
	ScreenSearchValues
	ScreenExport
	ScreenImport
	ScreenSlowLog
	ScreenLuaScript
	ScreenTestConnection
	ScreenLogs

	// New screens for additional features
	ScreenBulkDelete
	ScreenFavorites
	ScreenTreeView
	ScreenRecentKeys
	ScreenRegexSearch
	ScreenFuzzySearch
	ScreenWatchKey
	ScreenClientList
	ScreenMemoryStats
	ScreenSSHTunnel
	ScreenTLSConfig
	ScreenConnectionGroups
	ScreenClusterInfo
	ScreenCompareKeys
	ScreenBatchTTL
	ScreenKeyTemplate
	ScreenTemplates
	ScreenThemeSettings
	ScreenThemeSelect
	ScreenJSONPath
	ScreenKeyBindings
	ScreenValueHistory
	ScreenKeyspaceEvents
	ScreenExpiringKeys
)

// ScreenName returns a human-readable name for the screen
func (s Screen) String() string {
	names := map[Screen]string{
		ScreenConnections:          "Connections",
		ScreenAddConnection:        "Add Connection",
		ScreenEditConnection:       "Edit Connection",
		ScreenKeys:                 "Keys",
		ScreenKeyDetail:            "Key Detail",
		ScreenAddKey:               "Add Key",
		ScreenHelp:                 "Help",
		ScreenConfirmDelete:        "Confirm Delete",
		ScreenServerInfo:           "Server Info",
		ScreenTTLEditor:            "TTL Editor",
		ScreenEditValue:            "Edit Value",
		ScreenAddToCollection:      "Add to Collection",
		ScreenRemoveFromCollection: "Remove from Collection",
		ScreenRenameKey:            "Rename Key",
		ScreenCopyKey:              "Copy Key",
		ScreenPubSub:               "Pub/Sub",
		ScreenPublishMessage:       "Publish Message",
		ScreenSwitchDB:             "Switch Database",
		ScreenSearchValues:         "Search Values",
		ScreenExport:               "Export",
		ScreenImport:               "Import",
		ScreenSlowLog:              "Slow Log",
		ScreenLuaScript:            "Lua Script",
		ScreenTestConnection:       "Test Connection",
		ScreenLogs:                 "Logs",
		ScreenBulkDelete:           "Bulk Delete",
		ScreenFavorites:            "Favorites",
		ScreenTreeView:             "Tree View",
		ScreenRecentKeys:           "Recent Keys",
		ScreenRegexSearch:          "Regex Search",
		ScreenFuzzySearch:          "Fuzzy Search",
		ScreenWatchKey:             "Watch Key",
		ScreenClientList:           "Client List",
		ScreenMemoryStats:          "Memory Stats",
		ScreenSSHTunnel:            "SSH Tunnel",
		ScreenTLSConfig:            "TLS Config",
		ScreenConnectionGroups:     "Connection Groups",
		ScreenClusterInfo:          "Cluster Info",
		ScreenCompareKeys:          "Compare Keys",
		ScreenBatchTTL:             "Batch TTL",
		ScreenKeyTemplate:          "Key Template",
		ScreenTemplates:            "Templates",
		ScreenThemeSettings:        "Theme Settings",
		ScreenThemeSelect:          "Theme Select",
		ScreenJSONPath:             "JSON Path",
		ScreenKeyBindings:          "Key Bindings",
		ScreenValueHistory:         "Value History",
		ScreenKeyspaceEvents:       "Keyspace Events",
		ScreenExpiringKeys:         "Expiring Keys",
	}
	if name, ok := names[s]; ok {
		return name
	}
	return "Unknown"
}
