package cmd

import (
	"github.com/devsy-org/devsy-provider-orbstack/pkg/options"
	"github.com/devsy-org/devsy-provider-orbstack/pkg/orb"
	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create an OrbStack machine",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts, err := options.FromEnv(true)
			if err != nil {
				return err
			}
			log.Infof("creating OrbStack machine %q (distro=%s version=%s isolated=%t)",
				opts.MachineID, opts.Distro, opts.Version, opts.Isolated)
			return orb.NewClient(opts.OrbctlPath).Create(cmd.Context(), orb.CreateParams{
				Name:     opts.MachineID,
				Distro:   opts.Distro,
				Version:  opts.Version,
				Arch:     opts.Arch,
				CPUs:     opts.CPUs,
				Memory:   opts.Memory,
				Disk:     opts.Disk,
				Isolated: opts.Isolated,
			})
		},
	}
}
