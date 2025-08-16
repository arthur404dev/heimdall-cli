package appthemes

func init() {
	Register(&Template{
		Name:        "waybar",
		Description: "Waybar configuration",
		Content: `
/* Heimdall theme for Waybar */

* {
    border: none;
    border-radius: 0;
    font-family: monospace;
    font-size: 13px;
}

window#waybar {
    background-color: {{background}};
    color: {{foreground}};
}

#workspaces button {
    background-color: {{colour0}};
    color: {{foreground}};
    padding: 0 5px;
}

#workspaces button.active {
    background-color: {{colour4}};
    color: {{background}};
}

#workspaces button:hover {
    background-color: {{colour8}};
    color: {{foreground}};
}

#clock, #battery, #cpu, #memory, #network, #pulseaudio {
    padding: 0 10px;
    color: {{foreground}};
}

#battery.charging {
    color: {{colour2}};
}

#battery.critical:not(.charging) {
    color: {{colour1}};
}

#network.disconnected {
    color: {{colour1}};
}

#pulseaudio.muted {
    color: {{colour8}};
}
`,
	})
}
