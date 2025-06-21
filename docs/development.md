# 🔧 Development Guide

## Getting Started

### Prerequisites

- Go 1.21+
- OpenTofu
- LocalStack (for local testing)
- Git

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/rioprayogo/bolt.git
cd bolt

# Install dependencies
go mod tidy

# Build project
go build -o bolt .

# Run tests
go test ./...
```

## Project Structure

```
bold/
├── cmd/                    # CLI commands
│   ├── analyze.go         # Analysis command
│   ├── bootstrap.go       # Bootstrap command
│   ├── destroy.go         # Destroy command
│   └── main.go           # Main entry point
├── pkg/                   # Core packages
│   ├── compiler/         # OpenTofu code generation
│   ├── config/           # Configuration management
│   ├── cost/             # Cost estimation
│   ├── engine/           # Deployment engine
│   ├── errors/           # Error handling
│   ├── graph/            # Dependency graph
│   ├── logger/           # Logging
│   ├── parser/           # YAML parsing
│   └── workflow/         # Workflow management
├── docs/                  # Documentation
├── examples/              # Example configurations
├── tests/                 # Integration tests
└── README.md
```

## Architecture Overview

### Core Components

1. **Parser** (`pkg/parser/`)
   - YAML configuration parsing
   - Schema validation
   - Input sanitization

2. **Compiler** (`pkg/compiler/`)
   - OpenTofu code generation
   - Provider-specific resource mapping
   - Configuration optimization

3. **Engine** (`pkg/engine/`)
   - Orchestration layer
   - State management
   - Provider coordination

4. **Workflow** (`pkg/workflow/`)
   - Deployment strategies
   - Rollback mechanisms
   - Progress tracking

## Development Workflow

### 1. Feature Development

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes
# ... edit files ...

# Run tests
go test ./...

# Build and test
go build -o bolt .
./bolt analyze service.yaml

# Commit changes
git add .
git commit -m "feat: add new feature"

# Push branch
git push origin feature/new-feature
```

### 2. Testing

#### Unit Tests
```bash
# Run all tests
go test ./...

# Run specific package
go test ./pkg/parser

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...
```

#### Integration Tests
```bash
# Test with LocalStack
docker run -d --name localstack -p 4566:4566 localstack/localstack

# Run integration tests
go test -tags=integration ./tests/

# Clean up
docker stop localstack && docker rm localstack
```

#### Manual Testing
```bash
# Build and test locally
go build -o bolt .
./bolt analyze service.yaml
./bolt bootstrap service.yaml
./bolt destroy service.yaml
```

### 3. Code Quality

#### Linting
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

#### Formatting
```bash
# Format code
go fmt ./...

# Check formatting
go vet ./...
```

## Adding New Features

### 1. New Resource Type

```go
// pkg/compiler/resources.go
type NewResource struct {
    Name     string `yaml:"name"`
    Provider string `yaml:"provider"`
    Spec     struct {
        // Resource-specific fields
    } `yaml:"spec"`
}

// pkg/compiler/compiler.go
func (c *Compiler) generateNewResource(resource *NewResource) string {
    // Generate OpenTofu code
    return `
resource "aws_new_resource" "` + resource.Name + `" {
  // Resource configuration
}
`
}
```

### 2. New Provider

```go
// pkg/parser/provider.go
type NewProvider struct {
    Name string `yaml:"name"`
    Type string `yaml:"type"`
    Spec struct {
        Region string `yaml:"region"`
        // Provider-specific fields
    } `yaml:"spec"`
}

// pkg/compiler/compiler.go
func (c *Compiler) generateNewProvider(provider *NewProvider) string {
    return `
provider "new_provider" {
  region = "` + provider.Spec.Region + `"
}
`
}
```

### 3. New Command

```go
// cmd/newcommand.go
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: bolt newcommand <service.yaml>")
        os.Exit(1)
    }

    serviceFile := os.Args[1]
    
    // Command implementation
    fmt.Printf("Running newcommand on %s\n", serviceFile)
}
```

## Testing Guidelines

### Unit Tests
- Test individual functions and methods
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Aim for >80% code coverage

### Integration Tests
- Test complete workflows
- Use real cloud providers (dev/test accounts)
- Test error scenarios
- Validate output formats

### Performance Tests
- Test with large configurations
- Measure memory usage
- Benchmark critical paths
- Test concurrent operations

## Error Handling

### Error Types
```go
// pkg/errors/errors.go
type ValidationError struct {
    Field   string
    Message string
}

type CompilationError struct {
    Resource string
    Message  string
}

type DeploymentError struct {
    Stage    string
    Message  string
    Rollback bool
}
```

### Error Handling Patterns
```go
// Always check errors
if err != nil {
    return fmt.Errorf("failed to parse config: %w", err)
}

// Use custom error types
if !isValid {
    return &ValidationError{
        Field:   "name",
        Message: "name is required",
    }
}
```

## Logging

### Log Levels
```go
// pkg/logger/logger.go
const (
    DEBUG = "debug"
    INFO  = "info"
    WARN  = "warn"
    ERROR = "error"
)

// Usage
logger.Info("Starting deployment", "service", serviceName)
logger.Debug("Generated OpenTofu code", "code", code)
logger.Error("Deployment failed", "error", err)
```

### Structured Logging
```go
logger.Info("Resource created",
    "resource", resource.Name,
    "provider", resource.Provider,
    "type", resource.Type,
)
```

## Configuration

### Environment Variables
```bash
# Development
export BOLT_LOG_LEVEL=debug
export BOLT_CONFIG_PATH=./config.yaml

# Production
export BOLT_LOG_LEVEL=info
export BOLT_CONFIG_PATH=/etc/bolt/config.yaml
```

### Configuration File
```yaml
# config.yaml
log_level: info
config_path: ./config
providers:
  aws:
    default_region: us-east-1
  azure:
    default_location: eastus
  gcp:
    default_region: us-central1
```

## Release Process

### 1. Version Bumping
```bash
# Update version in main.go
const Version = "1.1.0"

# Update go.mod if needed
go mod tidy
```

### 2. Testing
```bash
# Run all tests
go test ./...

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o bolt-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o bolt-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o bolt-windows-amd64.exe .
```

### 3. Release
```bash
# Create tag
git tag v1.1.0

# Push tag
git push origin v1.1.0

# Create GitHub release
# Upload binaries
```

## Contributing Guidelines

### Code Style
- Follow Go conventions
- Use meaningful variable names
- Add comments for complex logic
- Keep functions small and focused

### Commit Messages
```
feat: add new resource type
fix: resolve parsing error
docs: update README
test: add integration tests
refactor: improve error handling
```

### Pull Request Process
1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Update documentation
6. Submit pull request

### Review Process
- All PRs require review
- Tests must pass
- Code coverage should not decrease
- Documentation must be updated

## Troubleshooting

### Common Issues

#### Build Errors
```bash
# Clean and rebuild
go clean
go mod tidy
go build -o bolt .
```

#### Test Failures
```bash
# Run tests with verbose output
go test -v ./...

# Check for race conditions
go test -race ./...
```

#### LocalStack Issues
```bash
# Restart LocalStack
docker restart localstack

# Check LocalStack logs
docker logs localstack
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [OpenTofu Documentation](https://opentofu.org/docs)
- [LocalStack Documentation](https://docs.localstack.cloud/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)

## Support

For development questions:
- Open an issue on GitHub
- Join our discussions
- Check existing documentation 