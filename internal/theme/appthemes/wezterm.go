package appthemes

import (
	"os"
	"path/filepath"
	"github.com/arthur404dev/heimdall-cli/internal/config"
)

func init() {
	Register(&Template{
		Name:        "wezterm",
		Description: "Wezterm configuration",
		GetOutputPath: func() string {
			cfg := config.Get()
			if cfg != nil {
				if cfg.Theme.Paths.Wezterm != "" { return cfg.Theme.Paths.Wezterm }
			}
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".config", "wezterm", "colors", "heimdall.lua")
		},
		Content: `
-- Generated automatically

return {
  color_scheme = "Heimdall",
  color_schemes = {
    ["Heimdall"] = {
      background = "{{background}}",
      foreground = "{{foreground}}",
      cursor_bg = "{{foreground}}",
      cursor_fg = "{{background}}",
      cursor_border = "{{foreground}}",
      selection_bg = "{{colour8}}",
      selection_fg = "{{foreground}}",
      ansi = {
        "{{colour0}}", -- black
        "{{colour1}}", -- red
        "{{colour2}}", -- green
        "{{colour3}}", -- yellow
        "{{colour4}}", -- blue
        "{{colour5}}", -- magenta
        "{{colour6}}", -- cyan
        "{{colour7}}", -- white
      },
      brights = {
        "{{colour8}}",  -- bright black
        "{{colour9}}",  -- bright red
        "{{colour10}}", -- bright green
        "{{colour11}}", -- bright yellow
        "{{colour12}}", -- bright blue
        "{{colour13}}", -- bright magenta
        "{{colour14}}", -- bright cyan
        "{{colour15}}", -- bright white
      },
    },
  },
}
`,
	})
}
