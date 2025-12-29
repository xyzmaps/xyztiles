package cmd

import (
	"fmt"
	"os"

	"com.github.olifink.go-cli-template/src/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-cli-template",
	Short: "go-cli-template - A CLI tool",
	Long:  `go-cli-template is a CLI application for XYZ Maps`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("go-cli-template")
		fmt.Println("Run 'go-cli-template --help' for usage information")
	},
}

var versionFlag bool

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")

	// Handle version flag
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Println(version.GetFullVersion())
			os.Exit(0)
		}
	}
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
