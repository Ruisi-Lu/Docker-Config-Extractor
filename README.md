# Docker Config Extractor

> This project is vibe code, be careful when using it!

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A powerful Go tool that extracts Docker container configurations and creates development containers with enhanced debugging capabilities. This tool inspects running containers, parses their configurations, and generates development-ready clones with customizable features.

## 🌟 Features

- **Container Configuration Extraction**: Automatically extracts complete container configurations using `docker inspect`
- **Development Container Creation**: Creates dev-ready containers based on production configurations
- **Clean Architecture**: Separates parsing, generation, and execution logic into reusable packages
- **Debugger Support**: Automatically installs and configures Delve debugger for Go applications
- **Custom Volume Mounting**: Supports dev-swap directories for development workflows
- **Comprehensive Parsing**: Handles all Docker configuration elements:
  - Environment variables
  - Volume mounts (bind and volume types, with ro/rw support)
  - Port mappings
  - Networks
  - Working directories
  - Labels
  - Devices
  - Extra hosts
  - Restart policies
  - EntryPoints and Commands

## 📦 Installation

### Prerequisites

- Go 1.25 or higher
- Docker installed and running
- Access to Docker daemon

### Build from Source

```bash
git clone https://github.com/yourusername/docker-config-extractor.git
cd docker-config-extractor
go build -o docker-config-extractor
```

### Install

```bash
go install github.com/yourusername/docker-config-extractor@latest
```

## 🚀 Quick Start

### Basic Usage

Extract configuration from a container and create a dev container:

```bash
./docker-config-extractor <container-name> [dev-container-name] [dev-swap-dir]
```

### Examples

**Simple clone:**
```bash
./docker-config-extractor myapp
# Creates: myapp-dev
```

**Custom dev container name:**
```bash
./docker-config-extractor myapp myapp-development
```

**With dev-swap directory:**
```bash
./docker-config-extractor myapp myapp-dev /path/to/dev-workspace
# Mounts /path/to/dev-workspace as /dev-swap in container
```

## 🏗️ Architecture

### Project Structure

```
docker-config-extractor/
├── main.go                          # Manager and CLI entry point
├── go.mod                           # Go module definition
└── pkg/
    └── containerconfig/
        ├── spec.go                  # ContainerSpec data structures
        ├── parser.go                # JSON parsing logic
        └── generator.go             # Docker run command generation
```

### Core Components

#### 1. **ContainerSpec** (`pkg/containerconfig/spec.go`)

Defines the container configuration structure:

```go
type ContainerSpec struct {
    Name       string
    Image      string
    Env        []string
    Volumes    []string
    Ports      []string
    Networks   []string
    Command    []string
    WorkingDir string
    Labels     map[string]string
    EntryPoint []string
    Devices    []string
    ExtraHosts []string
    Restart    string
}
```

#### 2. **Parser** (`pkg/containerconfig/parser.go`)

Parses `docker inspect` JSON output into `ContainerSpec`:

```go
spec, err := containerconfig.ParseInspectJSON(jsonData)
```

#### 3. **Generator** (`pkg/containerconfig/generator.go`)

Generates `docker run` commands from `ContainerSpec`:

```go
runCmd := containerconfig.GenerateRunCommand(spec, opts)
```

#### 4. **Manager** (`main.go`)

Orchestrates container operations with clean, single-responsibility methods:

- `GetContainerConfig()` - Retrieves container configuration
- `CreateDevContainer()` - Creates development container
- `StopDevContainer()` - Stops container
- `RemoveDevContainer()` - Removes container
- `CheckDevContainerExists()` - Checks container existence

## 💡 Usage as a Library

You can use the `containerconfig` package in your own Go projects:

```go
package main

import (
    "fmt"
    "github.com/yourusername/docker-config-extractor/pkg/containerconfig"
)

func main() {
    // Parse container configuration
    spec, err := containerconfig.ParseInspectJSON(jsonData)
    if err != nil {
        panic(err)
    }

    // Generate docker run command
    opts := &containerconfig.RunOptions{
        Name: "my-dev-container",
    }
    runCmd := containerconfig.GenerateRunCommand(spec, opts)
    fmt.Println(runCmd)
}
```

## 🔧 Advanced Features

### Debugger Integration

When creating a dev container, the tool automatically:

1. Checks if Go is installed in the container
2. Installs Delve debugger (`dlv`)
3. Exposes port 2345 for remote debugging
4. Verifies successful installation

```bash
# After container creation, you can attach the debugger:
docker exec -it myapp-dev dlv attach <pid>
```

### Custom Script Injection

The tool supports injecting custom initialization scripts into the container after creation. This is useful for:

- Installing additional development tools
- Setting up environment configurations
- Running initialization scripts

### Container Lifecycle Management

```bash
# Check if container exists
# If exists, prompts for recreation

# Stop and remove existing container
# Create new dev container with updated configuration
```

## 🛠️ Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -v
```

### Code Structure Philosophy

This project follows clean architecture principles:

- **Separation of Concerns**: Parsing, generation, and execution are separate
- **No Shell Command Chaining**: Avoids `&&` operators; uses separate `docker exec` calls
- **Single Responsibility**: Each method does one thing well
- **Comprehensive Error Handling**: All errors include context
- **Extensive Logging**: Clear visibility into operations

## 📋 Requirements

- **Go 1.25+**
- **Docker Engine** (tested with Docker 20.10+)
- **Permissions**: User must have Docker access

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Guidelines

- Follow Go conventions and best practices
- Add tests for new features
- Update documentation for API changes
- Ensure `go build` passes before submitting

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
