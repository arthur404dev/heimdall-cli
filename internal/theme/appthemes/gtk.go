package appthemes

func init() {
	Register(&Template{
		Name:        "gtk",
		Aliases:     []string{"gtk3", "gtk4"},
		Description: "GTK theme configuration",
		Content: `
/* Generated automatically */

@define-color background {{background}};
@define-color foreground {{foreground}};
@define-color color0 {{colour0}};
@define-color color1 {{colour1}};
@define-color color2 {{colour2}};
@define-color color3 {{colour3}};
@define-color color4 {{colour4}};
@define-color color5 {{colour5}};
@define-color color6 {{colour6}};
@define-color color7 {{colour7}};
@define-color color8 {{colour8}};
@define-color color9 {{colour9}};
@define-color color10 {{colour10}};
@define-color color11 {{colour11}};
@define-color color12 {{colour12}};
@define-color color13 {{colour13}};
@define-color color14 {{colour14}};
@define-color color15 {{colour15}};

/* Material Design 3 color mappings (if available) */
@define-color surface {{surface}};
@define-color surface_variant {{surface_variant}};
@define-color primary {{primary}};
@define-color primary_container {{primary_container}};
@define-color secondary {{secondary}};
@define-color secondary_container {{secondary_container}};
@define-color tertiary {{tertiary}};
@define-color tertiary_container {{tertiary_container}};
@define-color error {{error}};
@define-color error_container {{error_container}};
@define-color outline {{outline}};
@define-color outline_variant {{outline_variant}};

/* Apply to GTK widgets */
window {
    background-color: @background;
    color: @foreground;
}

button {
    background-color: @color4;
    color: @foreground;
}

button:hover {
    background-color: @color12;
}

entry {
    background-color: @color0;
    color: @foreground;
    border-color: @color8;
}

/* Scrollbars */
scrollbar {
    background-color: @background;
}

scrollbar slider {
    background-color: @color8;
}

scrollbar slider:hover {
    background-color: @color7;
}
`,
	})
}
