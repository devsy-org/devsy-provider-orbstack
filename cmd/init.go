package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/devsy-org/devsy-provider-orbstack/pkg/options"
	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Validate the OrbStack provider is ready",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(false)
			if err != nil {
				return err
			}
			return checkOrbctl(cmd.Context(), opts.OrbctlPath)
		},
	}
}

func checkOrbctl(ctx context.Context, path string) error {
	//nolint:gosec // fixed orbctl binary from provider options.
	out, err := exec.CommandContext(ctx, path, "version").CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"orbctl not found or not runnable at %q: %w: %s (install OrbStack: https://orbstack.dev)",
			path,
			err,
			strings.TrimSpace(string(out)),
		)
	}
	return nil
}
