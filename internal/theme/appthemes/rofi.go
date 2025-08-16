package appthemes

func init() {
	Register(&Template{
		Name:        "rofi",
		Description: "Rofi configuration",
		Content: `
/* Heimdall theme for Rofi */

* {
    background: {{background}};
    foreground: {{foreground}};
    selected-background: {{colour4}};
    selected-foreground: {{background}};
    alternate-background: {{colour0}};
    border-color: {{colour8}};
}

window {
    background-color: @background;
    border: 2px;
    border-color: @border-color;
    padding: 10px;
}

mainbox {
    background-color: @background;
}

inputbar {
    background-color: @alternate-background;
    text-color: @foreground;
    padding: 10px;
}

entry {
    background-color: @alternate-background;
    text-color: @foreground;
}

listview {
    background-color: @background;
}

element {
    background-color: @background;
    text-color: @foreground;
    padding: 5px;
}

element selected {
    background-color: @selected-background;
    text-color: @selected-foreground;
}
`,
	})
}
