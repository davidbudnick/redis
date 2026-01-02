package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidbudnick/redis/internal/types"
)

func TestNewConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, err := NewConfig(path)
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}

	// Check defaults
	if cfg.TreeSeparator != ":" {
		t.Errorf("TreeSeparator = %q, want \":\"", cfg.TreeSeparator)
	}
	if cfg.MaxRecentKeys != 20 {
		t.Errorf("MaxRecentKeys = %d, want 20", cfg.MaxRecentKeys)
	}
	if cfg.MaxValueHistory != 50 {
		t.Errorf("MaxValueHistory = %d, want 50", cfg.MaxValueHistory)
	}
	if cfg.WatchInterval != 1000 {
		t.Errorf("WatchInterval = %d, want 1000", cfg.WatchInterval)
	}
}

func TestNewConfig_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nestedPath := filepath.Join(dir, "subdir", "config.json")

	_, err := NewConfig(nestedPath)
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(filepath.Dir(nestedPath)); os.IsNotExist(err) {
		t.Error("NewConfig did not create directory")
	}
}

func TestConfig_AddConnection(t *testing.T) {
	cfg := newTestConfig(t)

	conn, err := cfg.AddConnection("test", "localhost", 6379, "secret", 0)
	if err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}

	if conn.ID == 0 {
		t.Error("Connection ID should not be 0")
	}
	if conn.Name != "test" {
		t.Errorf("Name = %q, want \"test\"", conn.Name)
	}
	if conn.Host != "localhost" {
		t.Errorf("Host = %q, want \"localhost\"", conn.Host)
	}
	if conn.Port != 6379 {
		t.Errorf("Port = %d, want 6379", conn.Port)
	}
	if conn.Password != "secret" {
		t.Errorf("Password = %q, want \"secret\"", conn.Password)
	}
	if conn.DB != 0 {
		t.Errorf("DB = %d, want 0", conn.DB)
	}
	if conn.Created.IsZero() {
		t.Error("Created should not be zero")
	}
}

func TestConfig_AddConnection_IncrementingIDs(t *testing.T) {
	cfg := newTestConfig(t)

	conn1, _ := cfg.AddConnection("test1", "localhost", 6379, "", 0)
	conn2, _ := cfg.AddConnection("test2", "localhost", 6380, "", 0)
	conn3, _ := cfg.AddConnection("test3", "localhost", 6381, "", 0)

	if conn2.ID <= conn1.ID {
		t.Errorf("conn2.ID (%d) should be greater than conn1.ID (%d)", conn2.ID, conn1.ID)
	}
	if conn3.ID <= conn2.ID {
		t.Errorf("conn3.ID (%d) should be greater than conn2.ID (%d)", conn3.ID, conn2.ID)
	}
}

func TestConfig_ListConnections(t *testing.T) {
	cfg := newTestConfig(t)

	// Add connections in non-alphabetical order
	cfg.AddConnection("zebra", "localhost", 6379, "", 0)
	cfg.AddConnection("alpha", "localhost", 6380, "", 0)
	cfg.AddConnection("beta", "localhost", 6381, "", 0)

	connections, err := cfg.ListConnections()
	if err != nil {
		t.Fatalf("ListConnections failed: %v", err)
	}

	if len(connections) != 3 {
		t.Fatalf("Expected 3 connections, got %d", len(connections))
	}

	// Check sorted by name
	if connections[0].Name != "alpha" {
		t.Errorf("First connection = %q, want \"alpha\"", connections[0].Name)
	}
	if connections[1].Name != "beta" {
		t.Errorf("Second connection = %q, want \"beta\"", connections[1].Name)
	}
	if connections[2].Name != "zebra" {
		t.Errorf("Third connection = %q, want \"zebra\"", connections[2].Name)
	}
}

func TestConfig_UpdateConnection(t *testing.T) {
	cfg := newTestConfig(t)

	conn, _ := cfg.AddConnection("original", "localhost", 6379, "old", 0)
	originalCreated := conn.Created

	updated, err := cfg.UpdateConnection(conn.ID, "updated", "newhost", 6380, "new", 1)
	if err != nil {
		t.Fatalf("UpdateConnection failed: %v", err)
	}

	if updated.Name != "updated" {
		t.Errorf("Name = %q, want \"updated\"", updated.Name)
	}
	if updated.Host != "newhost" {
		t.Errorf("Host = %q, want \"newhost\"", updated.Host)
	}
	if updated.Port != 6380 {
		t.Errorf("Port = %d, want 6380", updated.Port)
	}
	if updated.Password != "new" {
		t.Errorf("Password = %q, want \"new\"", updated.Password)
	}
	if updated.DB != 1 {
		t.Errorf("DB = %d, want 1", updated.DB)
	}
	if !updated.Created.Equal(originalCreated) {
		t.Error("Created timestamp should not change")
	}
	if !updated.Updated.After(updated.Created) {
		t.Error("Updated should be after Created")
	}
}

func TestConfig_UpdateConnection_NotFound(t *testing.T) {
	cfg := newTestConfig(t)

	_, err := cfg.UpdateConnection(999, "test", "localhost", 6379, "", 0)
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestConfig_DeleteConnection(t *testing.T) {
	cfg := newTestConfig(t)

	conn, _ := cfg.AddConnection("test", "localhost", 6379, "", 0)

	err := cfg.DeleteConnection(conn.ID)
	if err != nil {
		t.Fatalf("DeleteConnection failed: %v", err)
	}

	connections, _ := cfg.ListConnections()
	if len(connections) != 0 {
		t.Errorf("Expected 0 connections, got %d", len(connections))
	}
}

func TestConfig_DeleteConnection_NotFound(t *testing.T) {
	cfg := newTestConfig(t)

	err := cfg.DeleteConnection(999)
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestConfig_Favorites(t *testing.T) {
	cfg := newTestConfig(t)

	// Add favorite
	fav, err := cfg.AddFavorite(1, "user:123", "Test User")
	if err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}
	if fav.Key != "user:123" {
		t.Errorf("Key = %q, want \"user:123\"", fav.Key)
	}

	// Check is favorite
	if !cfg.IsFavorite(1, "user:123") {
		t.Error("IsFavorite should return true")
	}
	if cfg.IsFavorite(1, "other:key") {
		t.Error("IsFavorite should return false for non-favorite")
	}

	// List favorites
	favs := cfg.ListFavorites(1)
	if len(favs) != 1 {
		t.Errorf("Expected 1 favorite, got %d", len(favs))
	}

	// Remove favorite
	err = cfg.RemoveFavorite(1, "user:123")
	if err != nil {
		t.Fatalf("RemoveFavorite failed: %v", err)
	}

	if cfg.IsFavorite(1, "user:123") {
		t.Error("IsFavorite should return false after removal")
	}
}

func TestConfig_Favorites_NoDuplicates(t *testing.T) {
	cfg := newTestConfig(t)

	// Add same favorite twice
	cfg.AddFavorite(1, "user:123", "Label 1")
	cfg.AddFavorite(1, "user:123", "Label 2")

	favs := cfg.ListFavorites(1)
	if len(favs) != 1 {
		t.Errorf("Expected 1 favorite (no duplicates), got %d", len(favs))
	}
}

func TestConfig_RecentKeys(t *testing.T) {
	cfg := newTestConfig(t)

	// Add recent keys
	cfg.AddRecentKey(1, "key1", types.KeyTypeString)
	cfg.AddRecentKey(1, "key2", types.KeyTypeHash)
	cfg.AddRecentKey(1, "key3", types.KeyTypeList)

	recents := cfg.ListRecentKeys(1)
	if len(recents) != 3 {
		t.Errorf("Expected 3 recent keys, got %d", len(recents))
	}

	// Most recent should be first
	if recents[0].Key != "key3" {
		t.Errorf("Most recent key = %q, want \"key3\"", recents[0].Key)
	}
}

func TestConfig_RecentKeys_MaxLimit(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.MaxRecentKeys = 3

	// Add more than max
	for i := 0; i < 5; i++ {
		cfg.AddRecentKey(1, "key"+string(rune('a'+i)), types.KeyTypeString)
	}

	recents := cfg.ListRecentKeys(1)
	if len(recents) != 3 {
		t.Errorf("Expected max 3 recent keys, got %d", len(recents))
	}
}

func TestConfig_RecentKeys_MovesToFront(t *testing.T) {
	cfg := newTestConfig(t)

	cfg.AddRecentKey(1, "key1", types.KeyTypeString)
	cfg.AddRecentKey(1, "key2", types.KeyTypeString)
	cfg.AddRecentKey(1, "key1", types.KeyTypeString) // Re-add key1

	recents := cfg.ListRecentKeys(1)
	if recents[0].Key != "key1" {
		t.Errorf("Most recent key = %q, want \"key1\"", recents[0].Key)
	}
	if len(recents) != 2 {
		t.Errorf("Expected 2 recent keys (no duplicates), got %d", len(recents))
	}
}

func TestConfig_Templates(t *testing.T) {
	cfg := newTestConfig(t)

	// List default templates
	templates := cfg.ListTemplates()
	if len(templates) == 0 {
		t.Fatal("Expected default templates")
	}

	// Add new template
	newTemplate := types.KeyTemplate{
		Name:       "Custom",
		KeyPattern: "custom:{id}",
		Type:       types.KeyTypeString,
	}
	err := cfg.AddTemplate(newTemplate)
	if err != nil {
		t.Fatalf("AddTemplate failed: %v", err)
	}

	templates = cfg.ListTemplates()
	found := false
	for _, tmpl := range templates {
		if tmpl.Name == "Custom" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Custom template not found")
	}

	// Delete template
	err = cfg.DeleteTemplate("Custom")
	if err != nil {
		t.Fatalf("DeleteTemplate failed: %v", err)
	}
}

func TestConfig_ValueHistory(t *testing.T) {
	cfg := newTestConfig(t)

	value := types.RedisValue{
		Type:        types.KeyTypeString,
		StringValue: "test value",
	}

	cfg.AddValueHistory("user:123", value, "update")

	history := cfg.GetValueHistory("user:123")
	if len(history) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(history))
	}
	if history[0].Action != "update" {
		t.Errorf("Action = %q, want \"update\"", history[0].Action)
	}
}

func TestConfig_ValueHistory_MaxLimit(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.MaxValueHistory = 3

	value := types.RedisValue{Type: types.KeyTypeString, StringValue: "test"}

	for i := 0; i < 5; i++ {
		cfg.AddValueHistory("key", value, "action")
	}

	// Note: GetValueHistory filters by key, but the max limit applies globally
	// The implementation stores all history together
	if len(cfg.ValueHistory) > 3 {
		t.Errorf("Expected max 3 history entries, got %d", len(cfg.ValueHistory))
	}
}

func TestConfig_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Create config and add data
	cfg1, _ := NewConfig(path)
	cfg1.AddConnection("test", "localhost", 6379, "pass", 0)
	cfg1.AddFavorite(1, "key1", "label")

	// Create new config from same file
	cfg2, err := NewConfig(path)
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	// Verify data persisted
	connections, _ := cfg2.ListConnections()
	if len(connections) != 1 {
		t.Errorf("Expected 1 connection after reload, got %d", len(connections))
	}
	if connections[0].Name != "test" {
		t.Errorf("Connection name = %q, want \"test\"", connections[0].Name)
	}
}

func TestConfig_TreeSeparator(t *testing.T) {
	cfg := newTestConfig(t)

	// Default
	if cfg.GetTreeSeparator() != ":" {
		t.Errorf("Default separator = %q, want \":\"", cfg.GetTreeSeparator())
	}

	// Set new separator
	err := cfg.SetTreeSeparator("/")
	if err != nil {
		t.Fatalf("SetTreeSeparator failed: %v", err)
	}

	if cfg.GetTreeSeparator() != "/" {
		t.Errorf("Separator = %q, want \"/\"", cfg.GetTreeSeparator())
	}
}

func TestConfig_WatchInterval(t *testing.T) {
	cfg := newTestConfig(t)

	interval := cfg.GetWatchInterval()
	expected := time.Duration(1000) * time.Millisecond
	if interval != expected {
		t.Errorf("WatchInterval = %v, want %v", interval, expected)
	}
}

func TestConfig_KeyBindings(t *testing.T) {
	cfg := newTestConfig(t)

	// Get default bindings
	bindings := cfg.GetKeyBindings()
	if bindings.Quit == "" {
		t.Error("Quit keybinding should not be empty")
	}

	// Modify and save
	bindings.Quit = "ctrl+x"
	err := cfg.SetKeyBindings(bindings)
	if err != nil {
		t.Fatalf("SetKeyBindings failed: %v", err)
	}

	if cfg.GetKeyBindings().Quit != "ctrl+x" {
		t.Errorf("Quit = %q, want \"ctrl+x\"", cfg.GetKeyBindings().Quit)
	}

	// Reset
	err = cfg.ResetKeyBindings()
	if err != nil {
		t.Fatalf("ResetKeyBindings failed: %v", err)
	}

	if cfg.GetKeyBindings().Quit == "ctrl+x" {
		t.Error("Keybindings should be reset to defaults")
	}
}

func TestConfig_Groups(t *testing.T) {
	cfg := newTestConfig(t)

	// Add group
	err := cfg.AddGroup("Production", "#ff0000")
	if err != nil {
		t.Fatalf("AddGroup failed: %v", err)
	}

	groups := cfg.ListGroups()
	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}
	if groups[0].Name != "Production" {
		t.Errorf("Group name = %q, want \"Production\"", groups[0].Name)
	}

	// Add connection to group
	conn, _ := cfg.AddConnection("test", "localhost", 6379, "", 0)
	err = cfg.AddConnectionToGroup("Production", conn.ID)
	if err != nil {
		t.Fatalf("AddConnectionToGroup failed: %v", err)
	}

	groups = cfg.ListGroups()
	if len(groups[0].Connections) != 1 {
		t.Errorf("Expected 1 connection in group, got %d", len(groups[0].Connections))
	}

	// Remove connection from group
	err = cfg.RemoveConnectionFromGroup("Production", conn.ID)
	if err != nil {
		t.Fatalf("RemoveConnectionFromGroup failed: %v", err)
	}

	groups = cfg.ListGroups()
	if len(groups[0].Connections) != 0 {
		t.Errorf("Expected 0 connections in group, got %d", len(groups[0].Connections))
	}
}

func TestDefaultTemplates(t *testing.T) {
	templates := defaultTemplates()

	if len(templates) == 0 {
		t.Fatal("Expected default templates")
	}

	// Check required templates exist
	requiredNames := []string{"Session", "Cache", "Rate Limit", "Queue", "Leaderboard"}
	for _, name := range requiredNames {
		found := false
		for _, tmpl := range templates {
			if tmpl.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing required template: %q", name)
		}
	}
}

// newTestConfig creates a config for testing with a temp file
func newTestConfig(t *testing.T) *Config {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg, err := NewConfig(path)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}
	return cfg
}
