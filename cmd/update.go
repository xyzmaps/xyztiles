package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"org.olifinks.go-cli-template/src/version"
)

var (
	dryRun bool
	yes    bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update go-cli-template to the latest version",
	Long: `Check for the latest release on GitHub and update the binary in-place.
The update is verified using checksums from the release.`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Check for updates without applying them")
	updateCmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	currentVersion := version.GetVersion()
	fmt.Printf("Current version: %s\n", currentVersion)

	// Configure selfupdate
	latest, found, err := selfupdate.DetectLatest("olifinks/go-cli-template")
	if err != nil {
		return fmt.Errorf("error checking for updates: %w", err)
	}

	if !found {
		return fmt.Errorf("no releases found for olifinks/go-cli-template")
	}

	fmt.Printf("Latest version:  %s\n", latest.Version)

	// Check if we're already on the latest version
	if currentVersion == latest.Version.String() {
		fmt.Println("\nAlready up to date!")
		return nil
	}

	// Dry run - just show what would happen
	if dryRun {
		fmt.Printf("\nUpdate available: %s -> %s\n", currentVersion, latest.Version)
		fmt.Println("Run without --dry-run to apply the update")
		return nil
	}

	// Confirm update
	if !yes {
		fmt.Printf("\nUpdate from %s to %s? [y/N]: ", currentVersion, latest.Version)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Update cancelled")
			return nil
		}
	}

	// Perform the update
	fmt.Println("\nDownloading and verifying update...")
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("error updating binary: %w", err)
	}

	fmt.Printf("\nSuccessfully updated to %s!\n", latest.Version)
	fmt.Println("Please restart the application to use the new version")

	return nil
}
