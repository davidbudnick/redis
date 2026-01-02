// Package cmd contains Bubble Tea commands for the Redis TUI.
//
// Commands are organized by domain:
//   - cmd_core.go: Global variables (Config, RedisClient)
//   - cmd_connection.go: Connection management commands
//   - cmd_keys.go: Key operations, TTL, watch, keyspace events
//   - cmd_collections.go: Collection add/remove operations
//   - cmd_server.go: Server info, flush, switch DB, slow log, Lua, pub/sub
//   - cmd_search.go: Search by value, regex, fuzzy, compare keys
//   - cmd_io.go: Export, import, clipboard operations
//   - cmd_features.go: Favorites, recent keys, templates
//   - commands.go: Dependency-injected command wrapper (for testing)
package cmd
