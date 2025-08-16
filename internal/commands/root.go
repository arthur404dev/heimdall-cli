package commands

import (
	"fmt"
	"os"

	"github.com/arthur404dev/heimdall-cli/internal/commands/clipboard"
	"github.com/arthur404dev/heimdall-cli/internal/commands/config"
	"github.com/arthur404dev/heimdall-cli/internal/commands/emoji"
	"github.com/arthur404dev/heimdall-cli/internal/commands/idle"
	"github.com/arthur404dev/heimdall-cli/internal/commands/pip"
	"github.com/arthur404dev/heimdall-cli/internal/commands/record"
	"github.com/arthur404dev/heimdall-cli/internal/commands/scheme"
	"github.com/arthur404dev/heimdall-cli/internal/commands/screenshot"
	"github.com/arthur404dev/heimdall-cli/internal/commands/shell"
	"github.com/arthur404dev/heimdall-cli/internal/commands/toggle"
	"github.com/arthur404dev/heimdall-cli/internal/commands/update"
	"github.com/arthur404dev/heimdall-cli/internal/commands/wallpaper"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	debug   bool

	// Version information (set via ldflags)
	Version = "0.2.0"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "heimdall",
	Short: "Main control script for the Heimdall dotfiles",
	Long: `Heimdall is a CLI tool for managing dotfiles, color schemes, 
wallpapers, and system theming. It provides seamless integration with 
Hyprland window manager and supports Material You color generation.

This tool is a Go rewrite of the original Caelestia CLI, offering 
improved performance and a single binary distribution.`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/heimdall/config.json)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Set version directly
	rootCmd.Version = fmt.Sprintf("%s\nBuilt:   %s\nCommit:  %s\nBuilt by: %s",
		Version, Date, Commit, BuiltBy)

	// Add custom version command that works with 'heimdall version'
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("heimdall version %s\n", Version)
			fmt.Printf("Built:   %s\n", Date)
			fmt.Printf("Commit:  %s\n", Commit)
			fmt.Printf("Built by: %s\n", BuiltBy)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Add completion command for generating shell completions
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for heimdall.

To load completions:

Bash:
  $ source <(heimdall completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ heimdall completion bash > /etc/bash_completion.d/heimdall
  # macOS:
  $ heimdall completion bash > $(brew --prefix)/etc/bash_completion.d/heimdall

Zsh:
  $ source <(heimdall completion zsh)
  # To load completions for each session, execute once:
  $ heimdall completion zsh > "${fpath[1]}/_heimdall"

Fish:
  $ heimdall completion fish | source
  # To load completions for each session, execute once:
  $ heimdall completion fish > ~/.config/fish/completions/heimdall.fish

PowerShell:
  PS> heimdall completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> heimdall completion powershell > heimdall.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	rootCmd.AddCommand(completionCmd)

	// Add commands
	addCommands()
}

// addCommands adds all subcommands to the root command
func addCommands() {
	// Add config command (new unified configuration system)
	rootCmd.AddCommand(config.Command())

	// Add shell command
	rootCmd.AddCommand(shell.Command())

	// Add toggle command
	rootCmd.AddCommand(toggle.NewCommand())

	// Add scheme command
	rootCmd.AddCommand(scheme.Command())

	// Add screenshot command
	rootCmd.AddCommand(screenshot.NewCommand())

	// Add record command
	rootCmd.AddCommand(record.NewCommand())

	// Add clipboard command
	rootCmd.AddCommand(clipboard.NewCommand())

	// Add emoji command
	rootCmd.AddCommand(emoji.Command())

	// Add wallpaper command
	rootCmd.AddCommand(wallpaper.Command())

	// Add pip command
	rootCmd.AddCommand(pip.Command())

	// Add idle command
	rootCmd.AddCommand(idle.Command())

	// Add update command
	rootCmd.AddCommand(update.NewUpdateCommand())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set up logging based on flags
	if debug {
		logger.SetDebug(true)
	} else if verbose {
		logger.SetVerbose(true)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".heimdall" (without extension).
		viper.AddConfigPath(home + "/.config/heimdall")
		viper.AddConfigPath(home + "/.config/caelestia") // Backward compatibility
		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("HEIMDALL")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
