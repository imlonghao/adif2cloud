package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"git.esd.cc/imlonghao/adif2cloud/pkg/s3uploader"
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

	// Create Wavelog clients and S3 Uploaders
	var wavelogClients []*wavelog.Client
	var s3Uploaders []s3uploader.Uploader
	var s3Configs []s3uploader.S3Config

	for _, target := range targets {
		if targetType, ok := target["type"].(string); ok {
			switch targetType {
			case "wavelog":
				stationProfileID, ok := target["station_profile_id"].(int)
				if !ok {
					slog.Error("Failed to parse station_profile_id for wavelog, or it's not a number", "target", target)
					continue
				}

				apiURL, _ := target["api_url"].(string)
				apiKey, _ := target["api_key"].(string)
				if apiURL == "" || apiKey == "" {
					slog.Error("api_url or api_key is missing for wavelog target", "target", target)
					continue
				}

				wavelogClient := wavelog.NewClient(apiURL, apiKey, stationProfileID)
				wavelogClients = append(wavelogClients, wavelogClient)
				slog.Info("Created Wavelog client", "api_url", apiURL, "station_profile_id", stationProfileID)

			case "s3":
				s3Config := s3uploader.S3Config{}
				if endpoint, ok := target["endpoint"].(string); ok {
					s3Config.Endpoint = endpoint
				}
				if region, ok := target["region"].(string); ok {
					s3Config.Region = region
				}
				if accessKeyID, ok := target["access_key_id"].(string); ok {
					s3Config.AccessKeyID = accessKeyID
				}
				if secretAccessKey, ok := target["secret_access_key"].(string); ok {
					s3Config.SecretAccessKey = secretAccessKey
				}
				if bucketName, ok := target["bucket_name"].(string); ok {
					s3Config.BucketName = bucketName
				}
				if usePathStyle, ok := target["use_path_style"].(bool); ok {
					s3Config.UsePathStyle = usePathStyle
				}
				if fileName, ok := target["file_name"].(string); ok {
					s3Config.FileName = fileName
				}

				if s3Config.BucketName == "" {
					slog.Error("bucket_name is required for s3 target", "target", target)
					continue
				}

				uploader, err := s3uploader.NewS3Uploader(s3Config)
				if err != nil {
					slog.Error("Failed to create S3 uploader", "error", err, "config", s3Config)
					continue
				}
				s3Uploaders = append(s3Uploaders, uploader)
				s3Configs = append(s3Configs, s3Config)
				slog.Info("Created S3 Uploader", "endpoint", s3Config.Endpoint, "bucket", s3Config.BucketName)
			default:
				slog.Warn("Unknown target type", "type", targetType)
			}
		} else {
			slog.Warn("Target type not specified or not a string", "target", target)
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
				slog.Error("Failed to send QSO record to Wavelog", "error", err)
				continue
			}
			slog.Info("QSO record sent to Wavelog")
		}

		// Upload the entire source file to S3, overwriting if it exists
		for i, uploader := range s3Uploaders {
			s3Cfg := s3Configs[i]
			objectKey := s3Cfg.FileName
			if objectKey == "" {
				objectKey = filepath.Base(sourceFile)
			}

			slog.Info("Attempting to upload file to S3", "source_file", sourceFile, "bucket", s3Cfg.BucketName, "object_key", objectKey)
			if err := uploader.UploadFile(sourceFile, objectKey); err != nil {
				slog.Error("Failed to upload file to S3", "error", err, "source_file", sourceFile, "bucket", s3Cfg.BucketName, "key", objectKey)
			} else {
				slog.Info("Successfully uploaded file to S3", "source_file", sourceFile, "bucket", s3Cfg.BucketName, "key", objectKey)
			}
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
