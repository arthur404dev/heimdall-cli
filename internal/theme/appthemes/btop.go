package appthemes

import (
	"os"
	"path/filepath"
	"github.com/arthur404dev/heimdall-cli/internal/config"
)

func init() {
	Register(&Template{
		Name:        "btop",
		Description: "Btop configuration",
		GetOutputPath: func() string {
			cfg := config.Get()
			if cfg != nil {
				if cfg.Theme.Paths.Btop != "" { return cfg.Theme.Paths.Btop }
			}
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".config", "btop", "themes", "heimdall.theme")
		},
		Content: `
# Generated automatically

# Main background and foreground
theme[main_bg]="{{background}}"
theme[main_fg]="{{foreground}}"

# Title
theme[title]="{{foreground}}"

# Highlight
theme[hi_fg]="{{colour4}}"

# Selected
theme[selected_bg]="{{colour8}}"
theme[selected_fg]="{{colour7}}"

# Status
theme[inactive_fg]="{{colour8}}"
theme[graph_text]="{{foreground}}"

# Process box
theme[proc_misc]="{{colour5}}"

# CPU box
theme[cpu_box]="{{colour4}}"
theme[cpu_text]="{{colour7}}"

# Memory/Disk box
theme[mem_box]="{{colour5}}"
theme[mem_text]="{{colour7}}"

# Network box
theme[net_box]="{{colour6}}"
theme[net_text]="{{colour7}}"

# Process list
theme[proc_box]="{{colour0}}"
theme[proc_text]="{{foreground}}"
`,
	})
}
