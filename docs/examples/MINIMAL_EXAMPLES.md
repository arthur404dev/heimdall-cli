# Minimal Configuration Examples

These minimal configuration files demonstrate common use cases with only the necessary settings.
Heimdall will use default values for any settings not specified.

## minimal-theme-only.json

Minimal config for theme application only

```json
{
  "theme": {
    "enableDiscord": true,
    "enableGtk": true,
    "enableQt": true
  },
  "version": "0.2.0"
}
```

## minimal-wallpaper-only.json

Minimal config for wallpaper management only

```json
{
  "version": "0.2.0",
  "wallpaper": {
    "directory": "~/Pictures/Wallpapers",
    "filter": true,
    "smart_mode": true
  }
}
```

## minimal-scheme-only.json

Minimal config for color scheme management

```json
{
  "scheme": {
    "auto_mode": true,
    "default": "catppuccin-mocha",
    "material_you": false
  },
  "version": "0.2.0"
}
```

## minimal-terminal-only.json

Minimal config for terminal theming only

```json
{
  "theme": {
    "enableKitty": true,
    "enableTerm": true
  },
  "version": "0.2.0"
}
```

## minimal-material-you.json

Minimal config for Material You wallpaper-based theming

```json
{
  "scheme": {
    "material_you": true
  },
  "version": "0.2.0",
  "wallpaper": {
    "directory": "~/Pictures/Wallpapers",
    "smart_mode": true
  }
}
```

## minimal-quickshell.json

Minimal config for Quickshell integration

```json
{
  "shell": {
    "args": [
      "-c",
      "heimdall",
      "-n"
    ],
    "command": "qs"
  },
  "version": "0.2.0"
}
```

