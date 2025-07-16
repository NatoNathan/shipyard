package cli

import (
	"fmt"
	"os"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "shipyard",
	Short: "Shipyard, Where releases are built",
	Long:  "Shipyard is a tool for managing change notes, versions, and releases.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		initLogger()
	},
	Run: func(cmd *cobra.Command, args []string) {
		println("Welcome to Shipyard!")
		_, err := config.LoadProjectConfig()
		if err != nil {
			println("You need to run `shipyard init` to set up your project.")
			logger.Error("Error loading config", "error", err)
			os.Exit(1)
		}
		logger.Info("Shipyard initialized successfully")
		println("You are in a Shipyard project!")
		println("Use `shipyard --help` to see available commands.")
		println("For more information, visit https://shipyard.tamez.dev/docs")
	},
}

func init() {
	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(AddCmd)
	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(StatusCmd)

	// Add global flags
	RootCmd.PersistentFlags().StringP("config", "c", ".shipyard/config.yaml", "Path to the Shipyard configuration file")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	RootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	RootCmd.PersistentFlags().String("log-file", "", "Log file path (default: .shipyard/logs/shipyard.log)")

	// Bind flags to the AppConfig viper instance
	config.AppConfig.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	config.AppConfig.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	config.AppConfig.BindPFlag("log.level", RootCmd.PersistentFlags().Lookup("log-level"))
	config.AppConfig.BindPFlag("log.file", RootCmd.PersistentFlags().Lookup("log-file"))

	// Add more commands here as needed
	// Example: rootCmd.AddCommand(otherCmd)
}

// initLogger initializes the global logger based on configuration and flags
func initLogger() {
	// Get log level from viper (flags are automatically bound)
	logLevel := config.AppConfig.GetString("log.level")

	// Get verbose flag - if true, override log level to debug
	if config.AppConfig.GetBool("verbose") {
		logLevel = "debug"
	}

	// Get log file from viper
	logFile := config.AppConfig.GetString("log.file")

	// Convert string log level to logger.LogLevel
	var level logger.LogLevel
	switch logLevel {
	case "debug":
		level = logger.DebugLevel
	case "info":
		level = logger.InfoLevel
	case "warn":
		level = logger.WarnLevel
	case "error":
		level = logger.ErrorLevel
	case "fatal":
		level = logger.FatalLevel
	default:
		level = logger.InfoLevel
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	logConfig := &logger.Config{
		Level:      level,
		Output:     os.Stderr,
		TimeFormat: "15:04:05",
		Prefix:     "shipyard",
		LogFile:    logFile,
		CurrentDir: cwd, // Set current working directory for log file
		Version:    "dev",
	}

	// Initialize logger
	if err := logger.Init(logConfig); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
}
