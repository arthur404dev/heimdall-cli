# Theme Integration Guide

This guide explains how to integrate Heimdall's generated theme files with your applications.

## Non-Invasive Design

Heimdall is designed to be **non-invasive** - it creates separate theme files that you can include/import into your main configuration files. This way, your personal configurations remain untouched.

## Terminal Emulators

### Kitty
Heimdall creates: `~/.config/kitty/themes/heimdall.conf`

Add to your `~/.config/kitty/kitty.conf`:
```conf
include themes/heimdall.conf
```

### Alacritty
Heimdall creates: `~/.config/alacritty/themes/heimdall.toml`

Add to your `~/.config/alacritty/alacritty.toml`:
```toml
import = ["~/.config/alacritty/themes/heimdall.toml"]
```

### WezTerm
Heimdall creates: `~/.config/wezterm/colors/heimdall.lua`

In your `~/.config/wezterm/wezterm.lua`:
```lua
local config = wezterm.config_builder()
config.color_scheme_dirs = { "~/.config/wezterm/colors" }
config.color_scheme = "heimdall"
return config
```

## Desktop Applications

### GTK 3/4
Heimdall creates: 
- `~/.config/gtk-3.0/colors.css`
- `~/.config/gtk-4.0/colors.css`

Add to your `~/.config/gtk-3.0/gtk.css`:
```css
@import url("colors.css");
```

Add to your `~/.config/gtk-4.0/gtk.css`:
```css
@import url("colors.css");
```

### Qt5/Qt6
Heimdall creates:
- `~/.config/qt5ct/colors/heimdall.conf`
- `~/.config/qt6ct/colors/heimdall.conf`

In qt5ct/qt6ct GUI, select "heimdall" from the color scheme dropdown.

### Fuzzel
Heimdall creates: `~/.config/fuzzel/colors.ini`

Add to your `~/.config/fuzzel/fuzzel.ini`:
```ini
include=colors.ini
```

### Btop
Heimdall creates: `~/.config/btop/themes/heimdall.theme`

In btop, press `Esc` → `Options` → `Color theme` → Select `heimdall`

### Spicetify
Heimdall creates: `~/.config/spicetify/Themes/heimdall/color.ini`

Apply with:
```bash
spicetify config current_theme heimdall
spicetify apply
```

## Discord Clients

Discord clients automatically load themes from their respective `themes/` directories:
- Vesktop: `~/.config/vesktop/themes/heimdall.css`
- Vencord: `~/.config/Vencord/themes/heimdall.css`
- BetterDiscord: `~/.config/BetterDiscord/themes/heimdall.theme.css`

Enable the theme in your Discord client's settings.

## Terminal Sequences

Heimdall creates: `~/.config/heimdall/sequences.txt`

This file contains ANSI escape sequences that can be sourced in your shell:
```bash
# Add to ~/.bashrc or ~/.zshrc
[ -f ~/.config/heimdall/sequences.txt ] && source ~/.config/heimdall/sequences.txt
```

## Custom Paths

You can customize where theme files are created by editing `~/.config/heimdall/config.json`:

```json
{
  "theme": {
    "paths": {
      "kitty": "/custom/path/to/kitty/theme.conf",
      "alacritty": "/custom/path/to/alacritty/theme.toml"
    }
  }
}
```

## Refreshing Configuration

After updating Heimdall, refresh your configuration to get new theme paths:
```bash
heimdall config refresh
```

Or reset to defaults if needed:
```bash
heimdall config defaults
```