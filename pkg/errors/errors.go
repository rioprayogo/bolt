package errors

import (
	"fmt"
)

// CompilationError represents errors during compilation
type CompilationError struct {
	Message string
	Details string
}

func (e CompilationError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("compilation error: %s (details: %s)", e.Message, e.Details)
	}
	return fmt.Sprintf("compilation error: %s", e.Message)
}

// ExecutionError represents errors during OpenTofu execution
type ExecutionError struct {
	Command  string
	Output   string
	ExitCode int
}

func (e ExecutionError) Error() string {
	return fmt.Sprintf("execution error in command '%s' (exit code: %d): %s", e.Command, e.ExitCode, e.Output)
}

// ConfigurationError represents configuration-related errors
type ConfigurationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error in field '%s' with value '%v': %s", e.Field, e.Value, e.Message)
}

// ResourceError represents resource-specific errors
type ResourceError struct {
	ResourceType string
	ResourceName string
	Message      string
}

func (e ResourceError) Error() string {
	return fmt.Sprintf("resource error in %s '%s': %s", e.ResourceType, e.ResourceName, e.Message)
}

// ProviderError represents provider-specific errors
type ProviderError struct {
	Provider string
	Message  string
}

func (e ProviderError) Error() string {
	return fmt.Sprintf("provider error for '%s': %s", e.Provider, e.Message)
}

// DependencyError represents dependency-related errors
type DependencyError struct {
	Dependent  string
	Dependency string
	Message    string
}

func (e DependencyError) Error() string {
	return fmt.Sprintf("dependency error: '%s' depends on '%s': %s", e.Dependent, e.Dependency, e.Message)
}
