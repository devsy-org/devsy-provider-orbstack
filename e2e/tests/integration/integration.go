package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/devsy-org/devsy/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func providerBinary() string {
	archDir := map[string]string{"amd64": "amd64_v1", "arm64": "arm64_v8.0"}[goruntime.GOARCH]
	dir := fmt.Sprintf("build_%s_%s", goruntime.GOOS, archDir)
	bin := fmt.Sprintf("devsy-provider-orbstack-%s-%s", goruntime.GOOS, goruntime.GOARCH)
	return filepath.Join("..", "dist", dir, bin)
}

func mustRun(cmd *exec.Cmd) {
	out, err := cmd.CombinedOutput()
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), string(out))
}

func isolateDevsyHome() {
	dir, err := os.MkdirTemp("", "devsy-orbstack-e2e-")
	framework.ExpectNoError(err)
	framework.ExpectNoError(os.Setenv("DEVSY_HOME", dir))
}

func setupProvider() {
	cmd := exec.Command("go", "run", "hack/provider/main.go", "0.0.0")
	cmd.Dir = "../"
	projectRoot, err := filepath.Abs("../")
	framework.ExpectNoError(err)
	cmd.Env = append(os.Environ(), "PROJECT_ROOT="+filepath.Join(projectRoot, "dist"))

	output, err := cmd.Output()
	framework.ExpectNoError(err)
	framework.ExpectNoError(os.WriteFile("../dist/provider.yaml", output, 0o600))
}

func setupDevsyCLI() {
	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf(
		"https://github.com/devsy-org/devsy/releases/latest/download/devsy-%s-%s",
		goruntime.GOOS, goruntime.GOARCH,
	)
	resp, err := client.Get(url) //nolint:gosec // fixed release URL
	framework.ExpectNoError(err)
	framework.ExpectEqual(resp.StatusCode, http.StatusOK)
	defer func() { _ = resp.Body.Close() }()

	framework.ExpectNoError(os.MkdirAll("bin", 0o750))
	binPath := filepath.Join("bin", "devsy")
	//nolint:gosec // fixed path; the devsy binary needs the executable bit
	out, err := os.OpenFile(binPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	framework.ExpectNoError(err)
	_, err = io.Copy(out, resp.Body)
	framework.ExpectNoError(err)
	framework.ExpectNoError(out.Close())
	framework.ExpectNoError(exec.Command(binPath, "--version").Run()) //nolint:gosec // fixed path
}

var _ = ginkgo.Describe(
	"devsy provider orbstack integration",
	ginkgo.Label("integration"),
	ginkgo.Ordered,
	func() {
		ginkgo.BeforeAll(func() {
			setupProvider()
		})

		ginkgo.It("should generate a non-empty provider.yaml", func() {
			data, err := os.ReadFile("../dist/provider.yaml")
			framework.ExpectNoError(err)
			framework.ExpectNotEqual(len(data), 0)
		})

		ginkgo.It("should have required provider options", func() {
			data, err := os.ReadFile("../dist/provider.yaml")
			framework.ExpectNoError(err)
			content := string(data)
			for _, opt := range []string{"ORBSTACK_DISTRO", "ORBSTACK_ISOLATED", "AGENT_PATH"} {
				gomega.Expect(content).To(gomega.ContainSubstring(opt))
			}
		})

		ginkgo.It("should fail init when orbctl is missing", func() {
			cmd := exec.Command(providerBinary(), "init") //nolint:gosec // built provider path
			cmd.Env = append(cmd.Environ(), "ORBSTACK_PATH=/nonexistent/orbctl")
			framework.ExpectError(cmd.Run())
		})

		ginkgo.It("should fail create without a machine id", func() {
			cmd := exec.Command(providerBinary(), "create") //nolint:gosec // built provider path
			cmd.Env = append(cmd.Environ(), "MACHINE_ID=")
			output, err := cmd.CombinedOutput()
			framework.ExpectError(err)
			gomega.Expect(string(output)).To(gomega.ContainSubstring("MACHINE_ID"))
		})
	},
)

var _ = ginkgo.Describe(
	"devsy provider orbstack vm lifecycle",
	ginkgo.Label("vm"),
	ginkgo.Ordered,
	func() {
		const workspaceID = "devsy-provider-orbstack"

		ginkgo.BeforeAll(func() {
			isolateDevsyHome()
			setupProvider()
			setupDevsyCLI()
			mustRun(exec.Command("bin/devsy", "provider", "add", "../dist/provider.yaml",
				"--name", "orbstack", "-o", "ORBSTACK_CPUS=2", "-o", "ORBSTACK_MEMORY=4G"))

			ginkgo.DeferCleanup(func() {
				_ = exec.Command("bin/devsy", "workspace", "delete", "--force", workspaceID).Run()
			})
		})

		ginkgo.It("should bring up a workspace on an OrbStack machine", func() {
			mustRun(exec.Command("bin/devsy", "workspace", "up", "--ide=none",
				"--id", workspaceID, "--provider", "orbstack", "fixtures/workspace"))
		})

		ginkgo.It("should run a command in the workspace", func() {
			cmd := exec.Command(
				"bin/devsy",
				"workspace",
				"ssh",
				workspaceID,
				"--command",
				"echo test",
			)
			output, err := cmd.Output()
			framework.ExpectNoError(err)
			gomega.Expect(strings.TrimSpace(string(output))).To(gomega.Equal("test"))
		})

		ginkgo.It("should delete the workspace", func() {
			mustRun(exec.Command("bin/devsy", "workspace", "delete", "--force", workspaceID))
		})
	},
)
