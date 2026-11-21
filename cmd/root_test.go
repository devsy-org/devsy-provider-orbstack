package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devsy-org/devsy/pkg/log"
)

const (
	testInfoLevel  = "info"
	testJSONFormat = "json"
	testDebugLevel = "debug"
	testTraceLevel = "trace"
	testWarnLevel  = "warn"
)

func TestProviderLogging(t *testing.T) {
	if os.Getenv("ORBSTACK_LOG_TEST_HELPER") == "1" {
		serveOrbLoggingFixture()
		return
	}

	for _, tc := range []struct {
		name   string
		level  string
		format string
		want   []string
	}{
		{"default", "", "", []string{testInfoLevel, testWarnLevel}},
		{"error", "error", testJSONFormat, nil},
		{testWarnLevel, testWarnLevel, testJSONFormat, []string{testWarnLevel}},
		{testInfoLevel, testInfoLevel, testJSONFormat, []string{testInfoLevel, testWarnLevel}},
		{testDebugLevel, testDebugLevel, testJSONFormat, []string{testDebugLevel, testInfoLevel, testWarnLevel}},
		{testTraceLevel, testTraceLevel, "auto", []string{testTraceLevel, testDebugLevel, testInfoLevel, testWarnLevel}},
		{"uppercase", "TRACE", "AUTO", []string{testTraceLevel, testDebugLevel, testInfoLevel, testWarnLevel}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			canary := t.Name() + "-private-value"
			cmd := orbLoggingFixture(t)
			cmd.Env = append(
				os.Environ(),
				"ORBSTACK_LOG_TEST_HELPER=1",
				"ORBSTACK_LOG_TEST_EXECUTE=0",
				"DEVSY_LOG_LEVEL="+tc.level,
				"DEVSY_LOG_FORMAT="+tc.format,
				"ORBSTACK_TEST_TOKEN="+canary,
			)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("helper: %v: %s", err, stderr.String())
			}
			if stdout.String() != "protocol output" {
				t.Fatalf("stdout = %q", stdout.String())
			}
			assertRedactedOrbInfo(t, stderr.String(), canary)
			levels := orbLoggedLevels(t, &stderr)
			if strings.Join(levels, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("levels = %v, want %v", levels, tc.want)
			}
		})
	}
}

func TestProviderTextFormats(t *testing.T) {
	for _, format := range []string{"text", "plain", "logfmt", "PLAIN", "LOGFMT"} {
		t.Run(format, func(t *testing.T) {
			cmd := orbLoggingFixture(t)
			cmd.Env = append(
				os.Environ(),
				"ORBSTACK_LOG_TEST_HELPER=1",
				"ORBSTACK_LOG_TEST_EXECUTE=0",
				"DEVSY_LOG_LEVEL=warn",
				"DEVSY_LOG_FORMAT="+format,
			)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("helper: %v: %s", err, stderr.String())
			}
			if stdout.String() != "protocol output" ||
				!strings.Contains(stderr.String(), "warn event") ||
				strings.Contains(stderr.String(), "info event") {
				t.Fatalf("stdout = %q, stderr = %q", stdout.String(), stderr.String())
			}
		})
	}
}

func TestInvalidLoggingConfigStopsBeforeBackend(t *testing.T) {
	for _, name := range []string{"DEVSY_LOG_LEVEL", "DEVSY_LOG_FORMAT"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			marker := filepath.Join(dir, "called")
			orbctl := prepareOrbBackend(t, dir)
			cmd := orbLoggingFixture(t)
			cmd.Env = append(
				os.Environ(),
				"ORBSTACK_LOG_TEST_HELPER=1",
				"ORBSTACK_LOG_TEST_EXECUTE=1",
				"DEVSY_LOG_LEVEL=info",
				"DEVSY_LOG_FORMAT=json",
				"ORBSTACK_PATH="+orbctl,
				"ORBSTACK_TEST_MARKER="+marker,
				name+"=invalid-secret-canary",
			)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Run(); err == nil {
				t.Fatal("invalid configuration succeeded")
			}
			if stdout.Len() != 0 || strings.Contains(stderr.String(), "invalid-secret-canary") ||
				!strings.Contains(stderr.String(), "invalid "+name) {
				t.Fatalf("stdout = %q, stderr = %q", stdout.String(), stderr.String())
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("backend invoked or marker inaccessible: %v", err)
			}
		})
	}
}

func serveOrbLoggingFixture() {
	if os.Getenv("ORBSTACK_LOG_TEST_EXECUTE") == "1" {
		os.Args = []string{"provider", "create"}
		Execute()
		os.Exit(0)
	}
	if err := initLogger(); err != nil {
		os.Exit(2)
	}
	log.Tracef("trace event")
	log.Debugf("debug event")
	log.Infof("info event %s", os.Getenv("ORBSTACK_TEST_TOKEN"))
	log.Warnf("warn event")
	_ = log.Sync()
	if _, err := os.Stdout.WriteString("protocol output"); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func orbLoggedLevels(t *testing.T, stderr *bytes.Buffer) []string {
	t.Helper()
	var levels []string
	dec := json.NewDecoder(stderr)
	for dec.More() {
		var record struct {
			Time    string `json:"ts"`
			Level   string `json:"level"`
			Message string `json:"msg"`
		}
		if err := dec.Decode(&record); err != nil {
			t.Fatal(err)
		}
		if record.Time == "" || !strings.Contains(record.Message, "event") {
			t.Fatalf("invalid diagnostic: %+v", record)
		}
		levels = append(levels, record.Level)
	}
	return levels
}

func orbLoggingFixture(t *testing.T) *exec.Cmd {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return &exec.Cmd{
		Path: executable,
		Args: []string{executable, "-test.run=^TestProviderLogging$"},
	}
}

func prepareOrbBackend(t *testing.T, dir string) string {
	t.Helper()
	orbctl := filepath.Join(dir, "orbctl")
	script := "#!/bin/sh\ntouch \"$ORBSTACK_TEST_MARKER\"\n"
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	file, err := root.OpenFile("orbctl", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	if _, err := file.WriteString(script); err != nil {
		t.Fatal(err)
	}
	if err := file.Chmod(0o700); err != nil {
		t.Fatal(err)
	}
	return orbctl
}

func assertRedactedOrbInfo(t *testing.T, output, canary string) {
	t.Helper()
	if strings.Contains(output, "info event") &&
		!strings.Contains(output, "info event ***") {
		t.Fatalf("secret was not redacted in useful message: %s", output)
	}
	if strings.Contains(output, canary) {
		t.Fatal("diagnostics exposed secret")
	}
}
