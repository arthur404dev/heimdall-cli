package appthemes

import (
	"os"
	"path/filepath"
	"github.com/arthur404dev/heimdall-cli/internal/config"
)

func init() {
	Register(&Template{
		Name:        "terminal",
		Description: "Terminal configuration",
		GetOutputPath: func() string {
			cfg := config.Get()
			if cfg != nil {
				if cfg.Theme.Paths.Terminal != "" { return cfg.Theme.Paths.Terminal }
			}
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".config", "heimdall", "sequences.txt")
		},
		Content: `
# Heimdall Terminal Color Sequences

# Special colors
printf '\033]10;{{foreground}}\007'  # foreground
printf '\033]11;{{background}}\007'  # background
printf '\033]12;{{cursor}}\007'  # cursor

# Standard colors
printf '\033]4;0;{{colour0}}\007'   # black
printf '\033]4;1;{{colour1}}\007'   # red
printf '\033]4;2;{{colour2}}\007'   # green
printf '\033]4;3;{{colour3}}\007'   # yellow
printf '\033]4;4;{{colour4}}\007'   # blue
printf '\033]4;5;{{colour5}}\007'   # magenta
printf '\033]4;6;{{colour6}}\007'   # cyan
printf '\033]4;7;{{colour7}}\007'   # white
printf '\033]4;8;{{colour8}}\007'   # bright black
printf '\033]4;9;{{colour9}}\007'   # bright red
printf '\033]4;10;{{colour10}}\007' # bright green
printf '\033]4;11;{{colour11}}\007' # bright yellow
printf '\033]4;12;{{colour12}}\007' # bright blue
printf '\033]4;13;{{colour13}}\007' # bright magenta
printf '\033]4;14;{{colour14}}\007' # bright cyan
printf '\033]4;15;{{colour15}}\007' # bright white
`,
	})
}
