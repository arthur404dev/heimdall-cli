package commands

import (
	"fmt"

	"github.com/arthur404dev/heimdall-cli/internal/utils/color"
	"github.com/arthur404dev/heimdall-cli/internal/utils/hypr"
	"github.com/arthur404dev/heimdall-cli/internal/utils/notify"
	"github.com/spf13/cobra"
)

// testCmd represents the test command (hidden, for development)
var testCmd = &cobra.Command{
	Use:    "test",
	Short:  "Test Phase 2 utilities",
	Hidden: true, // Hide from normal help
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Testing Phase 2 utilities...")

		// Test color utilities
		fmt.Println("\n=== Color Utilities ===")
		c, err := color.NewFromHex("#FF6B6B")
		if err != nil {
			fmt.Printf("Error creating color: %v\n", err)
		} else {
			fmt.Printf("Color: %s\n", c.Hex)
			fmt.Printf("RGB: R=%d, G=%d, B=%d\n", c.RGB.R, c.RGB.G, c.RGB.B)
			fmt.Printf("HSL: H=%.1f, S=%.1f, L=%.1f\n", c.HSL.H, c.HSL.S, c.HSL.L)
			fmt.Printf("Is Dark: %v\n", c.IsDark())

			lighter := c.Lighten(10)
			fmt.Printf("Lighter: %s\n", lighter.Hex)
		}

		// Test Hyprland IPC
		fmt.Println("\n=== Hyprland IPC ===")
		if hypr.IsRunning() {
			client, err := hypr.NewClient()
			if err != nil {
				fmt.Printf("Error creating Hyprland client: %v\n", err)
			} else {
				version, err := client.GetVersion()
				if err != nil {
					fmt.Printf("Error getting version: %v\n", err)
				} else {
					fmt.Printf("Hyprland version: %s\n", version)
				}

				workspaces, err := client.GetWorkspaces()
				if err != nil {
					fmt.Printf("Error getting workspaces: %v\n", err)
				} else {
					fmt.Printf("Found %d workspaces\n", len(workspaces))
				}
			}
		} else {
			fmt.Println("Hyprland is not running")
		}

		// Test notifications
		fmt.Println("\n=== Notifications ===")
		if notify.IsAvailable() {
			err := notify.Send("Heimdall Test", "Phase 2 utilities are working!")
			if err != nil {
				fmt.Printf("Error sending notification: %v\n", err)
			} else {
				fmt.Println("Notification sent successfully")
			}
		} else {
			fmt.Println("Notification system not available")
		}

		fmt.Println("\n✅ Phase 2 utilities test complete")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
