package appthemes

import (
	"github.com/arthur404dev/heimdall-cli/internal/config"
	"os"
	"path/filepath"
)

func init() {
	Register(&Template{
		Name:        "alacritty",
		Description: "Alacritty terminal configuration",
		GetOutputPath: func() string {
			cfg := config.Get()
			if cfg != nil && cfg.Theme.Paths.Alacritty != "" {
				return cfg.Theme.Paths.Alacritty
			}
			// Default path
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".config", "alacritty", "themes", "heimdall.toml")
		},
		Content: `
# Heimdall theme for Alacritty
# Generated automatically

[colors.primary]
background = "{{background}}"
foreground = "{{foreground}}"

[colors.cursor]
text = "{{background}}"
cursor = "{{foreground}}"

[colors.normal]
black = "{{colour0}}"
red = "{{colour1}}"
green = "{{colour2}}"
yellow = "{{colour3}}"
blue = "{{colour4}}"
magenta = "{{colour5}}"
cyan = "{{colour6}}"
white = "{{colour7}}"

[colors.bright]
black = "{{colour8}}"
red = "{{colour9}}"
green = "{{colour10}}"
yellow = "{{colour11}}"
blue = "{{colour12}}"
magenta = "{{colour13}}"
cyan = "{{colour14}}"
white = "{{colour15}}"

[colors.selection]
text = "{{background}}"
background = "{{foreground}}"
`,
	})
}
