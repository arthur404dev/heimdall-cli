# Minimal Configuration Guide

As of version 0.2.0, heimdall-cli supports minimal configuration files. You no longer need to maintain a complete configuration file with all settings - the system will automatically use sensible defaults for any settings you don't specify.

## How It Works

1. **No Config Required**: The system works perfectly with no `config.json` file at all. All defaults are applied automatically.

2. **Partial Configs**: You can create a minimal config file containing only the settings you want to customize. Everything else uses defaults.

3. **Runtime Merging**: Defaults are merged with your config at runtime, not saved to your config file. This keeps your config clean and minimal.

## Examples

### No Configuration File
Simply run heimdall without any config file:
```bash
heimdall scheme list
heimdall scheme apply catppuccin-mocha
```

### Minimal Configuration
Create a `~/.config/heimdall/config.json` with only what you want to change:

```json
{
  "scheme": {
    "default": "catppuccin-mocha"
  }
}
```

### Disable Specific Features
```json
{
  "theme": {
    "enableQt": false,
    "enableSpicetify": false
  }
}
```

### Custom Paths Only
```json
{
  "wallpaper": {
    "directory": "~/Pictures/MyWallpapers"
  },
  "screenshot": {
    "directory": "~/Screenshots"
  }
}
```

## New Config Commands

### View Effective Configuration
See the complete merged configuration (your settings + defaults):
```bash
heimdall config effective
```

### View User Configuration
See only your custom settings (what's in your config file):
```bash
heimdall config user
```

### Validate Configuration
Check if your configuration is valid:
```bash
heimdall config validate
```

## Benefits

1. **Cleaner Configs**: Your config file only contains your customizations
2. **Easier Updates**: New features automatically get sensible defaults
3. **Less Maintenance**: No need to update your config when defaults change
4. **Better Portability**: Minimal configs are easier to share and version control

## Migration from Old Config

If you have an existing complete config file, you can:

1. Back up your current config:
   ```bash
   cp ~/.config/heimdall/config.json ~/.config/heimdall/config.json.backup
   ```

2. Create a new minimal config with only your customizations

3. Test that everything works as expected:
   ```bash
   heimdall config validate
   heimdall config effective
   ```

## Default Values

All default values are defined in the source code and can be viewed:
- Via the `heimdall config defaults` command (coming in Phase 3)
- Via the `heimdall config effective` command when no config exists
- In the source code at `internal/config/config.go`

## Environment Variables

Environment variables still work as overrides:
```bash
HEIMDALL_SCHEME_PATHS="/custom/path1:/custom/path2" heimdall scheme list
```

These override both defaults and config file settings.