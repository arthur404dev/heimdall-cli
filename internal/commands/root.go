package commands

import (
	"fmt"
	"os"

	"github.com/heimdall-cli/heimdall/internal/commands/clipboard"
	"github.com/heimdall-cli/heimdall/internal/commands/emoji"
	"github.com/heimdall-cli/heimdall/internal/commands/pip"
	"github.com/heimdall-cli/heimdall/internal/commands/record"
	"github.com/heimdall-cli/heimdall/internal/commands/scheme"
	"github.com/heimdall-cli/heimdall/internal/commands/screenshot"
	"github.com/heimdall-cli/heimdall/internal/commands/shell"
	"github.com/heimdall-cli/heimdall/internal/commands/toggle"
	"github.com/heimdall-cli/heimdall/internal/commands/wallpaper"
	"github.com/heimdall-cli/heimdall/internal/utils/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	debug   bool

	// Version information (set via ldflags)
	Version = "dev"
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/heimdall/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Set version directly
	rootCmd.Version = fmt.Sprintf("%s\nBuilt:   %s\nCommit:  %s\nBuilt by: %s",
		Version, Date, Commit, BuiltBy)

	// Add commands
	addCommands()
}

// addCommands adds all subcommands to the root command
func addCommands() {
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
		viper.SetConfigType("yaml")
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
