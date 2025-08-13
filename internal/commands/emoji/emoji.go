package emoji

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/cobra"
)

// Emoji represents an emoji entry
type Emoji struct {
	Emoji    string   `json:"emoji"`
	Aliases  []string `json:"aliases"`
	Tags     []string `json:"tags"`
	Category string   `json:"category"`
	Unicode  string   `json:"unicode_version"`
}

// Command creates the emoji command
func Command() *cobra.Command {
	var (
		fetch  bool
		picker bool
	)

	cmd := &cobra.Command{
		Use:   "emoji",
		Short: "Emoji and glyph utilities",
		Long: `Emoji and glyph utilities for picking and fetching emoji data.
		
Features:
  - Interactive emoji picker using fuzzel
  - Fetch and update emoji database
  - Copy emoji to clipboard`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fetch {
				return updateEmojiData()
			}

			if picker || len(args) == 0 {
				return runEmojiPicker()
			}

			// Search for emoji by name
			query := strings.Join(args, " ")
			return searchEmoji(query)
		},
	}

	cmd.Flags().BoolVarP(&fetch, "fetch", "f", false, "Fetch emoji/glyph data from remote")
	cmd.Flags().BoolVarP(&picker, "picker", "p", false, "Run interactive emoji picker")

	return cmd
}

// updateEmojiData fetches the latest emoji data
func updateEmojiData() error {
	logger.Info("Updating emoji database...")

	// Emoji data sources
	sources := []struct {
		name string
		url  string
		file string
	}{
		{
			name: "emoji",
			url:  "https://raw.githubusercontent.com/github/gemoji/master/db/emoji.json",
			file: "emoji.json",
		},
		{
			name: "nerd-fonts",
			url:  "https://raw.githubusercontent.com/ryanoasis/nerd-fonts/master/glyphnames.json",
			file: "nerd-fonts.json",
		},
	}

	dataDir := filepath.Join(paths.DataDir, "emoji")
	if err := paths.EnsureDir(dataDir); err != nil {
		return fmt.Errorf("failed to create emoji data directory: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for _, source := range sources {
		logger.Info("Fetching", "source", source.name, "url", source.url)

		resp, err := client.Get(source.url)
		if err != nil {
			logger.Error("Failed to fetch", "source", source.name, "error", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Error("Bad response", "source", source.name, "status", resp.StatusCode)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Failed to read response", "source", source.name, "error", err)
			continue
		}

		filePath := filepath.Join(dataDir, source.file)
		if err := paths.AtomicWrite(filePath, data); err != nil {
			logger.Error("Failed to save", "source", source.name, "error", err)
			continue
		}

		logger.Info("Updated", "source", source.name, "file", filePath)
	}

	fmt.Println("Emoji database updated successfully")
	return nil
}

// runEmojiPicker runs the interactive emoji picker
func runEmojiPicker() error {
	// Load emoji data
	emojis, err := loadEmojiData()
	if err != nil {
		// Try to update if no data exists
		if os.IsNotExist(err) {
			logger.Info("No emoji data found, fetching...")
			if err := updateEmojiData(); err != nil {
				return fmt.Errorf("failed to fetch emoji data: %w", err)
			}
			emojis, err = loadEmojiData()
			if err != nil {
				return fmt.Errorf("failed to load emoji data: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load emoji data: %w", err)
		}
	}

	// Create temporary file with emoji list for fuzzel
	tmpFile, err := os.CreateTemp("", "emoji-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write emojis to temp file
	for _, emoji := range emojis {
		line := fmt.Sprintf("%s %s %s\n",
			emoji.Emoji,
			strings.Join(emoji.Aliases, " "),
			emoji.Category)
		tmpFile.WriteString(line)
	}
	tmpFile.Close()

	// Load config for fuzzel path
	if err := config.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := config.Get()

	fuzzelCmd := cfg.External.Fuzzel
	if fuzzelCmd == "" {
		fuzzelCmd = "fuzzel"
	}

	// Run fuzzel
	cmd := exec.Command(fuzzelCmd, "--dmenu", "--prompt", "Emoji> ")

	// Pipe emoji list to fuzzel
	input, err := os.Open(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer input.Close()

	cmd.Stdin = input

	output, err := cmd.Output()
	if err != nil {
		// User cancelled
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("failed to run fuzzel: %w", err)
	}

	// Extract emoji from output (first field)
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return nil
	}

	parts := strings.Fields(selected)
	if len(parts) == 0 {
		return nil
	}

	emoji := parts[0]

	// Copy to clipboard
	if err := copyToClipboard(emoji); err != nil {
		logger.Error("Failed to copy to clipboard", "error", err)
		fmt.Println(emoji)
	} else {
		fmt.Printf("Copied %s to clipboard\n", emoji)
	}

	return nil
}

// searchEmoji searches for an emoji by name
func searchEmoji(query string) error {
	emojis, err := loadEmojiData()
	if err != nil {
		return fmt.Errorf("failed to load emoji data: %w", err)
	}

	query = strings.ToLower(query)
	var matches []Emoji

	for _, emoji := range emojis {
		// Check aliases
		for _, alias := range emoji.Aliases {
			if strings.Contains(strings.ToLower(alias), query) {
				matches = append(matches, emoji)
				break
			}
		}

		// Check tags
		if len(matches) == 0 || matches[len(matches)-1].Emoji != emoji.Emoji {
			for _, tag := range emoji.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					matches = append(matches, emoji)
					break
				}
			}
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no emoji found for query: %s", query)
	}

	// Display matches
	for _, emoji := range matches {
		fmt.Printf("%s  %s  [%s]\n",
			emoji.Emoji,
			strings.Join(emoji.Aliases, ", "),
			emoji.Category)
	}

	// Copy first match to clipboard
	if len(matches) > 0 {
		if err := copyToClipboard(matches[0].Emoji); err != nil {
			logger.Error("Failed to copy to clipboard", "error", err)
		}
	}

	return nil
}

// loadEmojiData loads emoji data from disk
func loadEmojiData() ([]Emoji, error) {
	dataFile := filepath.Join(paths.DataDir, "emoji", "emoji.json")

	data, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, err
	}

	var emojis []Emoji
	if err := json.Unmarshal(data, &emojis); err != nil {
		return nil, fmt.Errorf("failed to parse emoji data: %w", err)
	}

	return emojis, nil
}

// copyToClipboard copies text to clipboard using wl-copy
func copyToClipboard(text string) error {
	// Load config for wl-clipboard path
	if err := config.Load(); err != nil {
		return err
	}
	cfg := config.Get()

	wlCopy := cfg.External.WlClipboard
	if wlCopy == "" {
		wlCopy = "wl-copy"
	}

	cmd := exec.Command(wlCopy)
	cmd.Stdin = strings.NewReader(text)

	return cmd.Run()
}
