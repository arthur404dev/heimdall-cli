package appthemes

import (
	"os"
	"path/filepath"
	"github.com/arthur404dev/heimdall-cli/internal/config"
)

func init() {
	Register(&Template{
		Name:        "kitty",
		Description: "Kitty configuration",
		GetOutputPath: func() string {
			cfg := config.Get()
			if cfg != nil {
				if cfg.Theme.Paths.Kitty != "" { return cfg.Theme.Paths.Kitty }
			}
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".config", "kitty", "themes", "heimdall.conf")
		},
		Content: `
# Generated automatically

foreground {{foreground}}
background {{background}}
cursor {{cursor}}

# Black
color0 {{colour0}}
color8 {{colour8}}

# Red
color1 {{colour1}}
color9 {{colour9}}

# Green
color2 {{colour2}}
color10 {{colour10}}

# Yellow
color3 {{colour3}}
color11 {{colour11}}

# Blue
color4 {{colour4}}
color12 {{colour12}}

# Magenta
color5 {{colour5}}
color13 {{colour13}}

# Cyan
color6 {{colour6}}
color14 {{colour14}}

# White
color7 {{colour7}}
color15 {{colour15}}

# Additional Kitty-specific settings
selection_foreground {{background}}
selection_background {{foreground}}
url_color {{colour4}}
active_border_color {{colour4}}
inactive_border_color {{colour8}}
bell_border_color {{colour3}}

# Tab bar
active_tab_foreground {{background}}
active_tab_background {{colour5}}
inactive_tab_foreground {{foreground}}
inactive_tab_background {{colour0}}
tab_bar_background {{background}}
`,
	})
}
