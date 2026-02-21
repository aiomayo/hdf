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