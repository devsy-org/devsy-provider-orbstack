// Package options reads the provider configuration devsy passes as environment
// variables when it invokes the provider binary.
package options

import (
	"fmt"
	"os"
)

// Options holds resolved provider configuration for a single invocation.
type Options struct {
	MachineID  string
	OrbctlPath string
	Distro     string
	Version    string
	Arch       string
	CPUs       string
	Memory     string
	Disk       string
	Isolated   bool
}

// FromEnv reads and validates options. requireMachineID is true for lifecycle
// verbs, which always operate on a specific machine.
func FromEnv(requireMachineID bool) (*Options, error) {
	o := &Options{
		MachineID:  os.Getenv("MACHINE_ID"),
		OrbctlPath: envOr("ORBSTACK_PATH", "orbctl"),
		Distro:     envOr("ORBSTACK_DISTRO", "ubuntu"),
		Version:    os.Getenv("ORBSTACK_VERSION"),
		Arch:       os.Getenv("ORBSTACK_ARCH"),
		CPUs:       os.Getenv("ORBSTACK_CPUS"),
		Memory:     os.Getenv("ORBSTACK_MEMORY"),
		Disk:       os.Getenv("ORBSTACK_DISK"),
		Isolated:   envOr("ORBSTACK_ISOLATED", "true") != "false",
	}

	if requireMachineID && o.MachineID == "" {
		return nil, fmt.Errorf("MACHINE_ID environment variable is missing")
	}

	return o, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
