package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zwo-bot/marks/db"
	"github.com/zwo-bot/marks/internal/logger"
	"github.com/zwo-bot/marks/plugins"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update bookmarks database",
	Long:  `Update bookmarks database with fresh data from configured browsers.`,
	Run:   updateBookmarks,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateBookmarks(cmd *cobra.Command, args []string) {
	log := logger.GetLogger()

	// Get fresh bookmarks from plugins
	p := plugins.Init()
	bookmarks := p.GetBookmarks()

	log.Debug("Got bookmarks from plugins", "count", len(bookmarks))

	// Update the database
	if err := db.UpdateBookmarks(bookmarks); err != nil {
		log.Error("Error updating bookmarks in database", "error", err)
	} else {
		log.Debug("Updated bookmarks in database", "count", len(bookmarks))
	}
}
