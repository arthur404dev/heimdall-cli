package appthemes

func init() {
	Register(&Template{
		Name:        "dunst",
		Description: "Dunst configuration",
		Content: `
# Heimdall theme for Dunst

[urgency_low]
    background = "{{background}}"
    foreground = "{{foreground}}"
    frame_color = "{{colour8}}"
    timeout = 10

[urgency_normal]
    background = "{{background}}"
    foreground = "{{foreground}}"
    frame_color = "{{colour4}}"
    timeout = 10

[urgency_critical]
    background = "{{background}}"
    foreground = "{{foreground}}"
    frame_color = "{{colour1}}"
    timeout = 0
`,
	})
}
