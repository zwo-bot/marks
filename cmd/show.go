package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/zwo-bot/marks/bookmark"
	"github.com/zwo-bot/marks/db"
	"github.com/zwo-bot/marks/internal/logger"
	"github.com/zwo-bot/marks/plugins"
)

var (
	outputFormat     string
	showDeduplicate bool
	showCmd         = &cobra.Command{
		Use:   "show",
		Short: "Show bookmarks",
		Long:  `Show bookmarks from all configured browsers in various formats.`,
		Run:   showBookmarks,
	}

	// Command to list plugins
	listPluginsCmd = &cobra.Command{
		Use:   "list-plugins",
		Short: "List available plugins",
		Long:  `List available browser plugins that can provide bookmarks`,
		Run:   listPlugins,
	}
)

func init() {
	showCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (text|json)")
	showCmd.Flags().BoolVarP(&showDeduplicate, "deduplicate", "d", true, "Remove duplicate bookmarks")
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(listPluginsCmd)
}

func showBookmarks(cmd *cobra.Command, args []string) {
	log := logger.GetLogger()

	// First get bookmarks from DB for fast response
	bookmarks, err := db.GetBookmarks()
	if err != nil {
		log.Error("Error getting bookmarks from database", "error", err)
		// If we can't get from DB, try getting directly from plugins
		p := plugins.Init()
		bookmarks = p.GetBookmarks()
	} else if len(bookmarks) == 0 {
		// If DB is empty, get from plugins
		log.Debug("No bookmarks in database, getting from plugins")
		p := plugins.Init()
		bookmarks = p.GetBookmarks()
		// Save to DB for next time
		if err := db.UpdateBookmarks(bookmarks); err != nil {
			log.Error("Error saving initial bookmarks to database", "error", err)
		}
	}

	// Deduplicate if requested
	if showDeduplicate {
		bookmarks = bookmarks.RemoveDuplicates()
	}

	// Output bookmarks in the requested format
	switch outputFormat {
	case "json":
		if err := outputJSON(bookmarks); err != nil {
			log.Error("Error outputting JSON", "error", err)
		}
	case "text":
		outputText(bookmarks)
	default:
		log.Debug("Unknown format, using text", "format", outputFormat)
		outputText(bookmarks)
	}

	// Spawn the update command as a separate process
	args = []string{"update"}
	
	// Pass config path if it was specified
	if rootOptions.configPath != "" {
		args = append(args, "--config", rootOptions.configPath)
	}
	
	// Pass log level to maintain consistent logging
	args = append(args, "--log-level", rootOptions.logLevel)

	updateCmd := exec.Command(os.Args[0], args...)
	// Start the command without waiting for it to complete
	if err := updateCmd.Start(); err != nil {
		log.Debug("Error starting update process", "error", err)
	}
}

func outputJSON(bookmarks bookmark.Bookmarks) error {
	data, err := json.MarshalIndent(bookmarks, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputText(bookmarks bookmark.Bookmarks) {
	for _, bookmark := range bookmarks {
		fmt.Printf("%s\t%s\n", bookmark.Title, bookmark.URI)
	}
}

func listPlugins(cmd *cobra.Command, args []string) {
	p := plugins.Init()
	for _, plugin := range p.ListPlugins() {
		fmt.Println(plugin)
	}
}
