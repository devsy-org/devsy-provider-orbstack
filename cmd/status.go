package cmd

import (
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-orbstack/pkg/options"
	"github.com/devsy-org/devsy-provider-orbstack/pkg/orb"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Get the status of an OrbStack machine",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}
			status, err := orb.NewClient(opts.OrbctlPath).Status(cmd.Context(), opts.MachineID)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(os.Stdout, status)
			return err
		},
	}
}
