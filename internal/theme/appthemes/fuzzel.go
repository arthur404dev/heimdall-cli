package appthemes

import (
	"os"
	"path/filepath"
	"github.com/arthur404dev/heimdall-cli/internal/config"
)

func init() {
	Register(&Template{
		Name:        "fuzzel",
		Description: "Fuzzel configuration",
		GetOutputPath: func() string {
			cfg := config.Get()
			if cfg != nil {
				if cfg.Theme.Paths.Fuzzel != "" { return cfg.Theme.Paths.Fuzzel }
			}
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".config", "fuzzel", "colors.ini")
		},
		Content: `
# Generated automatically

[main]
font=monospace:size=10
dpi-aware=yes
width=30
horizontal-pad=20
vertical-pad=10
inner-pad=10

[colors]
background={{background}}dd
text={{foreground}}ff
match={{colour4}}ff
selection={{colour0}}ff
selection-text={{foreground}}ff
selection-match={{colour4}}ff
border={{colour8}}ff
`,
	})
}
