package clipboard

import (
	"bytes"
	"fmt"
	"os/exec"

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

	// Get clipboard history from cliphist
	cliphist := exec.Command("cliphist", "list")
	clipOutput, err := cliphist.Output()
	if err != nil {
		return fmt.Errorf("failed to get clipboard history: %w", err)
	}

	// Prepare fuzzel arguments
	var fuzzelArgs []string
	fuzzelArgs = append(fuzzelArgs, "--dmenu")

	if deleteFlag {
		fuzzelArgs = append(fuzzelArgs, "--prompt=del > ", "--placeholder=Delete from clipboard")
	} else {
		fuzzelArgs = append(fuzzelArgs, "--placeholder=Type to search clipboard")
	}

	// Show fuzzel for selection
	fuzzel := exec.Command("fuzzel", fuzzelArgs...)
	fuzzel.Stdin = bytes.NewReader(clipOutput)

	chosenOutput, err := fuzzel.Output()
	if err != nil {
		// User cancelled selection (ESC or Ctrl+C)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// User cancelled, this is normal
			return nil
		}
		return fmt.Errorf("failed to run fuzzel: %w", err)
	}

	// Handle the selected item
	if deleteFlag {
		// Delete from clipboard history
		deleteCmd := exec.Command("cliphist", "delete")
		deleteCmd.Stdin = bytes.NewReader(chosenOutput)

		if err := deleteCmd.Run(); err != nil {
			return fmt.Errorf("failed to delete from clipboard: %w", err)
		}
	} else {
		// Decode and copy to clipboard
		decodeCmd := exec.Command("cliphist", "decode")
		decodeCmd.Stdin = bytes.NewReader(chosenOutput)

		decodedOutput, err := decodeCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to decode clipboard item: %w", err)
		}

		copyCmd := exec.Command("wl-copy")
		copyCmd.Stdin = bytes.NewReader(decodedOutput)

		if err := copyCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
	}

	return nil
}
