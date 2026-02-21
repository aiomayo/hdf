# hdf

A smart process killer CLI tool. Find and terminate processes by port, name, PID, or user with interactive selection, glob patterns, and graceful shutdown support.

## Installation

### Install script (recommended)

```sh
curl -sSfL https://raw.githubusercontent.com/aiomayo/hdf/main/install.sh | sh
```

To install system-wide:

```sh
curl -sSfL https://raw.githubusercontent.com/aiomayo/hdf/main/install.sh | sudo sh -s -- -b /usr/local/bin
```

To install a specific version:

```sh
curl -sSfL https://raw.githubusercontent.com/aiomayo/hdf/main/install.sh | sh -s -- v0.0.1
```

### Install script for Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/aiomayo/hdf/main/install.ps1 | iex
```

### Go

```sh
go install github.com/aiomayo/hdf@latest
```

### Binary releases

Download pre-built binaries for Linux, macOS, and Windows from the [releases page](https://github.com/aiomayo/hdf/releases).

### Build from source

```sh
git clone https://github.com/aiomayo/hdf.git
cd hdf
go build -o hdf .
```

## Usage

```sh
# Kill process by port
hdf --port 8080

# Kill process by name (supports glob patterns)
hdf --name "node*"

# Kill process by PID
hdf --pid 1234

# Kill process by user
hdf --user root

# Dry run — preview without killing
hdf --port 8080 --dry-run

# Force kill
hdf --port 8080 --force
```

## Configuration

hdf uses a TOML config file that is auto-created with defaults on first run.

### Config file location

| OS      | Path                                             |
|---------|--------------------------------------------------|
| macOS   | `~/Library/Application Support/hdf/config.toml`  |
| Linux   | `~/.config/hdf/config.toml`                      |
| Windows | `%LOCALAPPDATA%\hdf\config.toml`                 |

The path follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/). Set `XDG_CONFIG_HOME` to override.

### Environment variables

All config options can be overridden with environment variables using the `HDF_` prefix:

```sh
HDF_DEFAULT_FORCE=true hdf 8080
HDF_DEFAULT_VERBOSE=true hdf node
```

### Options

#### `protected` - process protection list

Processes matching these names are skipped during kill operations. Comparison is case-insensitive.

```toml
protected = ['init', 'systemd', 'launchd', 'kernel_task', 'WindowServer', 'loginwindow', 'sshd']
```

#### `aliases` - query shortcuts

Map short names to longer queries. Aliases are resolved before query classification, so they work with ports, names, and patterns.

```toml
[aliases]
db = "postgres"
web = "nginx"
dev = "3000"
```

Usage:

```sh
hdf db        # equivalent to: hdf postgres
hdf web       # equivalent to: hdf nginx
hdf dev       # equivalent to: hdf 3000
```

#### `graceful_timeout` - default graceful shutdown timeout

Default timeout for `--graceful` mode before escalating to SIGKILL.

```toml
graceful_timeout = '5s'
```

#### `default_force` - always use SIGKILL

When `true`, hdf uses SIGKILL by default (equivalent to always passing `--force`).

```toml
default_force = false
```

#### `default_verbose` - enable verbose output

When `true`, hdf runs in verbose mode by default (equivalent to always passing `--verbose`).

```toml
default_verbose = false
```

### Default config

The auto-generated config file looks like this:

```toml
default_force = false
default_verbose = false
graceful_timeout = '5s'
protected = ['init', 'systemd', 'launchd', 'kernel_task', 'WindowServer', 'loginwindow', 'sshd']

[aliases]
```