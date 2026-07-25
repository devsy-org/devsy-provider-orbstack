package cmd

import (
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-orbstack/pkg/options"
	"github.com/devsy-org/devsy-provider-orbstack/pkg/orb"
	"github.com/spf13/cobra"
)

func NewCommandCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "command",
		Short: "Run a command inside an OrbStack machine",
		RunE: func(cmd *cobra.Command, _ []string) error {
			command := os.Getenv("COMMAND")
			if command == "" {
				return fmt.Errorf("COMMAND environment variable is missing")
			}
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}
			return orb.NewClient(opts.OrbctlPath).Shell(cmd.Context(), opts.MachineID, command)
		},
	}
}
