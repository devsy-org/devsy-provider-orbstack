// Command provider generates provider.yaml for a given version, resolving binary
// download URLs and checksums from ./dist/checksums.txt.
//
// Usage: go run ./hack/provider/main.go <version> > ./dist/provider.yaml
package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	providerName = "orbstack"
	githubOwner  = "devsy-org"
	githubRepo   = "devsy-provider-orbstack"
	valueTrue    = "true"
)

type Provider struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	OptionGroups []OptionGroup     `yaml:"optionGroups"`
	Options      Options           `yaml:"options"`
	Agent        Agent             `yaml:"agent"`
	Binaries     Binaries          `yaml:"binaries"`
	Exec         map[string]string `yaml:"exec"`
}

type OptionGroup struct {
	Name           string   `yaml:"name"`
	DefaultVisible bool     `yaml:"defaultVisible"`
	Options        []string `yaml:"options"`
}

type Options map[string]Option

type Option struct {
	Description string   `yaml:"description,omitempty"`
	Default     string   `yaml:"default,omitempty"`
	Suggestions []string `yaml:"suggestions,omitempty"`
}

type Agent struct {
	Path                    string            `yaml:"path"`
	Driver                  string            `yaml:"driver"`
	Docker                  DockerAgent       `yaml:"docker"`
	InactivityTimeout       string            `yaml:"inactivityTimeout"`
	InjectGitCredentials    string            `yaml:"injectGitCredentials"`
	InjectDockerCredentials string            `yaml:"injectDockerCredentials"`
	Exec                    map[string]string `yaml:"exec"`
}

type DockerAgent struct {
	Path    string `yaml:"path"`
	Install string `yaml:"install"`
}

type Binaries struct {
	OrbstackProvider []Binary `yaml:"ORBSTACK_PROVIDER"`
}

type Binary struct {
	OS       string `yaml:"os"`
	Arch     string `yaml:"arch"`
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum"`
}

type buildConfig struct {
	version     string
	projectRoot string
	isRelease   bool
	checksums   map[string]string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("expected version as argument")
	}
	cfg, err := newBuildConfig(os.Args[1])
	if err != nil {
		return err
	}
	output, err := yaml.Marshal(buildProvider(cfg))
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	_, err = os.Stdout.Write(output)
	return err
}

func newBuildConfig(version string) (*buildConfig, error) {
	checksums, err := parseChecksums("./dist/checksums.txt")
	if err != nil {
		return nil, fmt.Errorf("parse checksums: %w", err)
	}
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		owner := envOr("GITHUB_OWNER", githubOwner)
		projectRoot = fmt.Sprintf(
			"https://github.com/%s/%s/releases/download/%s", owner, githubRepo, version)
	}
	isRelease := strings.Contains(projectRoot, "github.com") &&
		strings.Contains(projectRoot, "/releases/")
	return &buildConfig{version, projectRoot, isRelease, checksums}, nil
}

func buildProvider(cfg *buildConfig) Provider {
	return Provider{
		Name:         providerName,
		Version:      cfg.version,
		Description:  "Devsy on OrbStack machines.",
		OptionGroups: buildOptionGroups(),
		Options:      buildOptions(),
		Agent:        buildAgent(),
		Binaries:     Binaries{OrbstackProvider: buildBinaryList(cfg)},
		Exec: map[string]string{
			"init":    "${ORBSTACK_PROVIDER} init",
			"command": "${ORBSTACK_PROVIDER} command",
			"create":  "${ORBSTACK_PROVIDER} create",
			"delete":  "${ORBSTACK_PROVIDER} delete",
			"start":   "${ORBSTACK_PROVIDER} start",
			"stop":    "${ORBSTACK_PROVIDER} stop",
			"status":  "${ORBSTACK_PROVIDER} status",
		},
	}
}

func buildOptionGroups() []OptionGroup {
	return []OptionGroup{
		{
			Name:           "Machine options",
			DefaultVisible: true,
			Options: []string{
				"ORBSTACK_DISTRO", "ORBSTACK_VERSION", "ORBSTACK_ARCH",
				"ORBSTACK_CPUS", "ORBSTACK_MEMORY", "ORBSTACK_DISK", "ORBSTACK_ISOLATED",
			},
		},
		{
			Name:           "Agent options",
			DefaultVisible: false,
			Options: []string{
				"ORBSTACK_PATH", "AGENT_PATH", "INACTIVITY_TIMEOUT",
				"INJECT_DOCKER_CREDENTIALS", "INJECT_GIT_CREDENTIALS",
			},
		},
	}
}

func buildOptions() Options {
	return Options{
		"ORBSTACK_DISTRO": {
			Description: "Linux distribution for the machine.",
			Default:     "ubuntu",
			Suggestions: []string{"ubuntu", "debian", "fedora", "alma", "centos", "alpine", "arch"},
		},
		"ORBSTACK_VERSION": {Description: "Distro version (e.g. 24.04). Empty uses the latest."},
		"ORBSTACK_ARCH": {
			Description: "CPU architecture. Empty uses the host architecture.",
			Suggestions: []string{"arm64", "amd64"},
		},
		"ORBSTACK_CPUS": {
			Description: "CPU core limit for the machine. Empty uses the OrbStack default.",
		},
		"ORBSTACK_MEMORY": {
			Description: "Memory limit for the machine, e.g. 4G. Empty uses the OrbStack default.",
		},
		"ORBSTACK_DISK": {
			Description: "Disk limit for the machine, e.g. 64G. Empty uses the OrbStack default.",
		},
		"ORBSTACK_ISOLATED": {
			Description: "Create an isolated machine (disables host file sharing and integration). " +
				"Recommended for a clean workspace boundary.",
			Default: valueTrue,
		},
		"ORBSTACK_PATH": {Description: "Path to the orbctl binary.", Default: "orbctl"},
		"AGENT_PATH": {
			Description: "The path where to inject the Devsy agent inside the machine.",
			Default:     "/tmp/devsy",
		},
		"INACTIVITY_TIMEOUT": {
			Description: "If defined, automatically stops the machine after the inactivity period.",
			Default:     "10m",
		},
		"INJECT_GIT_CREDENTIALS": {
			Description: "If Devsy should inject git credentials into the machine.",
			Default:     valueTrue,
		},
		"INJECT_DOCKER_CREDENTIALS": {
			Description: "If Devsy should inject docker credentials into the machine.",
			Default:     valueTrue,
		},
	}
}

func buildAgent() Agent {
	//nolint:gosec // manifest template literals, not credentials
	return Agent{
		Path:                    "${AGENT_PATH}",
		Driver:                  "docker",
		Docker:                  DockerAgent{Path: "docker", Install: "false"},
		InactivityTimeout:       "${INACTIVITY_TIMEOUT}",
		InjectGitCredentials:    "${INJECT_GIT_CREDENTIALS}",
		InjectDockerCredentials: "${INJECT_DOCKER_CREDENTIALS}",
		// Halting the guest transitions the machine to stopped, which is how the
		// inactivity timeout stops it.
		Exec: map[string]string{"shutdown": "sudo shutdown -h now || shutdown -h now"},
	}
}

func buildBinaryList(cfg *buildConfig) []Binary {
	// OrbStack is macOS-only, so the provider binary ships for darwin only.
	platforms := []string{"darwin/amd64", "darwin/arm64"}
	result := make([]Binary, 0, len(platforms))
	for _, platform := range platforms {
		result = append(result, buildBinary(cfg, platform))
	}
	return result
}

func buildBinary(cfg *buildConfig, platform string) Binary {
	goos, arch, _ := strings.Cut(platform, "/")
	path := cfg.projectRoot
	if !cfg.isRelease {
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			path, _ = url.JoinPath(path, buildDir(platform))
		} else {
			abs, _ := filepath.Abs(path)
			path = filepath.Join(abs, buildDir(platform))
		}
	}
	filename := fmt.Sprintf("devsy-provider-%s-%s-%s", providerName, goos, arch)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		path, _ = url.JoinPath(path, filename)
	} else {
		path = filepath.Join(path, filename)
	}
	return Binary{OS: goos, Arch: arch, Path: path, Checksum: cfg.checksums[filename]}
}

func buildDir(platform string) string {
	return map[string]string{
		"darwin/amd64": "build_darwin_amd64_v1",
		"darwin/arm64": "build_darwin_arm64_v8.0",
	}[platform]
}

func parseChecksums(path string) (map[string]string, error) {
	file, err := os.Open(path) //nolint:gosec // build-time tool, fixed path.
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if checksum, filename, ok := strings.Cut(scanner.Text(), " "); ok {
			checksums[strings.TrimSpace(filename)] = checksum
		}
	}
	return checksums, scanner.Err()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
