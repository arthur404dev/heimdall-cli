package theme

import (
	"fmt"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
)

// String returns the string representation of the severity
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ThemeApplicationError represents an error during theme application
type ThemeApplicationError struct {
	Application string
	Operation   string
	Err         error
	Severity    ErrorSeverity
	Recoverable bool
	Suggestion  string
	Context     map[string]interface{}
}

// Error implements the error interface
func (e ThemeApplicationError) Error() string {
	return fmt.Sprintf("%s: %s failed: %v", e.Application, e.Operation, e.Err)
}

// UserMessage returns a user-friendly error message
func (e ThemeApplicationError) UserMessage() string {
	var msg strings.Builder

	// Build the main error message
	msg.WriteString(fmt.Sprintf("[%s] %s", e.Severity, e.Error()))

	// Add suggestion if available
	if e.Suggestion != "" {
		msg.WriteString(fmt.Sprintf("\n💡 Suggestion: %s", e.Suggestion))
	}

	// Add context if available
	if len(e.Context) > 0 {
		msg.WriteString("\n📋 Context:")
		for key, value := range e.Context {
			msg.WriteString(fmt.Sprintf("\n  - %s: %v", key, value))
		}
	}

	// Add recovery information
	if e.Recoverable {
		msg.WriteString("\n✅ This error is recoverable and the operation can continue.")
	} else if e.Severity == SeverityFatal {
		msg.WriteString("\n❌ This is a fatal error. The operation cannot continue.")
	}

	return msg.String()
}

// ErrorCollector collects errors during batch operations
type ErrorCollector struct {
	errors []ThemeApplicationError
	fatal  bool
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]ThemeApplicationError, 0),
		fatal:  false,
	}
}

// Add adds an error to the collector
func (ec *ErrorCollector) Add(err ThemeApplicationError) {
	ec.errors = append(ec.errors, err)
	if err.Severity == SeverityFatal {
		ec.fatal = true
	}
}

// AddError adds a simple error to the collector
func (ec *ErrorCollector) AddError(app, operation string, err error) {
	ec.Add(ThemeApplicationError{
		Application: app,
		Operation:   operation,
		Err:         err,
		Severity:    SeverityError,
		Recoverable: false,
	})
}

// AddWarning adds a warning to the collector
func (ec *ErrorCollector) AddWarning(app, operation string, err error) {
	ec.Add(ThemeApplicationError{
		Application: app,
		Operation:   operation,
		Err:         err,
		Severity:    SeverityWarning,
		Recoverable: true,
	})
}

// HasErrors returns true if there are any errors
func (ec *ErrorCollector) HasErrors() bool {
	for _, err := range ec.errors {
		if err.Severity >= SeverityError {
			return true
		}
	}
	return false
}

// HasFatal returns true if there are any fatal errors
func (ec *ErrorCollector) HasFatal() bool {
	return ec.fatal
}

// GetErrors returns all collected errors
func (ec *ErrorCollector) GetErrors() []ThemeApplicationError {
	return ec.errors
}

// GetErrorsBySeverity returns errors of a specific severity
func (ec *ErrorCollector) GetErrorsBySeverity(severity ErrorSeverity) []ThemeApplicationError {
	var filtered []ThemeApplicationError
	for _, err := range ec.errors {
		if err.Severity == severity {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// GetErrorsByApplication returns errors for a specific application
func (ec *ErrorCollector) GetErrorsByApplication(app string) []ThemeApplicationError {
	var filtered []ThemeApplicationError
	for _, err := range ec.errors {
		if err.Application == app {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// Summary returns a summary of all errors
func (ec *ErrorCollector) Summary() string {
	if len(ec.errors) == 0 {
		return "No errors occurred"
	}

	var summary strings.Builder

	// Count errors by severity
	counts := make(map[ErrorSeverity]int)
	for _, err := range ec.errors {
		counts[err.Severity]++
	}

	summary.WriteString("Error Summary:\n")
	for severity := SeverityFatal; severity >= SeverityInfo; severity-- {
		if count := counts[severity]; count > 0 {
			summary.WriteString(fmt.Sprintf("  %s: %d\n", severity, count))
		}
	}

	// List errors by application
	appErrors := make(map[string][]ThemeApplicationError)
	for _, err := range ec.errors {
		appErrors[err.Application] = append(appErrors[err.Application], err)
	}

	if len(appErrors) > 0 {
		summary.WriteString("\nErrors by Application:\n")
		for app, errors := range appErrors {
			summary.WriteString(fmt.Sprintf("  %s: %d error(s)\n", app, len(errors)))
		}
	}

	return summary.String()
}

// Clear clears all collected errors
func (ec *ErrorCollector) Clear() {
	ec.errors = []ThemeApplicationError{}
	ec.fatal = false
}

// HandleThemeError processes and logs a theme error appropriately
func HandleThemeError(err error) error {
	if err == nil {
		return nil
	}

	// Check if it's a ThemeApplicationError
	if themeErr, ok := err.(ThemeApplicationError); ok {
		// Log based on severity
		switch themeErr.Severity {
		case SeverityFatal:
			logger.Error("Fatal error",
				"application", themeErr.Application,
				"operation", themeErr.Operation,
				"error", themeErr.Err,
				"suggestion", themeErr.Suggestion)
			return fmt.Errorf("fatal error: %w", err)

		case SeverityError:
			if themeErr.Recoverable {
				logger.Error("Recoverable error",
					"application", themeErr.Application,
					"operation", themeErr.Operation,
					"error", themeErr.Err,
					"suggestion", themeErr.Suggestion)
				// Continue execution for recoverable errors
				return nil
			}
			logger.Error("Error",
				"application", themeErr.Application,
				"operation", themeErr.Operation,
				"error", themeErr.Err)
			return err

		case SeverityWarning:
			logger.Warn("Warning",
				"application", themeErr.Application,
				"operation", themeErr.Operation,
				"error", themeErr.Err)
			return nil // Warnings don't stop execution

		case SeverityInfo:
			logger.Info("Info",
				"application", themeErr.Application,
				"message", themeErr.Err.Error())
			return nil
		}
	}

	// For regular errors, just return them
	return err
}

// Common error creation helpers

// NewFileWriteError creates a file write error
func NewFileWriteError(app, path string, err error) ThemeApplicationError {
	return ThemeApplicationError{
		Application: app,
		Operation:   "write file",
		Err:         err,
		Severity:    SeverityError,
		Recoverable: false,
		Suggestion:  "Check file permissions and disk space",
		Context: map[string]interface{}{
			"path": path,
		},
	}
}

// NewTemplateError creates a template processing error
func NewTemplateError(app, template string, err error) ThemeApplicationError {
	return ThemeApplicationError{
		Application: app,
		Operation:   "process template",
		Err:         err,
		Severity:    SeverityError,
		Recoverable: false,
		Suggestion:  "Check template syntax and variables",
		Context: map[string]interface{}{
			"template": template,
		},
	}
}

// NewPermissionError creates a permission error
func NewPermissionError(app, path string) ThemeApplicationError {
	return ThemeApplicationError{
		Application: app,
		Operation:   "access file",
		Err:         fmt.Errorf("permission denied"),
		Severity:    SeverityError,
		Recoverable: false,
		Suggestion:  fmt.Sprintf("Ensure you have write permissions for %s", path),
		Context: map[string]interface{}{
			"path": path,
		},
	}
}

// NewMissingDependencyError creates a missing dependency error
func NewMissingDependencyError(app, dependency string) ThemeApplicationError {
	return ThemeApplicationError{
		Application: app,
		Operation:   "check dependencies",
		Err:         fmt.Errorf("%s not found", dependency),
		Severity:    SeverityWarning,
		Recoverable: true,
		Suggestion:  fmt.Sprintf("Install %s to enable theming for %s", dependency, app),
		Context: map[string]interface{}{
			"dependency": dependency,
		},
	}
}

// NewColorFormatError creates a color format error
func NewColorFormatError(app, color string) ThemeApplicationError {
	return ThemeApplicationError{
		Application: app,
		Operation:   "validate color",
		Err:         fmt.Errorf("invalid color format: %s", color),
		Severity:    SeverityError,
		Recoverable: false,
		Suggestion:  "Colors should be in hex format (#RRGGBB)",
		Context: map[string]interface{}{
			"color": color,
		},
	}
}

// NewBackupError creates a backup error
func NewBackupError(operation string, err error) ThemeApplicationError {
	return ThemeApplicationError{
		Application: "backup",
		Operation:   operation,
		Err:         err,
		Severity:    SeverityWarning,
		Recoverable: true,
		Suggestion:  "Theme application will continue without backup",
	}
}

// NewRollbackError creates a rollback error
func NewRollbackError(operation string, err error) ThemeApplicationError {
	return ThemeApplicationError{
		Application: "rollback",
		Operation:   operation,
		Err:         err,
		Severity:    SeverityError,
		Recoverable: false,
		Suggestion:  "Manual intervention may be required to restore original state",
	}
}
