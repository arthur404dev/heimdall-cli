package appthemes

func init() {
	Register(&Template{
		Name:        "qt",
		Aliases:     []string{"qt5", "qt6"},
		Description: "Qt theme configuration",
		Content: `
# Generated automatically

[ColorScheme]
active_colors={{foreground}}, {{background}}, {{colour8}}, {{colour0}}, {{colour8}}, {{colour7}}, {{foreground}}, {{foreground}}, {{foreground}}, {{background}}, {{background}}, {{colour8}}, {{colour4}}, {{foreground}}, {{colour4}}, {{colour5}}, {{colour0}}, {{foreground}}, {{background}}, {{foreground}}, {{colour8}}
disabled_colors={{colour8}}, {{background}}, {{colour8}}, {{colour0}}, {{colour8}}, {{colour8}}, {{colour8}}, {{colour8}}, {{colour8}}, {{background}}, {{background}}, {{colour8}}, {{colour0}}, {{colour8}}, {{colour4}}, {{colour5}}, {{colour0}}, {{colour8}}, {{background}}, {{colour8}}, {{colour8}}
inactive_colors={{foreground}}, {{background}}, {{colour8}}, {{colour0}}, {{colour8}}, {{colour7}}, {{foreground}}, {{foreground}}, {{foreground}}, {{background}}, {{background}}, {{colour8}}, {{colour4}}, {{foreground}}, {{colour4}}, {{colour5}}, {{colour0}}, {{foreground}}, {{background}}, {{foreground}}, {{colour8}}
`,
	})
}
