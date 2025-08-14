package schemes

import (
	"embed"
)

// Content embeds all Material You scheme files in JSON format
//
//go:embed catppuccin/*/*.json
//go:embed gruvbox/*/*.json
//go:embed rosepine/*/*.json
//go:embed onedark/*/*.json
//go:embed oldworld/*/*.json
//go:embed shadotheme/*/*.json
var Content embed.FS
