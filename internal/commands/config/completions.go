package config

import (
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/spf13/cobra"
)

// RegisterCompletions registers shell completions for config commands
func RegisterCompletions(cmd *cobra.Command) {
	// Register completions for the list command's --category flag
	if listCmd, _, err := cmd.Find([]string{"list"}); err == nil {
		listCmd.RegisterFlagCompletionFunc("category", categoryCompletion)
		listCmd.RegisterFlagCompletionFunc("type", typeCompletion)
		listCmd.RegisterFlagCompletionFunc("copy", configPathCompletion)
	}

	// Register completions for the describe command
	if describeCmd, _, err := cmd.Find([]string{"describe"}); err == nil {
		describeCmd.ValidArgsFunction = configPathCompletionFunc
	}

	// Register completions for get/set commands
	if getCmd, _, err := cmd.Find([]string{"get"}); err == nil {
		getCmd.ValidArgsFunction = domainAndPathCompletionFunc
	}

	if setCmd, _, err := cmd.Find([]string{"set"}); err == nil {
		setCmd.ValidArgsFunction = domainAndPathCompletionFunc
	}

	// Register completions for search command
	if searchCmd, _, err := cmd.Find([]string{"search"}); err == nil {
		searchCmd.ValidArgsFunction = searchSuggestionFunc
	}
}

// categoryCompletion provides completion for category names
func categoryCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	categories := []string{
		"theme",
		"scheme",
		"wallpaper",
		"discord",
		"pip",
		"idle",
		"notification",
		"paths",
	}

	var completions []string
	for _, cat := range categories {
		if strings.HasPrefix(cat, toComplete) {
			completions = append(completions, cat)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// typeCompletion provides completion for field types
func typeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	types := []string{
		"bool",
		"string",
		"int",
		"float",
		"[]string",
		"map",
		"object",
	}

	var completions []string
	for _, t := range types {
		if strings.HasPrefix(t, toComplete) {
			completions = append(completions, t)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// configPathCompletion provides completion for config paths
func configPathCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Initialize metadata registry if needed
	if err := config.InitializeRegistry(); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Get all field paths
	fields := config.MetadataRegistry.GetAllFields()
	var paths []string
	for path := range fields {
		paths = append(paths, path)
	}

	// Filter based on what's being typed
	var completions []string
	for _, path := range paths {
		if strings.HasPrefix(path, toComplete) {
			completions = append(completions, path)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// configPathCompletionFunc provides completion for config paths as ValidArgsFunction
func configPathCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return configPathCompletion(cmd, args, toComplete)
}

// domainAndPathCompletionFunc provides completion for domain and path arguments
func domainAndPathCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// First argument is domain
	if len(args) == 0 {
		domains := []string{"cli", "shell", "all"}
		var completions []string
		for _, d := range domains {
			if strings.HasPrefix(d, toComplete) {
				completions = append(completions, d)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// Second argument is path (only for cli domain for now)
	if len(args) == 1 && args[0] == "cli" {
		return configPathCompletion(cmd, args, toComplete)
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

// searchSuggestionFunc provides search suggestions based on common terms
func searchSuggestionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	suggestions := []string{
		"theme",
		"scheme",
		"wallpaper",
		"gtk",
		"qt",
		"discord",
		"enable",
		"disable",
		"default",
		"material",
		"notification",
		"pip",
		"idle",
		"path",
		"directory",
	}

	var completions []string
	for _, s := range suggestions {
		if strings.HasPrefix(s, toComplete) {
			completions = append(completions, s)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// GenerateCompletionScript generates shell completion script for the specified shell
func GenerateCompletionScript(rootCmd *cobra.Command, shell string) error {
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletionV2(rootCmd.OutOrStdout(), true)
	case "zsh":
		return rootCmd.GenZshCompletion(rootCmd.OutOrStdout())
	case "fish":
		return rootCmd.GenFishCompletion(rootCmd.OutOrStdout(), true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(rootCmd.OutOrStdout())
	default:
		return rootCmd.GenBashCompletionV2(rootCmd.OutOrStdout(), true)
	}
}
