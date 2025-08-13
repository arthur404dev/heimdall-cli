package types

// ConfigPaths holds the paths for configuration files
type ConfigPaths struct {
	// Base directory for all config files
	BaseDir string `json:"base_dir"`
	// Pattern for config file names (e.g., "%s.json" where %s is the domain)
	FilePattern string `json:"file_pattern"`
	// Schema directory for cached schemas
	SchemaDir string `json:"schema_dir"`
	// Backup directory for config backups
	BackupDir string `json:"backup_dir"`
	// Output paths for generated configs (optional, provider-specific)
	OutputPaths map[string]string `json:"output_paths,omitempty"`
}
