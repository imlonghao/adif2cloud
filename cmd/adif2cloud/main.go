package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"git.esd.cc/imlonghao/adif2cloud/internal/consts"
	_ "git.esd.cc/imlonghao/adif2cloud/internal/winres"
	"git.esd.cc/imlonghao/adif2cloud/pkg/clublog"
	"git.esd.cc/imlonghao/adif2cloud/pkg/git"
	"git.esd.cc/imlonghao/adif2cloud/pkg/provider"
	"git.esd.cc/imlonghao/adif2cloud/pkg/s3"
	"git.esd.cc/imlonghao/adif2cloud/pkg/watcher"
	"git.esd.cc/imlonghao/adif2cloud/pkg/wavelog"
	"git.esd.cc/imlonghao/adif2cloud/pkg/webhook"

	"github.com/spf13/viper"
)

func main() {
	// Set up logging format
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting adif2cloud", "version", consts.Version)

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

	// Create providers
	var providers []provider.Provider

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

				wavelogProvider := wavelog.NewWavelogProvider(apiURL, apiKey, stationProfileID)
				providers = append(providers, wavelogProvider)
				slog.Info("Created Wavelog provider", "api_url", apiURL, "station_profile_id", stationProfileID)

			case "clublog":
				email, _ := target["email"].(string)
				password, _ := target["password"].(string)
				callsign, _ := target["callsign"].(string)
				if email == "" || password == "" || callsign == "" {
					slog.Error("email, password or callsign is missing for clublog target", "target", target)
					continue
				}

				clublogProvider := clublog.NewClubLogProvider(clublog.ClubLogConfig{
					Email:    email,
					Password: password,
					Callsign: callsign,
				})
				if clublogProvider == nil {
					slog.Error("Failed to create Club Log provider", "email", email, "callsign", callsign)
					continue
				}
				providers = append(providers, clublogProvider)
				slog.Info("Created Club Log provider", "email", email, "callsign", callsign)

			case "git":
				gitConfig := git.GitConfig{}
				if repoURL, ok := target["repo_url"].(string); ok {
					gitConfig.RepoURL = repoURL
				}
				if branch, ok := target["branch"].(string); ok {
					gitConfig.Branch = branch
				}
				if fileName, ok := target["file_name"].(string); ok {
					gitConfig.FileName = fileName
				}
				if commitAuthor, ok := target["commit_author"].(string); ok {
					gitConfig.CommitAuthor = commitAuthor
				}
				if commitEmail, ok := target["commit_email"].(string); ok {
					gitConfig.CommitEmail = commitEmail
				}
				if authUsername, ok := target["auth_username"].(string); ok {
					gitConfig.AuthUsername = authUsername
				}
				if authPassword, ok := target["auth_password"].(string); ok {
					gitConfig.AuthPassword = authPassword
				}
				if authSSHKey, ok := target["auth_ssh_key"].(string); ok {
					gitConfig.AuthSSHKey = authSSHKey
				}
				if authSSHKeyPassphrase, ok := target["auth_ssh_key_passphrase"].(string); ok {
					gitConfig.AuthSSHKeyPassphrase = authSSHKeyPassphrase
				}

				if gitConfig.RepoURL == "" {
					slog.Error("repo_url is required for git target", "target", target)
					continue
				}

				gitProvider, err := git.NewGitProvider(gitConfig)
				if err != nil {
					slog.Error("Failed to create Git provider", "error", err, "config", gitConfig)
					continue
				}
				providers = append(providers, gitProvider)
				slog.Info("Created Git provider", "repo_url", gitConfig.RepoURL, "branch", gitConfig.Branch)

			case "webhook":
				url, _ := target["url"].(string)
				if url == "" {
					slog.Error("url is required for webhook target", "target", target)
					continue
				}

				webhookProvider := webhook.NewWebhookProvider(webhook.WebhookConfig{
					URL: url,
				})
				providers = append(providers, webhookProvider)
				slog.Info("Created Webhook provider", "url", url)

			case "s3":
				s3Config := s3.S3Config{}
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

				s3Provider, err := s3.NewS3Provider(s3Config)
				if err != nil {
					slog.Error("Failed to create S3 provider", "error", err, "config", s3Config)
					continue
				}
				providers = append(providers, s3Provider)
				slog.Info("Created S3 provider", "endpoint", s3Config.Endpoint, "bucket", s3Config.BucketName)
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

	// 获取本地文件大小
	localFileInfo, err := os.Stat(sourceFile)
	if err != nil {
		slog.Error("Failed to get local file info", "error", err)
		os.Exit(1)
	}
	localSize := localFileInfo.Size()

	// 遍历所有 providers 获取远程文件大小
	var maxRemoteSize int64
	var maxRemoteProvider provider.Provider
	for _, p := range providers {
		remoteSize, err := p.GetSize()
		if err != nil {
			slog.Warn("Failed to get remote file size", "error", err)
			continue
		}
		if remoteSize > maxRemoteSize {
			maxRemoteSize = remoteSize
			maxRemoteProvider = p
		}
	}

	// 如果远程文件更大，则下载替换本地文件
	if maxRemoteSize > localSize {
		slog.Info("Remote file is larger, downloading...",
			"local_size", localSize,
			"remote_size", maxRemoteSize)

		sourceFileWriter, err := os.OpenFile(sourceFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			slog.Error("Failed to open local file", "error", err)
			os.Exit(1)
		}
		// 直接下载到目标文件
		if err := maxRemoteProvider.Download(sourceFileWriter); err != nil {
			slog.Error("Failed to download remote file", "error", err)
			os.Exit(1)
		}
		sourceFileWriter.Close()

		slog.Info("Successfully replaced local file with remote file")
	}

	// Create watcher for the source file
	adiWatcher, err := watcher.NewADIWatcher(sourceFile, func(adiString string) {
		slog.Info("Found new QSO record", "adi", adiString)
		// Send to all providers
		for _, p := range providers {
			logger := slog.With("provider", p.GetName())
			if err := p.Upload(sourceFile, adiString); err != nil {
				logger.Error("Failed to upload to provider", "error", err)
				continue
			}
			logger.Info("Successfully uploaded to provider")
		}
	})
	if err != nil {
		slog.Error("Failed to create ADI file watcher", "error", err)
		os.Exit(1)
	}

	// Start the watcher
	if err := adiWatcher.Start(); err != nil {
		slog.Error("Failed to start ADI file watcher", "error", err)
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
		adiWatcher.Close()
	}
	slog.Info("Safely exited")
}
