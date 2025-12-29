package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"org.xyzmaps.xyztiles/src/server"
	"org.xyzmaps.xyztiles/src/version"
)

var (
	versionFlag bool
	port        int
	imagePath   string
)

var rootCmd = &cobra.Command{
	Use:   "xyztiles",
	Short: "xyztiles - Embedded World Map Tile Server",
	Long: `xyztiles is a single Go binary that serves web map tiles from an equirectangular world map image.
Zero external dependencies, no hosted services required - perfect for learning web mapping or offline/air-gapped environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle version flag
		if versionFlag {
			fmt.Println(version.GetFullVersion())
			os.Exit(0)
		}

		// Check if image path is provided
		if imagePath == "" {
			log.Fatal("Error: --image flag is required. Please provide a path to an equirectangular world map image.")
		}

		// Check if image file exists
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			log.Fatalf("Error: Image file not found at %s", imagePath)
		}

		// Create and start the server
		cfg := server.Config{
			Port:      port,
			ImagePath: imagePath,
		}

		srv, err := server.New(cfg)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	rootCmd.Flags().StringVarP(&imagePath, "image", "i", "", "Path to equirectangular world map image (JPEG)")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
