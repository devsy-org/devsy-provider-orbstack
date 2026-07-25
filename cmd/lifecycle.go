package cmd

import (
	"github.com/devsy-org/devsy-provider-orbstack/pkg/options"
	"github.com/devsy-org/devsy-provider-orbstack/pkg/orb"
	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start an OrbStack machine",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}
			return orb.NewClient(opts.OrbctlPath).Start(cmd.Context(), opts.MachineID)
		},
	}
}

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop an OrbStack machine",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}
			return orb.NewClient(opts.OrbctlPath).Stop(cmd.Context(), opts.MachineID)
		},
	}
}

func NewDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete an OrbStack machine",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}
			return orb.NewClient(opts.OrbctlPath).Delete(cmd.Context(), opts.MachineID)
		},
	}
}
