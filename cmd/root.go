package cmd

import (

	"github.com/spf13/cobra"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/logger"
)

var rootOptions struct {
	logFormat  string
	logLevel   string
	configFile string
}

var rootCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "Bookmark manager for rofi",
	Long:  `A simple bookmark manager for rofi`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&rootOptions.logLevel,
		"log-level", "l",
		"warn",
		"Choose log level [error,warn,info,debug]",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootOptions.logFormat,
		"log-format",
		"txt",
		"log format [json|text].",
	)

	cobra.OnInitialize( func() {
		logger.Initialize(rootOptions.logLevel)
	})
}
