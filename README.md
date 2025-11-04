prefer.go
=========

[![Test](https://github.com/LimpidTech/prefer.go/actions/workflows/test.yml/badge.svg)](https://github.com/LimpidTech/prefer.go/actions/workflows/test.yml)
[![Lint](https://github.com/LimpidTech/prefer.go/actions/workflows/lint.yml/badge.svg)](https://github.com/LimpidTech/prefer.go/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/LimpidTech/prefer.go)](https://goreportcard.com/report/github.com/LimpidTech/prefer.go)
[![codecov](https://codecov.io/gh/LimpidTech/prefer.go/branch/master/graph/badge.svg)](https://codecov.io/gh/LimpidTech/prefer.go)

Powerful configuration management in Go

## Features

- Load configuration files from standard system paths
- Support for YAML, JSON, and XML formats
- File watching for automatic configuration reloading
- Cross-platform support (Unix, Linux, macOS, Windows)

## Installation

```bash
go get github.com/LimpidTech/prefer.go
```

## Usage

### Basic Loading

```go
package main

import (
    "fmt"
    "log"
    "github.com/LimpidTech/prefer.go"
)

type Config struct {
    Name string `yaml:"name"`
    Port int    `yaml:"port"`
}

func main() {
    var config Config
    
    // Load config from standard paths (./config.yaml, /etc/config.yaml, etc.)
    cfg, err := prefer.Load("config", &config)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Loaded from: %s\n", cfg.Identifier)
    fmt.Printf("Name: %s, Port: %d\n", config.Name, config.Port)
}
```

### File Watching

```go
package main

import (
    "fmt"
    "log"
    "github.com/LimpidTech/prefer.go"
)

type Config struct {
    Name string `yaml:"name"`
}

func main() {
    var config Config
    
    // Watch for configuration changes
    channel, err := prefer.Watch("config", &config)
    if err != nil {
        log.Fatal(err)
    }
    
    for updatedConfig := range channel {
        cfg := updatedConfig.(*Config)
        fmt.Printf("Config updated: %s\n", cfg.Name)
    }
}
```

## Supported Formats

- YAML (`.yaml`, `.yml`)
- JSON (`.json`)
- XML (`.xml`)
- INI (`.ini`)

## Standard Search Paths

The library searches for configuration files in the following locations (in order):

### Unix/Linux/macOS
- Current directory (`.`)
- Working directory
- `$XDG_CONFIG_DIRS`
- `$HOME/.config`
- `$HOME`
- `/usr/local/etc`
- `/usr/etc`
- `/etc`

### Windows
- Current directory
- `%USERPROFILE%`
- `%LOCALPROFILE%`
- `%APPDATA%`
- `%CommonProgramFiles%`
- `%ProgramData%`
- `%ProgramFiles%`
- `%SystemRoot%`

## Requirements

- Go 1.25 or later

## License

See LICENSE file for details.
