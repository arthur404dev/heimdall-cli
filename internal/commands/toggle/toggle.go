package toggle

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/utils/hypr"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/cobra"
)

// ClientConfig represents configuration for a client/application
type ClientConfig struct {
	Enable  bool                     `json:"enable"`
	Match   []map[string]interface{} `json:"match"`
	Command []string                 `json:"command,omitempty"`
	Move    bool                     `json:"move,omitempty"`
}

// ToggleConfig represents the full toggle configuration
type ToggleConfig struct {
	Communication map[string]ClientConfig `json:"communication"`
	Music         map[string]ClientConfig `json:"music"`
	Sysmon        map[string]ClientConfig `json:"sysmon"`
	Todo          map[string]ClientConfig `json:"todo"`
}

var (
	defaultConfig = ToggleConfig{
		Communication: map[string]ClientConfig{
			"discord": {
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "discord"},
				},
				Command: []string{"discord"},
				Move:    true,
			},
			"whatsapp": {
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "whatsapp"},
				},
				Move: true,
			},
		},
		Music: map[string]ClientConfig{
			"spotify": {
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "Spotify"},
					{"initialTitle": "Spotify"},
					{"initialTitle": "Spotify Free"},
				},
				Command: []string{"spicetify", "watch", "-s"},
				Move:    true,
			},
			"feishin": {
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "feishin"},
				},
				Move: true,
			},
		},
		Sysmon: map[string]ClientConfig{
			"btop": {
				Enable: true,
				Match: []map[string]interface{}{
					{
						"class":     "btop",
						"title":     "btop",
						"workspace": map[string]interface{}{"name": "special:sysmon"},
					},
				},
				Command: []string{"foot", "-a", "btop", "-T", "btop", "fish", "-C", "exec btop"},
			},
		},
		Todo: map[string]ClientConfig{
			"todoist": {
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "Todoist"},
				},
				Command: []string{"todoist"},
				Move:    true,
			},
		},
	}
)

// NewCommand creates the toggle command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toggle [workspace]",
		Short: "Toggle special workspaces",
		Long: `Toggle special workspaces in Hyprland. Available workspaces:
  - communication: Discord, WhatsApp
  - music: Spotify, Feishin
  - sysmon: System monitor (btop)
  - specialws: Smart toggle for special workspaces
  - todo: Todoist`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"communication", "music", "sysmon", "specialws", "todo"},
		RunE:      run,
	}

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	workspace := args[0]

	// Create Hyprland client
	client, err := hypr.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Hyprland client: %w", err)
	}

	// Handle special workspace toggle
	if workspace == "specialws" {
		return handleSpecialWorkspace(client)
	}

	// Load configuration
	config := loadConfig()

	// Get the appropriate workspace config
	var workspaceConfig map[string]ClientConfig
	switch workspace {
	case "communication":
		workspaceConfig = config.Communication
	case "music":
		workspaceConfig = config.Music
	case "sysmon":
		workspaceConfig = config.Sysmon
	case "todo":
		workspaceConfig = config.Todo
	default:
		return fmt.Errorf("unknown workspace: %s", workspace)
	}

	// Get current windows
	windows, err := client.GetWindows()
	if err != nil {
		return fmt.Errorf("failed to get windows: %w", err)
	}

	// Handle each client configuration
	for _, clientConfig := range workspaceConfig {
		if clientConfig.Enable {
			handleClientConfig(client, clientConfig, workspace, windows)
		}
	}

	// Toggle the special workspace
	if err := client.Dispatch("togglespecialworkspace", workspace); err != nil {
		return fmt.Errorf("failed to toggle workspace: %w", err)
	}

	return nil
}

func loadConfig() ToggleConfig {
	config := defaultConfig

	// Try to load user configuration
	configPath := filepath.Join(paths.HeimdallConfigDir, "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var userConfig struct {
			Toggles ToggleConfig `json:"toggles"`
		}
		if err := json.Unmarshal(data, &userConfig); err == nil {
			// Merge user config with default config
			mergeConfig(&config, &userConfig.Toggles)
		}
	}

	return config
}

func mergeConfig(dst, src *ToggleConfig) {
	// Merge communication configs
	for k, v := range src.Communication {
		dst.Communication[k] = v
	}
	// Merge music configs
	for k, v := range src.Music {
		dst.Music[k] = v
	}
	// Merge sysmon configs
	for k, v := range src.Sysmon {
		dst.Sysmon[k] = v
	}
	// Merge todo configs
	for k, v := range src.Todo {
		dst.Todo[k] = v
	}
}

func handleClientConfig(client *hypr.Client, config ClientConfig, workspace string, windows []hypr.Window) {
	// Check if we need to spawn the client
	if len(config.Command) > 0 {
		shouldSpawn := true
		for _, w := range windows {
			if matchesWindow(w, config.Match) {
				shouldSpawn = false
				break
			}
		}

		if shouldSpawn {
			spawnClient(config.Command)
		}
	}

	// Move matching windows to the special workspace
	if config.Move {
		for _, w := range windows {
			if matchesWindow(w, config.Match) {
				targetWorkspace := fmt.Sprintf("special:%s", workspace)
				if w.Workspace.Name != targetWorkspace {
					client.Dispatch("movetoworkspacesilent", fmt.Sprintf("%s,address:%s", targetWorkspace, w.Address))
				}
			}
		}
	}
}

func matchesWindow(window hypr.Window, matches []map[string]interface{}) bool {
	for _, match := range matches {
		if isSubset(windowToMap(window), match) {
			return true
		}
	}
	return false
}

func windowToMap(window hypr.Window) map[string]interface{} {
	return map[string]interface{}{
		"address":      window.Address,
		"class":        window.Class,
		"title":        window.Title,
		"initialTitle": window.InitialTitle,
		"workspace": map[string]interface{}{
			"id":   window.Workspace.ID,
			"name": window.Workspace.Name,
		},
	}
}

func isSubset(superset, subset map[string]interface{}) bool {
	for key, value := range subset {
		superValue, exists := superset[key]
		if !exists {
			return false
		}

		switch v := value.(type) {
		case map[string]interface{}:
			superMap, ok := superValue.(map[string]interface{})
			if !ok || !isSubset(superMap, v) {
				return false
			}
		case string:
			superStr, ok := superValue.(string)
			if !ok || !strings.Contains(superStr, v) {
				return false
			}
		default:
			if superValue != value {
				return false
			}
		}
	}
	return true
}

func spawnClient(command []string) {
	if len(command) == 0 {
		return
	}

	// Check if app2unit exists
	if _, err := exec.LookPath("app2unit"); err == nil {
		cmd := exec.Command("app2unit", append([]string{"--"}, command...)...)
		cmd.Start()
	} else {
		// Fallback to direct execution
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Start()
	}
}

func handleSpecialWorkspace(client *hypr.Client) error {
	workspaces, err := client.GetWorkspaces()
	if err != nil {
		return fmt.Errorf("failed to get workspaces: %w", err)
	}

	// Check if special:special workspace exists
	onSpecialWs := false
	for _, ws := range workspaces {
		if ws.Name == "special:special" {
			onSpecialWs = true
			break
		}
	}

	toggleWs := "special"

	if !onSpecialWs {
		// Get active window to check current workspace
		windows, err := client.GetWindows()
		if err == nil {
			// Find the focused window
			for _, w := range windows {
				// Check if this window is on a special workspace
				if strings.HasPrefix(w.Workspace.Name, "special:") {
					toggleWs = w.Workspace.Name[8:] // Remove "special:" prefix
					break
				}
			}
		}
	}

	return client.Dispatch("togglespecialworkspace", toggleWs)
}
