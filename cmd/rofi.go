package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zwo-bot/marks/db"
	"github.com/zwo-bot/marks/internal/logger"
	"github.com/zwo-bot/marks/plugins"
)

var (
	rofiDeduplicate bool
	rofiCmd         = &cobra.Command{
		Use:   "rofi",
		Short: "Show bookmarks in rofi format",
		Long:  `Show bookmarks in rofi format and handle rofi interactions.`,
		Run:   showRofiBookmarks,
	}
)

func init() {
	rofiCmd.Flags().BoolVarP(&rofiDeduplicate, "deduplicate", "d", true, "Remove duplicate bookmarks")
	rootCmd.AddCommand(rofiCmd)
}

func showRofiBookmarks(cmd *cobra.Command, args []string) {
	log := logger.GetLogger()

	// Check if rofi has selected an item
	if os.Getenv("ROFI_RETV") == "1" {
		// Get the selected URL from ROFI_INFO
		url := os.Getenv("ROFI_INFO")
		if url != "" {
			// Open URL in default browser using xdg-open
			openCmd := exec.Command("xdg-open", url)
			openCmd.Start()
			return
		}
	}

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

	// Enable markup parsing
	fmt.Println("\x00markup-rows\x1ftrue")

	// Set custom prompt
	fmt.Println("\x00prompt\x1f ")

		// Deduplicate if requested
	if rofiDeduplicate {
		bookmarks = bookmarks.RemoveDuplicates()
	}

	// Output bookmarks in rofi format
	for _, bookmark := range bookmarks {
			// Create display text with Pango markup
			title := bookmark.Title
			if title == "" {
				title = bookmark.Path
			}
			
			// Escape Pango markup characters in title and URL
			title = strings.ReplaceAll(title, "&", "&amp;")
			title = strings.ReplaceAll(title, "<", "&lt;")
			title = strings.ReplaceAll(title, ">", "&gt;")
			url := strings.ReplaceAll(bookmark.URI, "&", "&amp;")
			url = strings.ReplaceAll(url, "<", "&lt;")
			url = strings.ReplaceAll(url, ">", "&gt;")
			
			// Build single line display with title and URL
			displayText := fmt.Sprintf("<b>%s</b>  <span color='#888888' alpha='70%%'>%s</span>", title, url)

			// Build the line in rofi format with the multi-line display text
			line := displayText + "\x00info\x1f" + bookmark.URI
			
			// Add icon if available
			if bookmark.Icon != "" {
				log.Debug("Adding icon to rofi output", 
					"title", bookmark.Title,
					"path", bookmark.Path,
					"icon_path", bookmark.Icon)
				line += "\x1ficon\x1f" + bookmark.Icon
			} else {
				log.Debug("No icon available for bookmark",
					"title", bookmark.Title,
					"path", bookmark.Path,
					"uri", bookmark.URI)
			}
			
			// Add meta field
			line += "\x1fmeta\x1f" + bookmark.URI

			fmt.Println(line)
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
