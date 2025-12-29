package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"org.xyzmaps.xyztiles/src/resources"
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

		// Create server configuration
		cfg := server.Config{
			Port: port,
		}

		// Use embedded image or custom image path
		if imagePath == "" {
			// Use embedded image
			if !resources.HasEmbeddedMap() {
				log.Fatal("Error: No embedded map available and --image flag not provided.")
			}
			log.Printf("Using embedded world map (%d bytes)", resources.DefaultMapSize())
			cfg.EmbeddedData = resources.DefaultWorldMap
		} else {
			// Use custom image from file
			if _, err := os.Stat(imagePath); os.IsNotExist(err) {
				log.Fatalf("Error: Image file not found at %s", imagePath)
			}
			cfg.ImagePath = imagePath
		}

		// Create and start the server
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
	rootCmd.Flags().StringVarP(&imagePath, "image", "i", "", "Path to custom equirectangular world map image (optional, uses embedded map if not specified)")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
