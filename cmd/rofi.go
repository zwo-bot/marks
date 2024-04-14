package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
	"github.com/zwo-bot/go-rofi-bookmarks/plugins"
	"github.com/zwo-bot/go-rofi-bookmarks/db"
)

var rofiCmd = &cobra.Command{
	Use:   "rofi",
	Short: "Test",
	Long:  `Provide temporary STS Credentials in ~/.aws/credentials to default or given profile`,
	Run:   showBookmarks,
}

// add command to list plugins
var listPluginsCmd = &cobra.Command{
	Use:   "list-plugins",
	Short: "List available plugins",
	Long:  `List available plugins`,
	Run:   listPlugins,
}

func init() {
	rootCmd.AddCommand(rofiCmd)
	rootCmd.AddCommand(listPluginsCmd)
}

func showBookmarks(cmd *cobra.Command, args []string) {

	var bookmarks bookmark.Bookmarks
	p := plugins.Init()
	bookmarks = p.GetBookmarks()

	for _, bookmark := range bookmarks {
		fmt.Println(bookmark.Title)
	}

	db.ConnectDatabase()
}

func listPlugins(cmd *cobra.Command, args []string) {
	plugins := plugins.Init()
	p := plugins.ListPlugins()
	for _, plugin := range p {
		fmt.Println(plugin)
	}
}
