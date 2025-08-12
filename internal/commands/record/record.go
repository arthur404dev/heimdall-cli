package record

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/notify"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
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
	cfg := config.Get()
	external := cfg.External

	// Check if wl-screenrec is already running
	pidofPath := external.Pidof
	if pidofPath == "" {
		pidofPath = "pidof"
	}

	pidofCmd := exec.Command(pidofPath, "wl-screenrec")
	if err := pidofCmd.Run(); err == nil {
		// wl-screenrec is running, stop it
		return stopRecording()
	}

	// wl-screenrec is not running, start it
	return startRecording()
}

func startRecording() error {
	cfg := config.Get()
	external := cfg.External
	recordingCfg := cfg.Recording

	var args []string

	// Handle region selection
	if regionFlag != "" {
		var region string
		if regionFlag == "slurp" {
			// Use slurp to select region
			slurpPath := external.Slurp
			if slurpPath == "" {
				slurpPath = "slurp"
			}
			slurpCmd := exec.Command(slurpPath)
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
	if soundFlag && recordingCfg.AudioSource != "none" {
		// Get audio sources
		pactlPath := external.Pactl
		if pactlPath == "" {
			pactlPath = "pactl"
		}
		pactlCmd := exec.Command(pactlPath, "list", "short", "sources")
		output, err := pactlCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to list audio sources: %w", err)
		}

		// Find running audio source
		var audioDevice string
		if recordingCfg.AudioSource == "auto" {
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
		} else {
			audioDevice = recordingCfg.AudioSource
		}

		if audioDevice == "" {
			return fmt.Errorf("no audio source found")
		}

		args = append(args, "--audio", "--audio-device", audioDevice)
	}

	// Ensure state directory exists
	stateDir := filepath.Join(paths.HeimdallStateDir)
	if err := paths.EnsureDir(stateDir); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Use configured temp file name
	recordingPath := filepath.Join(stateDir, recordingCfg.TempFileName)
	if err := paths.EnsureParentDir(recordingPath); err != nil {
		return fmt.Errorf("failed to create recording directory: %w", err)
	}

	// Start recording
	args = append(args, "-f", recordingPath)
	wlScreenrecPath := external.WlScreenrec
	if wlScreenrecPath == "" {
		wlScreenrecPath = "wl-screenrec"
	}
	recordCmd := exec.Command(wlScreenrecPath, args...)
	recordCmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Start new session
	}

	if err := recordCmd.Start(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	// Wait a bit to see if the process started successfully
	time.Sleep(100 * time.Millisecond)

	// Check if process is still running
	if recordCmd.Process != nil && recordingCfg.ShowNotification {
		// Send notification
		notify.Send("Recording started", "Recording...")
	}

	return nil
}

func stopRecording() error {
	cfg := config.Get()
	external := cfg.External
	recordingCfg := cfg.Recording

	// Kill wl-screenrec
	pkillPath := external.Pkill
	if pkillPath == "" {
		pkillPath = "pkill"
	}
	if err := exec.Command(pkillPath, "wl-screenrec").Run(); err != nil {
		// Process might have already stopped
	}

	// Move recording to recordings folder
	recordingPath := filepath.Join(paths.HeimdallStateDir, recordingCfg.TempFileName)

	// Use configured recordings directory
	recordingsDir := recordingCfg.Directory
	if recordingsDir == "" {
		recordingsDir = paths.RecordingsDir
	}

	// Ensure recordings directory exists
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create recordings directory: %w", err)
	}

	// Generate new filename with configured pattern
	filePattern := recordingCfg.FileNamePattern
	if filePattern == "" {
		filePattern = "recording_%Y%m%d_%H%M%S"
	}
	// Replace timestamp patterns
	filename := strings.ReplaceAll(filePattern, "%Y%m%d", time.Now().Format("20060102"))
	filename = strings.ReplaceAll(filename, "%H%M%S", time.Now().Format("150405"))
	newPath := filepath.Join(recordingsDir, fmt.Sprintf("%s.%s", filename, recordingCfg.FileFormat))

	// Move the file
	if err := os.Rename(recordingPath, newPath); err != nil {
		return fmt.Errorf("failed to move recording: %w", err)
	}

	// Close start notification if it exists
	notifPath := filepath.Join(paths.HeimdallStateDir, "recording_notif")
	if notifData, err := os.ReadFile(notifPath); err == nil {
		notifID := string(notifData)
		// Try to close the notification
		gdbusPath := external.Gdbus
		if gdbusPath == "" {
			gdbusPath = "gdbus"
		}
		closeCmd := exec.Command(gdbusPath, "call",
			"--session",
			"--dest=org.freedesktop.Notifications",
			"--object-path=/org/freedesktop/Notifications",
			"--method=org.freedesktop.Notifications.CloseNotification",
			notifID)
		closeCmd.Run() // Ignore errors
		os.Remove(notifPath)
	}

	// Send completion notification if configured
	if recordingCfg.ShowNotification {
		notify.Send("Recording stopped", fmt.Sprintf("Recording saved in %s", newPath))
	}
	return nil
}
