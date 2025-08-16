package appthemes

import (
	"os"
	"path/filepath"
	"github.com/arthur404dev/heimdall-cli/internal/config"
)

func init() {
	Register(&Template{
		Name:        "spicetify",
		Description: "Spicetify configuration",
		GetOutputPath: func() string {
			cfg := config.Get()
			if cfg != nil {
				if cfg.Theme.Paths.Spicetify != "" { return cfg.Theme.Paths.Spicetify }
			}
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".config", "spicetify", "Themes", "heimdall", "color.ini")
		},
		Content: `
# Generated automatically

[Base]
main_bg = {{background.raw}}
sidebar_bg = {{colour0.raw}}
player_bg = {{colour8.raw}}
card_bg = {{colour0.raw}}
shadow = 000000
main_fg = {{foreground.raw}}
sidebar_fg = {{foreground.raw}}
secondary_fg = {{colour7.raw}}
selected_button = {{colour4.raw}}
pressing_button_bg = {{colour0.raw}}
pressing_button_fg = {{foreground.raw}}
miscellaneous_bg = {{colour8.raw}}
miscellaneous_hover_bg = {{colour0.raw}}
preserve_1 = ffffff
`,
	})
}
