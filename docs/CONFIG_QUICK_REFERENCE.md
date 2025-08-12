# Configuration Quick Reference

## File Location
`~/.config/heimdall/config.json`

## Quick Commands

```bash
# View current configuration
cat ~/.config/heimdall/config.json | jq .

# Edit configuration
nano ~/.config/heimdall/config.json

# Validate JSON syntax
jq . ~/.config/heimdall/config.json

# Reset to defaults
rm ~/.config/heimdall/config.json && heimdall scheme list
```

## Most Common Settings

### Change Screenshot Directory
```json
{
  "screenshot": {
    "directory": "/home/user/Screenshots"
  }
}
```

### Disable Notifications
```json
{
  "notification": {
    "enabled": false
  }
}
```

### Custom Wallpaper Directory
```json
{
  "wallpaper": {
    "directory": "/home/user/Pictures/Wallpapers"
  }
}
```

### Change Screenshot Format to JPG
```json
{
  "screenshot": {
    "file_format": "jpg"
  }
}
```

### Disable Clipboard Copy for Screenshots
```json
{
  "screenshot": {
    "copy_to_clipboard": false
  }
}
```

### Custom External Tool Paths
```json
{
  "external_tools": {
    "grim": "/usr/local/bin/grim",
    "fuzzel": "/usr/local/bin/fuzzel"
  }
}
```

### Change Default Color Scheme
```json
{
  "scheme": {
    "default": "catppuccin"
  }
}
```

### Disable Specific Theme Applications
```json
{
  "theme": {
    "enableDiscord": false,
    "enableSpicetify": false
  }
}
```

## Full Documentation
See [CONFIGURATION.md](CONFIGURATION.md) for complete documentation of all options.