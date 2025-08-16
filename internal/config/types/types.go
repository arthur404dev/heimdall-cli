package types

// ConfigPaths holds the paths for configuration files
type ConfigPaths struct {
	// Base directory for all config files
	BaseDir string `json:"base_dir" desc:"Base directory for all configuration files" example:"~/.config/heimdall"`
	// Pattern for config file names (e.g., "%s.json" where %s is the domain)
	FilePattern string `json:"file_pattern" desc:"Pattern for config file names (%s is replaced with domain)" default:"%s.json" example:"%s-config.json"`
	// Schema directory for cached schemas
	SchemaDir string `json:"schema_dir" desc:"Directory for cached configuration schemas" example:"~/.cache/heimdall/schemas"`
	// Backup directory for config backups
	BackupDir string `json:"backup_dir" desc:"Directory for configuration backups" example:"~/.local/share/heimdall/backups"`
	// Output paths for generated configs (optional, provider-specific)
	OutputPaths map[string]string `json:"output_paths,omitempty" desc:"Provider-specific output paths for generated configs" example:"{\"shell\": \"~/.config/quickshell/heimdall.json\"}"`
}
