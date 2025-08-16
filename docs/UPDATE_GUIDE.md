# Heimdall Update Guide

## Overview

Heimdall includes a built-in self-update mechanism that allows you to easily update to the latest version directly from the command line. The update system supports multiple release channels, automatic update checks, and safe rollback capabilities.

## Quick Start

### Check for Updates
```bash
# Check if an update is available
heimdall update --check
```

### Update to Latest Version
```bash
# Update to the latest stable release
heimdall update

# Update to a specific channel
heimdall update --channel beta
```

### Rollback to Previous Version
```bash
# Rollback to the previous version if something goes wrong
heimdall update --rollback
```

## Features

### 1. Multi-Channel Support

Heimdall supports multiple release channels:
- **stable**: Production-ready releases (default)
- **beta**: Pre-release versions for testing new features
- **nightly**: Development builds with the latest changes

Switch channels using:
```bash
heimdall update --channel beta
```

### 2. Automatic Update Checks

Heimdall can automatically check for updates in the background and notify you when a new version is available.

Configure automatic checks:
```bash
# Enable/disable automatic checks
heimdall config set update.check_enabled true

# Set check frequency (hourly, daily, weekly, monthly)
heimdall config set update.check_frequency daily

# Enable/disable notifications
heimdall config set update.notify_on_available true
```

### 3. Git Repository Support

If you're running Heimdall from a git repository (e.g., for development), the update command will:
1. Perform a `git pull` to fetch the latest changes
2. Automatically rebuild the binary
3. Preserve your local modifications

### 4. Binary Updates

For standard installations, Heimdall will:
1. Download the appropriate binary for your platform
2. Verify the checksum (if available)
3. Create a backup of the current version
4. Atomically replace the binary
5. Clean up old backups (keeping the last 3)

### 5. Safe Rollback

Every update creates a backup of the previous version. If something goes wrong:
```bash
# Rollback to the previous version
heimdall update --rollback
```

## Command Options

```bash
heimdall update [flags]
```

### Flags
- `--check`: Check for updates without installing
- `--force`: Force update even if already on latest version
- `--channel <channel>`: Release channel (stable, beta, nightly)
- `--rollback`: Rollback to previous version
- `--verbose`: Show detailed output during update

## Configuration

Update settings are stored in `~/.config/heimdall/update.json`:

```json
{
  "check_enabled": true,
  "check_frequency": "daily",
  "last_check": "2024-01-16T10:30:00Z",
  "channel": "stable",
  "notify_on_available": true
}
```

## Platform Support

The update mechanism supports:
- **Linux** (x86_64, arm64)
- **macOS** (x86_64, arm64)
- **Windows** (x86_64)

Platform-specific binaries are automatically selected based on your system.

## Security

### Checksum Verification
When available, Heimdall verifies SHA256 checksums of downloaded binaries to ensure integrity.

### HTTPS Only
All downloads use HTTPS to prevent man-in-the-middle attacks.

### GitHub Token Support
For increased rate limits, you can provide a GitHub token:
```bash
export GITHUB_TOKEN=your_token_here
heimdall update
```

## Troubleshooting

### Permission Denied
If you get permission errors during update:
```bash
# Use sudo if heimdall is installed system-wide
sudo heimdall update
```

### Update Fails
If an update fails:
1. Check your internet connection
2. Verify GitHub is accessible
3. Try again with verbose output: `heimdall update --verbose`
4. If needed, rollback: `heimdall update --rollback`

### Git Repository Issues
For git-based installations:
1. Ensure you have no uncommitted changes
2. Check that your branch tracks a remote
3. Verify you have push/pull permissions

### Manual Update
If automatic update fails, you can always update manually:
```bash
# For git installations
git pull
make build

# For binary installations
# Download from: https://github.com/arthur404dev/heimdall-cli/releases
```

## Environment Variables

- `GITHUB_TOKEN`: GitHub personal access token for API requests
- `HEIMDALL_UPDATE_CHANNEL`: Default update channel
- `HEIMDALL_NO_UPDATE_CHECK`: Disable automatic update checks

## Examples

### Check and Update if Available
```bash
if heimdall update --check | grep -q "Update available"; then
    heimdall update
fi
```

### Update to Beta Channel
```bash
heimdall update --channel beta --verbose
```

### Force Reinstall Current Version
```bash
heimdall update --force
```

### Automated Update Script
```bash
#!/bin/bash
# Auto-update heimdall if needed

heimdall update --check > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "Updating heimdall..."
    heimdall update
fi
```

## Best Practices

1. **Regular Updates**: Keep Heimdall updated for the latest features and bug fixes
2. **Test Beta Releases**: Help test new features by using the beta channel
3. **Backup Important Configs**: Always backup your configuration before major updates
4. **Monitor Release Notes**: Check release notes for breaking changes
5. **Use Automation**: Set up automatic update checks for convenience

## Related Commands

- `heimdall version`: Show current version information
- `heimdall config`: Configure update settings
- `heimdall completion`: Generate shell completions

## Support

For issues or questions about updates:
1. Check the [GitHub Issues](https://github.com/arthur404dev/heimdall-cli/issues)
2. Review the [Release Notes](https://github.com/arthur404dev/heimdall-cli/releases)
3. Join the community discussions