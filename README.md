# Redis TUI Manager

[![CI](https://github.com/davidbudnick/redis/actions/workflows/ci.yml/badge.svg)](https://github.com/davidbudnick/redis/actions/workflows/ci.yml)
[![Release](https://github.com/davidbudnick/redis/actions/workflows/release.yml/badge.svg)](https://github.com/davidbudnick/redis/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidbudnick/redis)](https://goreportcard.com/report/github.com/davidbudnick/redis)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful terminal user interface (TUI) for managing Redis databases, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Screenshots

### Connection Management
```
┌──────────────────────────────────────────────────────────────────┐
│                       Redis Connections                          │
├──────────────────────────────────────────────────────────────────┤
│   Name                 Host                      Port     DB     │
│  ─────────────────────────────────────────────────────────────   │
│ ▶ Production           redis.example.com         6379     0      │
│   Staging              staging-redis.local       6379     0      │
│   Local Development    localhost                 6379     0      │
│   Cache Server         cache.internal            6380     1      │
│                                                                  │
│  j/k:navigate  enter:connect  a:add  e:edit  d:delete  q:quit    │
└──────────────────────────────────────────────────────────────────┘
```

### Key Browser
```
┌──────────────────────────────────────────────────────────────────┐
│  Production (localhost:6379/0)                    Keys: 1,234    │
├──────────────────────────────────────────────────────────────────┤
│  Pattern: user:*                                                 │
│                                                                  │
│   Key                              Type      TTL        Size     │
│  ─────────────────────────────────────────────────────────────   │
│ ▶ user:1001                        string    -1         128B     │
│   user:1002                        string    3600       256B     │
│   user:1003:profile                hash      -1         512B     │
│   user:1004:sessions               list      7200       1.2KB    │
│   user:1005:followers              set       -1         2.4KB    │
│   user:1006:scores                 zset      -1         890B     │
│                                                                  │
│  [1-6 of 1234]  /:filter  a:add  d:delete  s:sort  ?:help        │
└──────────────────────────────────────────────────────────────────┘
```

### Key Detail View
```
┌──────────────────────────────────────────────────────────────────┐
│  Key: user:1001                                                  │
├──────────────────────────────────────────────────────────────────┤
│  Type: string                                                    │
│  TTL:  -1 (no expiry)                                            │
│  Size: 128 bytes                                                 │
│                                                                  │
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

### Hash View
```
┌──────────────────────────────────────────────────────────────────┐
│  Key: user:1003:profile                                          │
├──────────────────────────────────────────────────────────────────┤
│  Type: hash (5 fields)                                           │
│  TTL:  -1 (no expiry)                                            │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ Field          Value                                        │ │
│  │ ─────────────────────────────────────────────────────────── │ │
│  │ name           John Doe                                     │ │
│  │ email          john@example.com                             │ │
│  │ age            30                                           │ │
│  │ city           New York                                     │ │
│  │ status         active                                       │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  a:add  x:remove  t:TTL  d:del  esc:back                         │
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
│                        🌲 Tree View                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ▼ user: (1,234 keys)                                            │
│    ▼ sessions: (456 keys)                                        │
│        session:abc123                                            │
│        session:def456                                            │
│    ▶ profiles: (234 keys)                                        │
│    ▶ settings: (89 keys)                                         │
│  ▶ cache: (5,678 keys)                                           │
│  ▶ queue: (123 keys)                                             │
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
|-----|--------|
| `q` | Quit / Go back |
| `?` | Show help |
| `j/k` | Navigate up/down |
| `Ctrl+C` | Force quit |

### Connections Screen
| Key | Action |
|-----|--------|
| `Enter` | Connect to selected |
| `a/n` | Add new connection |
| `e` | Edit connection |
| `d` | Delete connection |
| `r` | Refresh list |
| `Ctrl+T` | Test connection |

### Keys Screen
| Key | Action |
|-----|--------|
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
|-----|--------|
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

### Custom Keybindings

Keybindings can be customized in the configuration file under the `keybindings` section.

## Requirements

- Go 1.21 or later
- A terminal that supports 256 colors

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
