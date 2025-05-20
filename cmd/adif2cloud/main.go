package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"git.esd.cc/imlonghao/adif2cloud/pkg/watcher"
	"git.esd.cc/imlonghao/adif2cloud/pkg/wavelog"

	"github.com/spf13/viper"
)

func main() {
	// Set up logging format
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Parse command line arguments
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	viper.SetConfigFile(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("Cannot read configuration file", "error", err)
		os.Exit(1)
	}

	// Get target configurations
	var targets []map[string]interface{}
	if err := viper.UnmarshalKey("target", &targets); err != nil {
		slog.Error("Failed to parse target configuration", "error", err)
		os.Exit(1)
	}

	// Create Wavelog clients
	var wavelogClients []*wavelog.Client
	for _, target := range targets {
		if target["type"] == "wavelog" {
			// Get station_profile_id
			stationProfileID := target["station_profile_id"].(int)

			wavelogClient := wavelog.NewClient(
				target["api_url"].(string),
				target["api_key"].(string),
				stationProfileID,
			)
			wavelogClients = append(wavelogClients, wavelogClient)
			slog.Info("Created Wavelog client",
				"api_url", target["api_url"].(string),
				"station_profile_id", stationProfileID)
		}
	}

	// Get source file configuration
	sourceFile := viper.GetString("source")
	if sourceFile == "" {
		slog.Error("Source file path is empty in configuration")
		os.Exit(1)
	}

	// Create watcher for the source file
	var adiWatcher *watcher.ADIWatcher

	// Create callback function to send to all targets
	adiWatcher, err := watcher.NewADIWatcher(sourceFile, func(adiString string) {
		slog.Info("Found new QSO record", "adi", adiString, "source", sourceFile)

		// Send to all Wavelog clients
		for _, client := range wavelogClients {
			if err := client.SendQSO(adiString); err != nil {
				slog.Error("Failed to send QSO record", "error", err)
				continue
			}
			slog.Info("QSO record sent to Wavelog")
		}
	})
	if err != nil {
		slog.Error("Failed to create ADI file watcher", "error", err, "source", sourceFile)
		os.Exit(1)
	}

	// Start the watcher
	if err := adiWatcher.Start(); err != nil {
		slog.Error("Failed to start ADI file watcher", "error", err, "source", sourceFile)
		os.Exit(1)
	}
	slog.Info("Started monitoring ADI file", "path", sourceFile)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	slog.Info("Shutting down...")
	if adiWatcher != nil {
		if err := adiWatcher.Close(); err != nil {
			slog.Error("Failed to close ADI file watcher", "error", err, "source", sourceFile)
		}
	}
	slog.Info("Safely exited")
}
