# Kitty Auto-Reload Feature

## Overview
The heimdall-cli now automatically reloads all running kitty terminal instances when you change the color scheme using `heimdall scheme set`. This means you no longer need to manually press `Ctrl+Shift+F5` to apply the new colors.

## How It Works

The implementation uses two methods to reload kitty:

### Method 1: Remote Control (Preferred)
- Uses `kitten @ load-config` command to reload configuration
- Works best when `allow_remote_control` is enabled in kitty.conf
- Provides the cleanest reload experience

### Method 2: Signal-based Reload (Fallback)
- Sends `SIGUSR1` signal to all kitty processes
- Works even without remote control enabled
- Kitty automatically reloads its configuration when receiving this signal

## Requirements

### No Special Requirements for Basic Functionality
- The signal-based method works out of the box on most Linux systems
- No sudo/root privileges required for sending signals to your own processes

### For Optimal Experience
To enable the remote control method (which is slightly faster and more reliable):

1. Add to your `~/.config/kitty/kitty.conf`:
   ```
   allow_remote_control yes
   ```

2. Or start kitty with:
   ```bash
   kitty -o allow_remote_control=yes
   ```

## Usage

Simply use the scheme set command as usual:

```bash
heimdall scheme set rosepine
heimdall scheme set catppuccin mocha dark
heimdall scheme set --apps kitty gruvbox
```

When kitty is included in the themed applications (either explicitly or through configuration), all running kitty instances will automatically reload with the new colors.

## Troubleshooting

### Kitty doesn't reload automatically
1. Check if kitty processes are running: `pgrep -x kitty`
2. Verify you have permission to signal the processes (they should be owned by your user)
3. Enable `allow_remote_control` in kitty.conf for better reliability

### Some kitty instances don't reload
- If kitty instances are running under different users, only instances owned by your user will reload
- Kitty instances started with `--config` pointing to a different config file may not pick up the changes

## Security Considerations

- No sudo/root privileges are required or used
- The implementation only signals processes owned by the current user
- Remote control (if enabled) is limited to the local machine
- No network access is required or used

## Technical Details

The implementation is in `/internal/theme/kitty_reload.go` and includes:
- Automatic detection of available reload methods
- Graceful fallback from remote control to signal-based reload
- Timeout handling to prevent hanging
- Silent failure if no kitty instances are running (not an error condition)