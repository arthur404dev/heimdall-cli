# User-Defined Color Schemes Guide

## Overview

Heimdall-cli now supports user-defined color schemes, allowing you to create and use custom themes without modifying the source code. This guide explains how to create, install, and manage your own color schemes.

## Quick Start

1. Create a directory for your custom schemes:
```bash
mkdir -p ~/.config/heimdall/schemes
```

2. Add your scheme following the required structure:
```bash
~/.config/heimdall/schemes/
└── my-theme/
    └── default/
        ├── dark.json
        └── light.json (optional)
```

3. Use your scheme:
```bash
heimdall scheme set my-theme
```

## Scheme Structure

### Directory Layout

Each scheme must follow this directory structure:
```
scheme-name/
├── variant1/
│   ├── dark.json
│   └── light.json
└── variant2/
    ├── dark.json
    └── light.json
```

- **scheme-name**: The name of your color scheme (e.g., "my-custom-theme")
- **variant**: Different flavors of your scheme (e.g., "default", "vibrant", "pastel")
- **mode files**: `dark.json` and/or `light.json` containing the color definitions

### JSON Format

Each color scheme file must be a valid JSON file with the following structure:

```json
{
  "name": "My Custom Theme",
  "author": "Your Name",
  "variant": "default",
  "colours": {
    "base00": "#1a1a2e",
    "base01": "#16213e",
    "base02": "#0f3460",
    "base03": "#53354a",
    "base04": "#903749",
    "base05": "#e84545",
    "base06": "#f5f5f5",
    "base07": "#ffffff",
    "base08": "#ff6b6b",
    "base09": "#ffa500",
    "base0A": "#ffd700",
    "base0B": "#4caf50",
    "base0C": "#00bcd4",
    "base0D": "#2196f3",
    "base0E": "#9c27b0",
    "base0F": "#e91e63"
  }
}
```

### Required Color Keys

All 16 base colors are required for a valid scheme:

| Key | Purpose | Common Use |
|-----|---------|------------|
| `base00` | Background | Default background |
| `base01` | Lighter Background | Status bars, line numbers |
| `base02` | Selection Background | Visual selection, search highlights |
| `base03` | Comments | Comments, invisibles, line highlighting |
| `base04` | Dark Foreground | Status bar text |
| `base05` | Default Foreground | Normal text, caret |
| `base06` | Light Foreground | Light text (rarely used) |
| `base07` | Light Background | Light background (rarely used) |
| `base08` | Red | Variables, errors, diff deleted |
| `base09` | Orange | Integers, booleans, constants |
| `base0A` | Yellow | Classes, search background |
| `base0B` | Green | Strings, diff inserted |
| `base0C` | Cyan | Support, regex, escape chars |
| `base0D` | Blue | Functions, methods, headings |
| `base0E` | Purple | Keywords, tags, diff changed |
| `base0F` | Brown/Pink | Deprecated, special purposes |

## Configuration

### Setting User Scheme Paths

By default, heimdall looks for user schemes in `~/.config/heimdall/schemes/`. You can configure additional paths:

#### Via Configuration File

Edit `~/.config/heimdall/config.json`:
```json
{
  "scheme": {
    "user_paths": [
      "~/.config/heimdall/schemes",
      "~/.local/share/themes",
      "/usr/share/heimdall/schemes"
    ]
  }
}
```

#### Via Environment Variable

Set `HEIMDALL_SCHEME_PATHS` to override config paths:
```bash
export HEIMDALL_SCHEME_PATHS="~/my-schemes:/usr/share/custom-schemes"
```

## Commands

### List All Schemes

Show all available schemes (bundled and user):
```bash
heimdall scheme list
```

Filter by source:
```bash
heimdall scheme list --source user      # Only user schemes
heimdall scheme list --source bundled   # Only bundled schemes
heimdall scheme list --source generated # Only generated schemes
```

### Get Scheme Information

View details about a specific scheme:
```bash
heimdall scheme get my-theme
```

### Set Active Scheme

Apply a user scheme:
```bash
heimdall scheme set my-theme           # Uses default variant
heimdall scheme set my-theme vibrant   # Uses specific variant
```

### Install Bundled Scheme to User Directory

Copy a bundled scheme for customization:
```bash
heimdall scheme install catppuccin --user
```

## Creating Your Own Scheme

### Method 1: From Scratch

1. Create the directory structure:
```bash
mkdir -p ~/.config/heimdall/schemes/my-awesome-theme/default
```

2. Create `dark.json`:
```bash
cat > ~/.config/heimdall/schemes/my-awesome-theme/default/dark.json << 'EOF'
{
  "name": "My Awesome Theme",
  "author": "Your Name",
  "colours": {
    "base00": "#1a1a1a",
    "base01": "#2a2a2a",
    "base02": "#3a3a3a",
    "base03": "#4a4a4a",
    "base04": "#5a5a5a",
    "base05": "#dadada",
    "base06": "#eaeaea",
    "base07": "#fafafa",
    "base08": "#ff5555",
    "base09": "#ffb86c",
    "base0A": "#f1fa8c",
    "base0B": "#50fa7b",
    "base0C": "#8be9fd",
    "base0D": "#6272a4",
    "base0E": "#bd93f9",
    "base0F": "#ff79c6"
  }
}
EOF
```

3. Test your scheme:
```bash
heimdall scheme set my-awesome-theme
```

### Method 2: Modify Existing Scheme

1. Install a bundled scheme to user directory:
```bash
heimdall scheme install gruvbox --user
```

2. Edit the colors:
```bash
$EDITOR ~/.config/heimdall/schemes/gruvbox/medium/dark.json
```

3. Apply your modified version:
```bash
heimdall scheme set gruvbox
```

### Method 3: Generate from Wallpaper

Generate a scheme from your wallpaper:
```bash
heimdall wallpaper set /path/to/wallpaper.jpg
# This creates a 'generated' scheme
heimdall scheme set generated
```

## Validation

Heimdall validates user schemes when loading them. Common validation errors:

### Missing Required Colors
```
Error: missing required color keys: base0C, base0D
```
**Solution**: Add the missing color definitions to your JSON file.

### Invalid Hex Color Format
```
Error: invalid hex color format: not-a-color
```
**Solution**: Use valid hex color codes (#RGB, #RRGGBB, or #RRGGBBAA).

### Invalid JSON Syntax
```
Error: invalid JSON syntax at position 245
```
**Solution**: Check for missing commas, quotes, or brackets in your JSON.

## Migration from Old Formats

If you have schemes in the old text format, heimdall can automatically convert them:

### Old Text Format Example
```
base00=#1a1a1a
base01=#2a2a2a
base02=#3a3a3a
...
```

This will be automatically converted to the new JSON format when loaded.

## Tips and Best Practices

1. **Start with a Template**: Use the example scheme in `docs/examples/oldworld/default/dark.json` as a starting point.

2. **Test Incrementally**: After creating your scheme, test it with different applications to ensure colors work well.

3. **Use Meaningful Names**: Name your schemes and variants descriptively (e.g., "ocean-breeze/pastel" instead of "theme1/var1").

4. **Maintain Contrast**: Ensure sufficient contrast between background (base00) and foreground (base05) colors for readability.

5. **Consider Both Modes**: If possible, create both dark and light variants for maximum flexibility.

6. **Version Control**: Keep your custom schemes in a git repository for backup and sharing.

## Troubleshooting

### Scheme Not Found
- Check that the directory structure is correct
- Verify the scheme is in one of the configured user paths
- Run `heimdall scheme list` to see all available schemes

### Colors Not Applying
- Ensure all required color keys are present
- Check for typos in color values
- Verify JSON syntax is valid

### Precedence Issues
- User schemes override bundled schemes with the same name
- First user path in configuration takes precedence over later paths
- Environment variable paths override configuration file paths

## Sharing Schemes

To share your scheme with others:

1. Create a repository with your scheme files
2. Users can clone and copy to their schemes directory:
```bash
git clone https://github.com/username/my-schemes
cp -r my-schemes/* ~/.config/heimdall/schemes/
```

3. Consider submitting popular schemes as pull requests to be included as bundled schemes

## Example Schemes

See `docs/examples/oldworld/default/dark.json` for a complete example scheme you can use as a template.

## Support

For issues or questions about user-defined schemes:
- Check the validation errors for specific problems
- Review this guide for proper structure and format
- Open an issue on the heimdall-cli repository with details
