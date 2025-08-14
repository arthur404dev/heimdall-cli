package theme

import (
	"context"
)

// ThemeEngine is the main interface for the theme application system
type ThemeEngine interface {
	// ApplyTheme applies a theme to specified applications
	ApplyTheme(ctx context.Context, scheme *ColorScheme, options ApplyOptions) error

	// ValidateTheme validates a theme before application
	ValidateTheme(scheme *ColorScheme) error

	// GetSupportedApplications returns a list of supported applications
	GetSupportedApplications() []string

	// RegisterHandler registers a new application handler
	RegisterHandler(name string, handler ApplicationHandler) error
}

// ApplicationHandler handles theme application for a specific application
type ApplicationHandler interface {
	// Name returns the name of the application
	Name() string

	// Apply applies the theme to the application
	Apply(colors map[string]string, options HandlerOptions) error

	// Validate validates that the handler can apply the theme
	Validate(colors map[string]string) error

	// GetOutputPath returns the output path for the themed configuration
	GetOutputPath() string

	// IsInstalled checks if the application is installed
	IsInstalled() bool

	// RequiredColors returns the list of required color keys
	RequiredColors() []string
}

// TemplateProcessor processes templates with color replacements
type TemplateProcessor interface {
	// ProcessSimple performs simple {{variable}} replacements
	ProcessSimple(template string, colors map[string]string) (string, error)

	// ProcessAdvanced performs advanced template processing with conditionals
	ProcessAdvanced(name, template string, data TemplateData) (string, error)

	// ValidateTemplate validates template syntax
	ValidateTemplate(template string) error
}

// ColorMapper maps color schemes to application-specific formats
type ColorMapper interface {
	// MapColors maps a color scheme to application-specific color names
	MapColors(scheme *ColorScheme, targetApp string) (map[string]string, error)

	// ConvertFormat converts a color to a specific format
	ConvertFormat(color string, format ColorFormat) (string, error)

	// ValidateColor validates a color string
	ValidateColor(color string) error
}

// BackupManager manages backups of configuration files
type BackupManager interface {
	// CreateBackup creates a backup of specified files
	CreateBackup(files []string) (backupID string, err error)

	// RestoreBackup restores files from a backup
	RestoreBackup(backupID string) error

	// ListBackups lists available backups
	ListBackups() ([]BackupInfo, error)

	// CleanOldBackups removes old backups based on retention policy
	CleanOldBackups() error
}

// TransactionManager manages atomic theme application transactions
type TransactionManager interface {
	// Begin starts a new transaction
	Begin() (Transaction, error)

	// Commit commits all changes in the transaction
	Commit(tx Transaction) error

	// Rollback rolls back all changes in the transaction
	Rollback(tx Transaction) error
}

// Transaction represents an atomic theme application transaction
type Transaction interface {
	// AddOperation adds an operation to the transaction
	AddOperation(op Operation) error

	// GetOperations returns all operations in the transaction
	GetOperations() []Operation

	// GetID returns the transaction ID
	GetID() string
}

// Operation represents a single operation in a transaction
type Operation interface {
	// Execute performs the operation
	Execute() error

	// Rollback reverses the operation
	Rollback() error

	// Description returns a human-readable description
	Description() string

	// GetType returns the operation type
	GetType() OperationType
}

// ColorScheme represents a complete color scheme
type ColorScheme struct {
	Name    string            `json:"name"`
	Flavour string            `json:"flavour,omitempty"`
	Mode    string            `json:"mode"` // dark or light
	Variant string            `json:"variant,omitempty"`
	Colors  map[string]string `json:"colors"`
	Special map[string]string `json:"special,omitempty"`
}

// ApplyOptions contains options for theme application
type ApplyOptions struct {
	// Applications to apply theme to (empty means all)
	Applications []string

	// DryRun performs a dry run without making changes
	DryRun bool

	// Force forces regeneration even if cached
	Force bool

	// Parallel enables parallel application
	Parallel bool

	// TemplateDir specifies a custom template directory
	TemplateDir string

	// Verbose enables verbose output
	Verbose bool

	// NoBackup disables backup creation
	NoBackup bool
}

// HandlerOptions contains options for individual handlers
type HandlerOptions struct {
	// TemplateOverride allows overriding the default template
	TemplateOverride string

	// OutputPath allows overriding the default output path
	OutputPath string

	// Mode specifies dark or light mode
	Mode string

	// Verbose enables verbose output
	Verbose bool
}

// TemplateData contains data for template processing
type TemplateData struct {
	Colors map[string]string      `json:"colors"`
	Mode   string                 `json:"mode"`
	Dark   bool                   `json:"dark"`
	Light  bool                   `json:"light"`
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// ColorFormat represents different color format types
type ColorFormat string

const (
	ColorFormatHex       ColorFormat = "hex"         // #RRGGBB
	ColorFormatHexNoHash ColorFormat = "hex_no_hash" // RRGGBB (for QuickShell)
	ColorFormatRGB       ColorFormat = "rgb"         // rgb(r, g, b)
	ColorFormatRGBA      ColorFormat = "rgba"        // rgba(r, g, b, a)
	ColorFormatHSL       ColorFormat = "hsl"         // hsl(h, s%, l%)
	ColorFormatHSLA      ColorFormat = "hsla"        // hsla(h, s%, l%, a)
)

// OperationType represents the type of operation
type OperationType string

const (
	OperationTypeWrite   OperationType = "write"
	OperationTypeDelete  OperationType = "delete"
	OperationTypeExecute OperationType = "execute"
	OperationTypeSymlink OperationType = "symlink"
	OperationTypeCommand OperationType = "command"
	OperationTypeBatch   OperationType = "batch"
)

// BackupInfo contains information about a backup
type BackupInfo struct {
	ID        string   `json:"id"`
	Timestamp int64    `json:"timestamp"`
	Files     []string `json:"files"`
	Size      int64    `json:"size"`
}

// ThemeError represents a theme-related error with context
type ThemeError struct {
	Application string
	Operation   string
	Err         error
	Severity    ErrorSeverity
	Recoverable bool
	Suggestion  string
}

// ErrorSeverity represents the severity of an error
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityFatal
)

// Error implements the error interface
func (e ThemeError) Error() string {
	return e.Err.Error()
}

// UserMessage returns a user-friendly error message
func (e ThemeError) UserMessage() string {
	msg := e.Error()
	if e.Suggestion != "" {
		msg += "\nSuggestion: " + e.Suggestion
	}
	return msg
}
