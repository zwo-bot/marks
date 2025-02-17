package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zwo-bot/go-rofi-bookmarks/db"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/config"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/logger"
)

var rootOptions struct {
	logFormat  string
	logLevel   string
	configPath string
}

var rootCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "Bookmark manager for rofi",
	Long:  `A simple bookmark manager for rofi`,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Close database connection
		if err := db.CloseDatabase(); err != nil {
			logger.GetLogger().Error("Error closing database", "error", err)
		}
	},
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
		"log format [json|text]",
	)
	rootCmd.PersistentFlags().StringVarP(
		&rootOptions.configPath,
		"config", "c",
		"",
		"Path to config file",
	)

	cobra.OnInitialize(func() {
		// Initialize logger first
		logger.Initialize(rootOptions.logLevel)
		log := logger.GetLogger()

		// Initialize config
		if rootOptions.configPath != "" {
			absPath, err := filepath.Abs(rootOptions.configPath)
			if err != nil {
				log.Error("Error getting absolute path for config", "error", err)
				os.Exit(1)
			}
			config.SetCustomConfigPath(absPath)
			log.Debug("Using custom config path", "path", absPath)
		}

		err := config.InitializeConfig()
		if err != nil {
			log.Error("Error initializing config", "error", err)
			os.Exit(1)
		}
		log.Debug("Config loaded", "config", config.GlobalConfig)

		// Initialize database after config is loaded
		err = db.ConnectDatabase()
		if err != nil {
			log.Error("Error connecting to database", "error", err)
			os.Exit(1)
		}
		log.Debug("Database connected")
	})
}
