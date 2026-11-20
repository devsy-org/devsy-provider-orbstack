package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/devsy-org/devsy/logging"
	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "devsy-provider-orbstack",
		Short:         "OrbStack Provider commands",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}

func BuildRoot() *cobra.Command {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewDeleteCmd())
	rootCmd.AddCommand(NewCommandCmd())
	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewStopCmd())
	rootCmd.AddCommand(NewStatusCmd())
	return rootCmd
}

func initLogger() error {
	cfg, invalid := logging.ResolveProvider(
		os.Getenv("DEVSY_LOG_LEVEL"), os.Getenv("DEVSY_LOG_FORMAT"),
	)
	if invalid.Level {
		return errors.New("invalid DEVSY_LOG_LEVEL")
	}
	if invalid.Format {
		return errors.New("invalid DEVSY_LOG_FORMAT")
	}
	log.Init(
		log.Config{
			Level:        logging.LevelName(cfg.Level),
			DefaultLevel: log.LevelInfoName,
			Format:       string(cfg.Format),
		},
	)
	return nil
}

// Execute propagates the inner process's exit code so devsy sees it verbatim.
func Execute() {
	if err := initLogger(); err != nil {
		log.Init(log.Config{DefaultLevel: log.LevelInfoName, Format: "json"})
		log.Errorf("%v", err)
		_ = log.Sync()
		os.Exit(1)
	}
	if err := BuildRoot().Execute(); err != nil {
		_ = log.Sync()
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		log.Errorf("%v", err)
		_ = log.Sync()
		os.Exit(1)
	}
	_ = log.Sync()
}
