package schemes

import (
	"embed"
)

// Content embeds all Material You scheme files in .txt format
//
//go:embed catppuccin/*/*.txt
//go:embed gruvbox/*/*.txt
//go:embed rosepine/*/*.txt
//go:embed onedark/*/*.txt
//go:embed oldworld/*/*.txt
//go:embed shadotheme/*/*.txt
var Content embed.FS
