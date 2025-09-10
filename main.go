package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/vincenty1ung/lastfm-scrobbler/api"
	"github.com/vincenty1ung/lastfm-scrobbler/cmd"
	"github.com/vincenty1ung/lastfm-scrobbler/config"
	"github.com/vincenty1ung/lastfm-scrobbler/core/log"
	"github.com/vincenty1ung/lastfm-scrobbler/core/telemetry"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/model"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/scrobbler"
)

var (
	configFile = new(string)
	isMobile   = new(bool)
)

func main() {
	rootCmd := NewCommand("lastfm-scrobbler", "", "")
	// command.SetHelpTemplate("使用-c 设置配置文件路径\n使用-m 设置true/false")
	rootCmd.Version = "1.0.0"
	rootCmd.Args = cobra.NoArgs
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return initServer() }

	flags := rootCmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(configFile, "config", "c", "config/config.yaml", "config file")
	flags.BoolVarP(isMobile, "mobile", "m", false, "it a mobile")

	// Add sync-records subcommand
	rootCmd.AddCommand(newSyncRecordsCommand())

	// Add memory-tool subcommand
	rootCmd.AddCommand(newMemoryToolCommand())

	// Add music-analysis subcommand
	rootCmd.AddCommand(cmd.NewMusicAnalysisCommand())

	cobra.CheckErr(rootCmd.Execute())
}

func newSyncRecordsCommand() *cobra.Command {
	return cmd.NewSyncRecordsCommand()
}

func newMemoryToolCommand() *cobra.Command {
	return cmd.NewMemoryToolCommand()
}

func initServer() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	c := make(chan struct{})
	config.InitConfig(*configFile)
	logger := log.LogInit(config.ConfigObj.Log.Path, config.ConfigObj.Log.Level, c)

	// Initialize telemetry
	if err := telemetry.Init(config.ConfigObj.Telemetry); err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	defer func(ctx context.Context) {
		err := telemetry.Shutdown(ctx)
		if err != nil {
			panic(err)
		}
	}(context.Background())

	// Initialize database
	if err := model.InitDB(config.ConfigObj.Database.Path, logger); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Start scrobblerRun goroutine
	go api.StartHTTPServer(ctx, config.ConfigObj.Telemetry.Name)

	// Start HTTP server in a separate goroutine
	_ = scrobblerRun(c)

	<-ctx.Done()
	fmt.Println("system exiting")
	close(c)
	return nil
}

func scrobblerRun(c <-chan struct{}) error {
	scrobbler.Init(
		context.Background(),
		config.ConfigObj.Lastfm.ApiKey,
		config.ConfigObj.Lastfm.SharedSecret,
		config.ConfigObj.Lastfm.UserLoginToken,
		*isMobile,
		config.ConfigObj.Lastfm.UserUsername,
		config.ConfigObj.Lastfm.UserPassword,
		config.ConfigObj.Scrobblers,
		c,
	)

	// musixmatch.InitMxmClient(config.ConfigObj.Musixmatch.ApiKey)
	return nil
}
