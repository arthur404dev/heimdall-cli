package clipboard

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/spf13/cobra"
)

var (
	deleteFlag bool
)

// NewCommand creates the clipboard command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clipboard",
		Short: "Manage clipboard history",
		Long:  `Display and manage clipboard history using cliphist and fuzzel`,
		RunE:  run,
	}

	cmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete selected item from clipboard history")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	external := cfg.External
	clipboardCfg := cfg.Clipboard

	// Get clipboard history from cliphist
	cliphistPath := external.Cliphist
	if cliphistPath == "" {
		cliphistPath = "cliphist"
	}

	cliphist := exec.Command(cliphistPath, "list")
	clipOutput, err := cliphist.Output()
	if err != nil {
		return fmt.Errorf("failed to get clipboard history: %w", err)
	}

	// Prepare fuzzel arguments from config
	fuzzelPath := external.Fuzzel
	if fuzzelPath == "" {
		fuzzelPath = "fuzzel"
	}

	var fuzzelArgs []string
	// Add configured fuzzel args
	fuzzelArgs = append(fuzzelArgs, clipboardCfg.FuzzelArgs...)

	// Add prompt based on mode
	if deleteFlag {
		fuzzelArgs = append(fuzzelArgs, "--prompt", "del > ")
		fuzzelArgs = append(fuzzelArgs, "--placeholder", "Delete from clipboard")
	} else {
		fuzzelArgs = append(fuzzelArgs, "--prompt", clipboardCfg.FuzzelPrompt)
		fuzzelArgs = append(fuzzelArgs, "--placeholder", "Type to search clipboard")
	}

	// Show fuzzel for selection
	fuzzel := exec.Command(fuzzelPath, fuzzelArgs...)
	fuzzel.Stdin = bytes.NewReader(clipOutput)

	chosenOutput, err := fuzzel.Output()
	if err != nil {
		// User cancelled selection (ESC or Ctrl+C)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// User cancelled, this is normal
			logger.Debug("User cancelled clipboard selection")
			return nil
		}
		return fmt.Errorf("failed to run fuzzel: %w", err)
	}

	// Handle the selected item
	if deleteFlag {
		// Delete from clipboard history
		deleteCmd := exec.Command(cliphistPath, "delete")
		deleteCmd.Stdin = bytes.NewReader(chosenOutput)

		if err := deleteCmd.Run(); err != nil {
			return fmt.Errorf("failed to delete from clipboard: %w", err)
		}

		// Delete from selection if configured
		if clipboardCfg.DeleteOnSelect {
			logger.Info("Item deleted from clipboard history")
		}
	} else {
		// Decode and copy to clipboard
		decodeCmd := exec.Command(cliphistPath, "decode")
		decodeCmd.Stdin = bytes.NewReader(chosenOutput)

		decodedOutput, err := decodeCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to decode clipboard item: %w", err)
		}

		// Use configured clipboard tool
		wlCopyPath := external.WlClipboard
		if wlCopyPath == "" {
			wlCopyPath = "wl-copy"
		}

		copyCmd := exec.Command(wlCopyPath)
		copyCmd.Stdin = bytes.NewReader(decodedOutput)

		if err := copyCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}

		logger.Debug("Clipboard item selected and copied")
	}

	return nil
}
