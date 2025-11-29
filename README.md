# Redis TUI Manager

[![CI](https://github.com/davidbudnick/redis/actions/workflows/ci.yml/badge.svg)](https://github.com/davidbudnick/redis/actions/workflows/ci.yml)
[![Release](https://github.com/davidbudnick/redis/actions/workflows/release.yml/badge.svg)](https://github.com/davidbudnick/redis/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidbudnick/redis?v=1)](https://goreportcard.com/report/github.com/davidbudnick/redis)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful terminal user interface (TUI) for managing Redis databases, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

**Redis TUI** is a feature-rich Redis client for the terminal that lets you browse, edit, and manage your Redis keys with ease. Perfect for developers and DevOps engineers who prefer working in the command line.

## Quick Install

```bash
# Homebrew (macOS and Linux)
brew tap davidbudnick/tap
brew install redis-tui
```

## Why Redis TUI?

- **No GUI Required** - Manage Redis directly from your terminal over SSH
- **Fast and Lightweight** - Built in Go for speed and minimal resource usage
- **Full Redis Support** - Works with all Redis data types: strings, lists, sets, sorted sets, hashes, and streams
- **Secure Connections** - TLS/SSL and SSH tunnel support for secure access
- **Multiple Connections** - Save and switch between multiple Redis instances easily

## Screenshots

### Connection Management
```
          ██████╗ ███████╗██████╗ ██╗███████╗
          ██╔══██╗██╔════╝██╔══██╗██║██╔════╝
          ██████╔╝█████╗  ██║  ██║██║███████╗
          ██╔══██╗██╔══╝  ██║  ██║██║╚════██║
          ██║  ██║███████╗██████╔╝██║███████║
          ╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝╚══════╝

       ┌──────────────────┐ ┌──────────────────┐
       │ Connections      │ │ Time             │
       │ 4 saved          │ │ 14:32:15         │
       └──────────────────┘ └──────────────────┘

╭─ Saved Connections (4) ──────────────────────────────╮

╭────────────────────────────────────────────────────────╮
│ ● Production                                           │
│   redis.example.com:6379  [db0] [TLS]                  │
╰────────────────────────────────────────────────────────╯
╭────────────────────────────────────────────────────────╮
│ ○ Staging                                              │
│   staging-redis.local:6379  [db0]                      │
╰────────────────────────────────────────────────────────╯
╭────────────────────────────────────────────────────────╮
│ ○ Local Development                                    │
│   localhost:6379  [db0]                                │
╰────────────────────────────────────────────────────────╯

╰──────────────────────────────────────────────────────╯

[↑/↓] navigate  [enter] connect  [a] add  [e] edit  [d] delete  [q] quit
```

### Key Browser with Preview
```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│   Keys - Production [1234]                            │  Preview                                │
├───────────────────────────────────────────────────────┼─────────────────────────────────────────┤
│  Filter: user:*                                       │  ─────────────────────────────────────  │
│                                                       │                                         │
│   Key                            Type      TTL        │  Key: user:1001                         │
│  ──────────────────────────────────────────────────   │                                         │
│ ▶ user:1001                      string    ∞          │  Type: string                           │
│   user:1002                      string    1h         │                                         │
│   user:1003:profile              hash      ∞          │  TTL: No expiry                         │
│   user:1004:sessions             list      2h         │                                         │
│   user:1005:followers            set       ∞          │  ─────────────────────────────────────  │
│   user:1006:scores               zset      ∞          │                                         │
│   user:1007                      string    30m        │  Value                                  │
│   user:1008:cart                 hash      15m        │                                         │
│                                                       │  {                                      │
│                                                       │    "id": 1001,                          │
│  1-8 of 1234 • l:more                                 │    "name": "John Doe",                  │
│                                                       │    "email": "john@example.com"          │
│                                                       │  }                                      │
│                                                       │                                         │
│  j/k:nav  enter:view  a:add  d:del  /:filter  O:logs  i:info  q:back                            │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Hash Preview
```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│   Keys - Production [1234]                            │  Preview                                │
├───────────────────────────────────────────────────────┼─────────────────────────────────────────┤
│  Filter: *                                            │  ─────────────────────────────────────  │
│                                                       │                                         │
│   Key                            Type      TTL        │  Key: user:1003:profile                 │
│  ──────────────────────────────────────────────────   │                                         │
│   user:1001                      string    ∞          │  Type: hash                             │
│   user:1002                      string    1h         │                                         │
│ ▶ user:1003:profile              hash      ∞          │  TTL: No expiry                         │
│   user:1004:sessions             list      2h         │                                         │
│   user:1005:followers            set       ∞          │  ─────────────────────────────────────  │
│   user:1006:scores               zset      ∞          │                                         │
│                                                       │  Value                                  │
│                                                       │                                         │
│                                                       │  Fields: 5                              │
│                                                       │                                         │
│                                                       │  age: 30                                │
│                                                       │  city: New York                         │
│                                                       │  email: john@example.com                │
│                                                       │  name: John Doe                         │
│                                                       │  status: active                         │
│                                                       │                                         │
│  j/k:nav  enter:view  a:add  d:del  /:filter  O:logs  i:info  q:back                            │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Key Detail View
```
┌──────────────────────────────────────────────────────────────────┐
│                           Key Detail                             │
├──────────────────────────────────────────────────────────────────┤
│  Key: user:1001                                                  │
│  Type: string                                                    │
│  TTL: No expiry  Memory: 128 B                                   │
│                                                                  │
│  Value:                                                          │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ {                                                           │ │
│  │   "id": 1001,                                               │ │
│  │   "name": "John Doe",                                       │ │
│  │   "email": "john@example.com",                              │ │
│  │   "created_at": "2024-01-15T10:30:00Z"                      │ │
│  │ }                                                           │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  t:TTL  d:del  r:refresh  R:rename  c:copy  e:edit  esc:back     │
└──────────────────────────────────────────────────────────────────┘
```

### Server Info
```
┌────────────────────────────────────────┐
│            Server Info                 │
├────────────────────────────────────────┤
│  Version:     7.2.4                    │
│  Mode:        standalone               │
│  OS:          Linux 5.15.0             │
│  Memory:      1.2GB / 4GB              │
│  Clients:     42                       │
│  Keys:        125,432                  │
│  Uptime:      45 days, 12:34:56        │
│                                        │
│  r:refresh  esc:back                   │
└────────────────────────────────────────┘
```

### Tree View
```
┌──────────────────────────────────────────────────────────────────┐
│                           Tree View                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  v user: (1,234 keys)                                            │
│    v sessions: (456 keys)                                        │
│        session:abc123                                            │
│        session:def456                                            │
│    > profiles: (234 keys)                                        │
│    > settings: (89 keys)                                         │
│  > cache: (5,678 keys)                                           │
│  > queue: (123 keys)                                             │
│                                                                  │
│  j/k:nav  enter:expand/select  esc:back                          │
└──────────────────────────────────────────────────────────────────┘
```

## Features

### Core Features
- 🔌 **Connection Management** - Save and manage multiple Redis connections
- 🔑 **Key Browser** - Browse, search, and filter keys with pattern matching
- 📊 **Type Support** - Full support for all Redis data types (String, List, Set, Sorted Set, Hash, Stream)
- ✏️ **Edit Values** - Edit string values directly, add/remove items from collections
- ⏱️ **TTL Management** - View and set TTL with live countdown
- 📋 **Export/Import** - Export keys to JSON and import from files
- 🔍 **Search** - Search keys by name pattern or value content

### Advanced Features
- ⭐ **Favorites** - Mark frequently used keys as favorites
- 🕒 **Recent Keys** - Quick access to recently viewed keys
- 🌳 **Tree View** - Browse keys in a hierarchical tree structure
- 🔎 **Regex Search** - Search keys using regular expressions
- 🔍 **Fuzzy Search** - Find keys with fuzzy matching
- 👀 **Watch Mode** - Monitor key values for changes in real-time
- 🗑️ **Bulk Delete** - Delete multiple keys matching a pattern
- ⏲️ **Batch TTL** - Set TTL on multiple keys at once
- ⚖️ **Compare Keys** - Compare values between two keys
- 📝 **Key Templates** - Create new keys from predefined templates
- 📜 **Value History** - View and restore previous values
- 📡 **Keyspace Events** - Subscribe to keyspace notifications
- 👥 **Client List** - View connected Redis clients
- 📊 **Memory Stats** - Detailed memory usage statistics
- 🌐 **Cluster Support** - View cluster node information
- ⌨️ **Customizable Keybindings** - Configure your own keyboard shortcuts
- 📋 **Clipboard Support** - Copy values to clipboard
- 🔐 **TLS Support** - Connect with TLS/SSL encryption
- 🚇 **SSH Tunneling** - Connect through SSH tunnels
- 📁 **Connection Groups** - Organize connections into groups

### Other Features
- 🔧 **Lua Scripts** - Execute Lua scripts directly
- 📨 **Pub/Sub** - Publish messages to channels
- 📈 **Slow Log** - View slow query log
- ℹ️ **Server Info** - View Redis server information
- 🗄️ **Database Switch** - Switch between Redis databases
- 📝 **Application Logs** - View internal application logs

## Installation

### Homebrew (macOS and Linux)

```bash
# Add the tap
brew tap davidbudnick/tap

# Install redis-tui
brew install redis-tui
```

### From Source

```bash
# Clone the repository
git clone https://github.com/davidbudnick/redis.git
cd redis

# Build
make build

# Install to GOPATH/bin
make install
```

### Using Go Install

```bash
go install github.com/davidbudnick/redis@latest
```

### Pre-built Binaries

Download the latest release from the [Releases](https://github.com/davidbudnick/redis/releases) page.

## Usage

```bash
# Run the application
redis-tui

# Or if installed via go install
redis
```

## Keyboard Shortcuts

### Global

| Key | Action |
| --- | --- |
| `q` | Quit / Go back |
| `?` | Show help |
| `j/k` | Navigate up/down |
| `Ctrl+C` | Force quit |

### Connections Screen

| Key | Action |
| --- | --- |
| `Enter` | Connect to selected |
| `a/n` | Add new connection |
| `e` | Edit connection |
| `d` | Delete connection |
| `r` | Refresh list |
| `Ctrl+T` | Test connection |

### Keys Screen

| Key | Action |
| --- | --- |
| `Enter` | View key details |
| `a/n` | Add new key |
| `d` | Delete key |
| `/` | Filter by pattern |
| `s/S` | Sort / Toggle direction |
| `v` | Search by value |
| `e` | Export to JSON |
| `I` | Import from JSON |
| `i` | Server info |
| `D` | Switch database |
| `O` | View logs |
| `B` | Bulk delete |
| `T` | Batch set TTL |
| `F` | View favorites |
| `W` | Tree view |
| `Ctrl+R` | Regex search |
| `Ctrl+F` | Fuzzy search |
| `Ctrl+L` | Client list |
| `M` | Memory stats |
| `C` | Cluster info |
| `P` | Key templates |

### Key Detail Screen

| Key | Action |
| --- | --- |
| `e` | Edit value (string) |
| `a` | Add to collection |
| `x` | Remove from collection |
| `t` | Set TTL |
| `R` | Rename key |
| `c` | Copy key |
| `d` | Delete key |
| `f` | Toggle favorite |
| `w` | Watch for changes |
| `h` | View value history |
| `y` | Copy to clipboard |
| `J` | JSON path query |

## Configuration

Configuration is stored in `~/.config/redis-tui/config.json`.

### Example Configuration

```json
{
  "connections": [
    {
      "id": 1,
      "name": "Local Redis",
      "host": "localhost",
      "port": 6379,
      "password": "",
      "db": 0,
      "use_tls": false
    },
    {
      "id": 2,
      "name": "Production",
      "host": "redis.example.com",
      "port": 6379,
      "password": "your-password",
      "db": 0,
      "use_tls": true
    }
  ],
  "key_bindings": {
    "up": "k",
    "down": "j",
    "select": "enter",
    "back": "esc",
    "quit": "q",
    "help": "?",
    "refresh": "r",
    "delete": "d",
    "add": "a",
    "edit": "e",
    "filter": "/",
    "server_info": "i",
    "export": "E",
    "import": "I"
  },
  "tree_separator": ":",
  "max_recent_keys": 20,
  "max_value_history": 50,
  "watch_interval_ms": 1000
}
```

### Connection Options

| Option | Description |
| --- | --- |
| `name` | Display name for the connection |
| `host` | Redis server hostname or IP |
| `port` | Redis server port (default: 6379) |
| `password` | Redis password (optional) |
| `db` | Redis database number (0-15) |
| `use_tls` | Enable TLS/SSL connection |
| `ssh_host` | SSH tunnel hostname (optional) |
| `ssh_user` | SSH tunnel username (optional) |
| `ssh_key_path` | Path to SSH private key (optional) |

### Custom Keybindings

Keybindings can be customized in the configuration file under the `key_bindings` section. All navigation and action keys can be remapped to your preference.

## Requirements

- Go 1.21 or later (for building from source)
- A terminal that supports 256 colors
- Redis server 4.0 or later

## Supported Platforms

- macOS (Intel and Apple Silicon)
- Linux (amd64, arm64)
- Windows (amd64)

## Development

```bash
# Install development dependencies
make dev-deps

# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Build for all platforms
make build-all
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling library
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [go-redis](https://github.com/redis/go-redis) - Redis client

## Related Projects

If you're looking for Redis tools, you might also be interested in:

- [redis-cli](https://redis.io/docs/ui/cli/) - Official Redis command line interface
- [RedisInsight](https://redis.com/redis-enterprise/redis-insight/) - Official Redis GUI

## Keywords

redis, redis-cli, redis-client, redis-tui, redis-gui, redis-manager, terminal, tui, cli, go, golang, database, key-value, cache, devops, sysadmin
