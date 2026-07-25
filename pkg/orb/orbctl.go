// Package orb is a thin wrapper around the orbctl CLI. The provider shells out
// to orbctl rather than embedding OrbStack, which is a closed-source macOS app.
package orb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/devsy-org/devsy/pkg/log"
)

// Status values devsy expects on stdout from the status command.
const (
	StatusRunning  = "Running"
	StatusStopped  = "Stopped"
	StatusBusy     = "Busy"
	StatusNotFound = "NotFound"
)

// Client drives OrbStack machines through the orbctl binary.
type Client struct {
	path string
}

// NewClient returns a Client using the given orbctl binary path.
func NewClient(orbctlPath string) *Client {
	return &Client{path: orbctlPath}
}

// CreateParams configure a new machine.
type CreateParams struct {
	Name     string
	Distro   string
	Version  string
	Arch     string
	CPUs     string
	Memory   string
	Disk     string
	Isolated bool
}

// Create creates the machine if absent, installs docker inside it, and ensures
// it is running. Reusing an existing machine skips creation and provisioning.
func (c *Client) Create(ctx context.Context, p CreateParams) error {
	exists, err := c.exists(ctx, p.Name)
	if err != nil {
		return err
	}
	if exists {
		return c.Start(ctx, p.Name)
	}

	if err := c.run(ctx, createArgs(p)...); err != nil {
		return fmt.Errorf("create machine %q: %w", p.Name, err)
	}
	return c.provisionDocker(ctx, p.Name)
}

func createArgs(p CreateParams) []string {
	args := []string{"create"}
	if p.Arch != "" {
		args = append(args, "--arch", p.Arch)
	}
	if p.CPUs != "" {
		args = append(args, "--cpus", p.CPUs)
	}
	if p.Memory != "" {
		args = append(args, "--memory", p.Memory)
	}
	if p.Disk != "" {
		args = append(args, "--disk", p.Disk)
	}
	if p.Isolated {
		args = append(args, "--isolated")
	}
	return append(args, image(p.Distro, p.Version), p.Name)
}

// Start starts a stopped machine.
func (c *Client) Start(ctx context.Context, name string) error {
	if err := c.run(ctx, "start", name); err != nil {
		return fmt.Errorf("start machine %q: %w", name, err)
	}
	return nil
}

// Stop stops a running machine.
func (c *Client) Stop(ctx context.Context, name string) error {
	if err := c.run(ctx, "stop", name); err != nil {
		return fmt.Errorf("stop machine %q: %w", name, err)
	}
	return nil
}

// Delete force-deletes a machine.
func (c *Client) Delete(ctx context.Context, name string) error {
	if err := c.run(ctx, "delete", "--force", name); err != nil {
		return fmt.Errorf("delete machine %q: %w", name, err)
	}
	return nil
}

// Status returns the devsy status for a machine.
func (c *Client) Status(ctx context.Context, name string) (string, error) {
	m, err := c.find(ctx, name)
	if err != nil {
		return "", err
	}
	if m == nil {
		return StatusNotFound, nil
	}
	switch m.State {
	case "running":
		return StatusRunning, nil
	case "stopped":
		return StatusStopped, nil
	default:
		return StatusBusy, nil
	}
}

// Shell runs a command inside the machine with stdio wired through, returning
// orbctl's exit error verbatim so the caller can propagate the exit code.
func (c *Client) Shell(ctx context.Context, name, command string) error {
	//nolint:gosec // fixed orbctl binary; command is provided by devsy.
	cmd := exec.CommandContext(ctx, c.path, "run", "-m", name, "sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// provisionDocker installs and enables docker inside the machine and grants the
// machine's default user access, so the injected devsy agent can drive it.
func (c *Client) provisionDocker(ctx context.Context, name string) error {
	m, err := c.find(ctx, name)
	if err != nil {
		return err
	}
	user := ""
	if m != nil {
		user = m.Config.DefaultUsername
	}

	script := dockerProvisionScript(user)
	//nolint:gosec // fixed orbctl binary; script is a compile-time constant.
	cmd := exec.CommandContext(ctx, c.path, "run", "-m", name, "-u", "root", "sh", "-c", script)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("provision docker in %q: %w", name, err)
	}
	return nil
}

func dockerProvisionScript(user string) string {
	return `set -e
if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable --now docker
` + fmt.Sprintf("usermod -aG docker %q\n", user)
}

type machine struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Config struct {
		DefaultUsername string `json:"default_username"`
	} `json:"config"`
}

func (c *Client) find(ctx context.Context, name string) (*machine, error) {
	out, err := c.output(ctx, "list", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("list machines: %w", err)
	}
	var machines []machine
	if err := json.Unmarshal([]byte(out), &machines); err != nil {
		return nil, fmt.Errorf("parse machine list: %w", err)
	}
	for i := range machines {
		if machines[i].Name == name {
			return &machines[i], nil
		}
	}
	return nil, nil
}

func (c *Client) exists(ctx context.Context, name string) (bool, error) {
	m, err := c.find(ctx, name)
	return m != nil, err
}

func (c *Client) run(ctx context.Context, args ...string) error {
	log.Debugf("running: %s %s", c.path, strings.Join(args, " "))
	//nolint:gosec // fixed orbctl binary; internal args.
	cmd := exec.CommandContext(ctx, c.path, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) output(ctx context.Context, args ...string) (string, error) {
	//nolint:gosec // fixed orbctl binary; internal args.
	out, err := exec.CommandContext(ctx, c.path, args...).Output()
	return string(out), err
}

func image(distro, version string) string {
	if version != "" {
		return distro + ":" + version
	}
	return distro
}
