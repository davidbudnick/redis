// Package testutil provides testing utilities and helpers for the redis-tui application.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbudnick/redis-tui/internal/db"
	"github.com/davidbudnick/redis-tui/internal/types"
)

// TempConfigPath creates a temporary config file path for testing.
// The file and directory will be cleaned up after the test.
func TempConfigPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "config.json")
}

// NewTestConfig creates a new Config instance using a temporary file for testing.
func NewTestConfig(t *testing.T) *db.Config {
	t.Helper()
	path := TempConfigPath(t)
	cfg, err := db.NewConfig(path)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}
	return cfg
}

// MustAddConnection adds a connection to the config or fails the test.
func MustAddConnection(t *testing.T, cfg *db.Config, name, host string, port int, password string, dbNum int) types.Connection {
	t.Helper()
	conn, err := cfg.AddConnection(name, host, port, password, dbNum)
	if err != nil {
		t.Fatalf("failed to add connection: %v", err)
	}
	return conn
}

// AssertConnectionExists checks that a connection with the given ID exists.
func AssertConnectionExists(t *testing.T, cfg *db.Config, id int64) types.Connection {
	t.Helper()
	connections, err := cfg.ListConnections()
	if err != nil {
		t.Fatalf("failed to list connections: %v", err)
	}
	for _, conn := range connections {
		if conn.ID == id {
			return conn
		}
	}
	t.Fatalf("connection with ID %d not found", id)
	return types.Connection{}
}

// AssertConnectionNotExists checks that a connection with the given ID does not exist.
func AssertConnectionNotExists(t *testing.T, cfg *db.Config, id int64) {
	t.Helper()
	connections, err := cfg.ListConnections()
	if err != nil {
		t.Fatalf("failed to list connections: %v", err)
	}
	for _, conn := range connections {
		if conn.ID == id {
			t.Fatalf("connection with ID %d should not exist", id)
		}
	}
}

// AssertEqual checks if two values are equal.
func AssertEqual[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

// AssertNoError checks that an error is nil.
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertError checks that an error is not nil.
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got nil", msg)
	}
}

// AssertSliceLen checks that a slice has the expected length.
func AssertSliceLen[T any](t *testing.T, slice []T, expectedLen int, msg string) {
	t.Helper()
	if len(slice) != expectedLen {
		t.Errorf("%s: got slice length %d, want %d", msg, len(slice), expectedLen)
	}
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SampleConnection returns a sample connection for testing.
func SampleConnection() types.Connection {
	return types.Connection{
		ID:       1,
		Name:     "test-redis",
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}
}

// SampleRedisKey returns a sample Redis key for testing.
func SampleRedisKey(name string, keyType types.KeyType) types.RedisKey {
	return types.RedisKey{
		Key:  name,
		Type: keyType,
		TTL:  -1, // No expiration
	}
}

// SampleFavorite returns a sample favorite for testing.
func SampleFavorite(connID int64, key string) types.Favorite {
	return types.Favorite{
		ConnectionID: connID,
		Key:          key,
		Label:        "Test Favorite",
	}
}
