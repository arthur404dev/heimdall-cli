package record

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/heimdall-cli/heimdall/internal/utils/notify"
	"github.com/heimdall-cli/heimdall/internal/utils/paths"
	"github.com/spf13/cobra"
)

var (
	regionFlag string
	soundFlag  bool
)

// NewCommand creates the record command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Record the screen",
		Long:  `Start or stop screen recording using wl-screenrec`,
		RunE:  run,
	}

	cmd.Flags().StringVarP(&regionFlag, "region", "r", "", "Region to record (use 'slurp' for selection)")
	cmd.Flags().BoolVarP(&soundFlag, "sound", "s", false, "Record with sound")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	// Check if wl-screenrec is already running
	pidofCmd := exec.Command("pidof", "wl-screenrec")
	if err := pidofCmd.Run(); err == nil {
		// wl-screenrec is running, stop it
		return stopRecording()
	}

	// wl-screenrec is not running, start it
	return startRecording()
}

func startRecording() error {
	var args []string

	// Handle region selection
	if regionFlag != "" {
		var region string
		if regionFlag == "slurp" {
			// Use slurp to select region
			slurpCmd := exec.Command("slurp")
			output, err := slurpCmd.Output()
			if err != nil {
				return fmt.Errorf("failed to select region with slurp: %w", err)
			}
			region = strings.TrimSpace(string(output))
		} else {
			region = regionFlag
		}
		args = append(args, "-g", region)
	}

	// Handle audio recording
	if soundFlag {
		// Get audio sources
		pactlCmd := exec.Command("pactl", "list", "short", "sources")
		output, err := pactlCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to list audio sources: %w", err)
		}

		// Find running audio source
		var audioDevice string
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "RUNNING") {
				fields := strings.Fields(line)
				if len(fields) > 1 {
					audioDevice = fields[1]
					break
				}
			}
		}

		if audioDevice == "" {
			return fmt.Errorf("no audio source found")
		}

		args = append(args, "--audio", "--audio-device", audioDevice)
	}

	// Ensure recordings directory exists
	recordingPath := filepath.Join(paths.HeimdallStateDir, "recording.mp4")
	if err := paths.EnsureParentDir(recordingPath); err != nil {
		return fmt.Errorf("failed to create recording directory: %w", err)
	}

	// Start recording
	args = append(args, "-f", recordingPath)
	recordCmd := exec.Command("wl-screenrec", args...)
	recordCmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Start new session
	}

	if err := recordCmd.Start(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	// Wait a bit to see if the process started successfully
	time.Sleep(100 * time.Millisecond)

	// Check if process is still running
	if recordCmd.Process != nil {
		// Send notification
		notify.Send("Recording started", "Recording...")
	}

	return nil
}

func stopRecording() error {
	// Kill wl-screenrec
	if err := exec.Command("pkill", "wl-screenrec").Run(); err != nil {
		// Process might have already stopped
	}

	// Move recording to recordings folder
	recordingPath := filepath.Join(paths.HeimdallStateDir, "recording.mp4")
	recordingsDir := filepath.Join(paths.HeimdallDataDir, "recordings")

	// Ensure recordings directory exists
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create recordings directory: %w", err)
	}

	// Generate new filename with timestamp
	timestamp := time.Now().Format("20060102_15-04-05")
	newPath := filepath.Join(recordingsDir, fmt.Sprintf("recording_%s.mp4", timestamp))

	// Move the file
	if err := os.Rename(recordingPath, newPath); err != nil {
		return fmt.Errorf("failed to move recording: %w", err)
	}

	// Close start notification if it exists
	notifPath := filepath.Join(paths.HeimdallStateDir, "recording_notif")
	if notifData, err := os.ReadFile(notifPath); err == nil {
		notifID := string(notifData)
		// Try to close the notification
		closeCmd := exec.Command("gdbus", "call",
			"--session",
			"--dest=org.freedesktop.Notifications",
			"--object-path=/org/freedesktop/Notifications",
			"--method=org.freedesktop.Notifications.CloseNotification",
			notifID)
		closeCmd.Run() // Ignore errors
		os.Remove(notifPath)
	}

	// Send completion notification
	notify.Send("Recording stopped", fmt.Sprintf("Recording saved in %s", newPath))
	return nil
}
