package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"     // Version will be set at build time
	GitCommit = "unknown" // Git commit hash
	BuildDate = "unknown" // Build date
)

// showVersion prints version information
func showVersion() {
	fmt.Printf("Shipyard %s\n", Version)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

var rootCmd = &cobra.Command{
	Use:   "shipyard",
	Short: "Shipyard, Where releases are built",
	Long:  "Shipyard is a tool for managing change notes, versions, and releases.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		initLogger()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Check if version flag is set
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			showVersion()
			return
		}
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
	rootCmd.AddCommand(InitCmd)
	rootCmd.AddCommand(AddCmd)
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(StatusCmd)

	// Add global flags
	rootCmd.PersistentFlags().StringP("config", "c", ".shipyard/config.yaml", "Path to the Shipyard configuration file")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-file", "", "Log file path (default: .shipyard/logs/shipyard.log)")

	// Add version flag
	rootCmd.Flags().BoolP("version", "V", false, "Show version information")

	// Bind flags to the AppConfig viper instance
	config.AppConfig.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	config.AppConfig.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	config.AppConfig.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
	config.AppConfig.BindPFlag("log.file", rootCmd.PersistentFlags().Lookup("log-file"))

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
	// Create logger config
	logConfig := &logger.Config{
		Level:      level,
		Output:     os.Stderr,
		TimeFormat: "15:04:05",
		Prefix:     "shipyard",
		LogFile:    logFile,
		CurrentDir: cwd,     // Set current working directory for log file
		Version:    Version, // Set version for logging
	}

	// Initialize logger
	if err := logger.Init(logConfig); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
