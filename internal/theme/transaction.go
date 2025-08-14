package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// TransactionOperation implements the Operation interface for transactions
type TransactionOperation struct {
	execute     func() error
	rollback    func() error
	description string
	opType      OperationType
}

// Execute performs the operation
func (to *TransactionOperation) Execute() error {
	if to.execute != nil {
		return to.execute()
	}
	return nil
}

// Rollback undoes the operation
func (to *TransactionOperation) Rollback() error {
	if to.rollback != nil {
		return to.rollback()
	}
	return nil
}

// Description returns a human-readable description
func (to *TransactionOperation) Description() string {
	return to.description
}

// GetType returns the operation type
func (to *TransactionOperation) GetType() OperationType {
	return to.opType
}

// ThemeTransaction manages atomic theme application transactions
type ThemeTransaction struct {
	operations []Operation
	backup     *FileBackupManager
	backupID   string
	mu         sync.Mutex
	executed   []int // Track which operations were executed
}

// NewThemeTransaction creates a new transaction
func NewThemeTransaction(backup *FileBackupManager) *ThemeTransaction {
	return &ThemeTransaction{
		operations: make([]Operation, 0),
		backup:     backup,
		executed:   make([]int, 0),
	}
}

// AddOperation adds an operation to the transaction
func (tt *ThemeTransaction) AddOperation(op Operation) {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	tt.operations = append(tt.operations, op)
}

// Execute runs all operations in the transaction
func (tt *ThemeTransaction) Execute() error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	// Create backup of all affected files
	files := tt.getAffectedFiles()
	if len(files) > 0 && tt.backup != nil {
		backupID, err := tt.backup.Backup(files)
		if err != nil {
			logger.Warn("Failed to create backup, proceeding without backup", "error", err)
			// Continue without backup - not fatal
		} else {
			tt.backupID = backupID
			logger.Info("Created backup", "id", backupID)
		}
	}

	// Execute operations
	for i, op := range tt.operations {
		logger.Debug("Executing operation", "index", i, "description", op.Description())

		if err := op.Execute(); err != nil {
			logger.Error("Operation failed",
				"index", i,
				"operation", op.Description(),
				"error", err)

			// Rollback on failure
			if rollbackErr := tt.rollback(i); rollbackErr != nil {
				logger.Error("Rollback failed", "error", rollbackErr)
			}

			return fmt.Errorf("operation %d (%s) failed: %w", i, op.Description(), err)
		}

		// Track successful execution
		tt.executed = append(tt.executed, i)
	}

	logger.Info("Transaction completed successfully", "operations", len(tt.operations))
	return nil
}

// rollback undoes executed operations
func (tt *ThemeTransaction) rollback(failedIndex int) error {
	logger.Info("Rolling back transaction", "failed_at", failedIndex, "executed", len(tt.executed))

	var rollbackErrors []error

	// Rollback executed operations in reverse order
	for i := len(tt.executed) - 1; i >= 0; i-- {
		opIndex := tt.executed[i]
		op := tt.operations[opIndex]

		logger.Debug("Rolling back operation", "index", opIndex, "description", op.Description())

		if err := op.Rollback(); err != nil {
			logger.Error("Rollback failed for operation",
				"index", opIndex,
				"operation", op.Description(),
				"error", err)
			rollbackErrors = append(rollbackErrors, err)
		}
	}

	// Restore from backup as final fallback
	if tt.backupID != "" && tt.backup != nil {
		logger.Info("Restoring from backup", "backup", tt.backupID)
		if err := tt.backup.Restore(tt.backupID); err != nil {
			logger.Error("Backup restore failed", "backup", tt.backupID, "error", err)
			rollbackErrors = append(rollbackErrors, err)
		} else {
			logger.Info("Successfully restored from backup", "backup", tt.backupID)
		}
	}

	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback completed with %d errors", len(rollbackErrors))
	}

	return nil
}

// getAffectedFiles returns all files that will be modified
func (tt *ThemeTransaction) getAffectedFiles() []string {
	fileMap := make(map[string]bool)

	for _, op := range tt.operations {
		if fileOp, ok := op.(*FileOperation); ok {
			fileMap[fileOp.Path] = true
		}
	}

	files := make([]string, 0, len(fileMap))
	for file := range fileMap {
		files = append(files, file)
	}

	return files
}

// GetOperationCount returns the number of operations in the transaction
func (tt *ThemeTransaction) GetOperationCount() int {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	return len(tt.operations)
}

// GetExecutedCount returns the number of successfully executed operations
func (tt *ThemeTransaction) GetExecutedCount() int {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	return len(tt.executed)
}

// FileOperation represents a file write operation
type FileOperation struct {
	Path       string
	Content    []byte
	OldContent []byte
	Mode       os.FileMode
}

// NewFileOperation creates a new file operation
func NewFileOperation(path string, content []byte) *FileOperation {
	return &FileOperation{
		Path:    path,
		Content: content,
		Mode:    0644, // Default mode
	}
}

// SetMode sets the file mode for the operation
func (fo *FileOperation) SetMode(mode os.FileMode) {
	fo.Mode = mode
}

// Execute performs the file write
func (fo *FileOperation) Execute() error {
	// Save old content for rollback if file exists
	if info, err := os.Stat(fo.Path); err == nil {
		data, err := os.ReadFile(fo.Path)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
		fo.OldContent = data
		fo.Mode = info.Mode() // Preserve existing mode
	}

	// Ensure directory exists
	dir := filepath.Dir(fo.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Use atomic write
	return paths.AtomicWrite(fo.Path, fo.Content)
}

// Rollback restores the original file state
func (fo *FileOperation) Rollback() error {
	if fo.OldContent != nil {
		// Restore original content
		return paths.AtomicWrite(fo.Path, fo.OldContent)
	}

	// If no old content, remove the file
	if err := os.Remove(fo.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

// Description returns a description of the operation
func (fo *FileOperation) Description() string {
	return fmt.Sprintf("Write file %s", fo.Path)
}

// GetType returns the operation type
func (fo *FileOperation) GetType() OperationType {
	return OperationTypeWrite
}

// TemplateOperation represents a template processing operation
type TemplateOperation struct {
	*FileOperation
	TemplateName string
	Variables    map[string]string
}

// NewTemplateOperation creates a new template operation
func NewTemplateOperation(path, templateName string, variables map[string]string) *TemplateOperation {
	return &TemplateOperation{
		FileOperation: NewFileOperation(path, nil),
		TemplateName:  templateName,
		Variables:     variables,
	}
}

// Execute processes the template and writes the result
func (to *TemplateOperation) Execute() error {
	// Process template (this would use the actual template processor)
	// For now, this is a placeholder
	content := []byte(fmt.Sprintf("# Generated from template: %s\n", to.TemplateName))
	to.FileOperation.Content = content

	return to.FileOperation.Execute()
}

// Description returns a description of the operation
func (to *TemplateOperation) Description() string {
	return fmt.Sprintf("Apply template %s to %s", to.TemplateName, to.Path)
}

// CommandOperation represents a command execution operation
type CommandOperation struct {
	Command string
	Args    []string
	Desc    string // Renamed to avoid conflict with Description() method
	Undo    func() error
}

// NewCommandOperation creates a new command operation
func NewCommandOperation(command string, args []string, description string) *CommandOperation {
	return &CommandOperation{
		Command: command,
		Args:    args,
		Desc:    description,
	}
}

// SetUndo sets the undo function for the command
func (co *CommandOperation) SetUndo(undo func() error) {
	co.Undo = undo
}

// Execute runs the command
func (co *CommandOperation) Execute() error {
	// This would execute the actual command
	// For now, this is a placeholder
	logger.Debug("Would execute command", "command", co.Command, "args", co.Args)
	return nil
}

// Rollback undoes the command if an undo function is provided
func (co *CommandOperation) Rollback() error {
	if co.Undo != nil {
		return co.Undo()
	}
	// No undo available
	logger.Warn("No undo available for command", "command", co.Command)
	return nil
}

// Description returns the operation description
func (co *CommandOperation) Description() string {
	if co.Desc != "" {
		return co.Desc
	}
	return fmt.Sprintf("Execute %s", co.Command)
}

// GetType returns the operation type
func (co *CommandOperation) GetType() OperationType {
	return OperationTypeCommand
}

// BatchOperation groups multiple operations
type BatchOperation struct {
	Operations []Operation
	Desc       string // Renamed to avoid conflict
	executed   []int
	mu         sync.Mutex
}

// NewBatchOperation creates a new batch operation
func NewBatchOperation(description string) *BatchOperation {
	return &BatchOperation{
		Operations: make([]Operation, 0),
		Desc:       description,
		executed:   make([]int, 0),
	}
}

// AddOperation adds an operation to the batch
func (bo *BatchOperation) AddOperation(op Operation) {
	bo.mu.Lock()
	defer bo.mu.Unlock()
	bo.Operations = append(bo.Operations, op)
}

// Execute runs all operations in the batch
func (bo *BatchOperation) Execute() error {
	bo.mu.Lock()
	defer bo.mu.Unlock()

	for i, op := range bo.Operations {
		if err := op.Execute(); err != nil {
			// Rollback executed operations
			bo.rollbackInternal()
			return fmt.Errorf("batch operation %d failed: %w", i, err)
		}
		bo.executed = append(bo.executed, i)
	}

	return nil
}

// Rollback undoes all executed operations in the batch
func (bo *BatchOperation) Rollback() error {
	bo.mu.Lock()
	defer bo.mu.Unlock()
	return bo.rollbackInternal()
}

// rollbackInternal performs the actual rollback
func (bo *BatchOperation) rollbackInternal() error {
	var errors []error

	// Rollback in reverse order
	for i := len(bo.executed) - 1; i >= 0; i-- {
		opIndex := bo.executed[i]
		if err := bo.Operations[opIndex].Rollback(); err != nil {
			errors = append(errors, err)
		}
	}

	bo.executed = []int{} // Clear executed list

	if len(errors) > 0 {
		return fmt.Errorf("batch rollback had %d errors", len(errors))
	}

	return nil
}

// Description returns the batch description
func (bo *BatchOperation) Description() string {
	if bo.Desc != "" {
		return bo.Desc
	}
	return fmt.Sprintf("Batch of %d operations", len(bo.Operations))
}

// GetType returns the operation type
func (bo *BatchOperation) GetType() OperationType {
	return OperationTypeBatch
}
