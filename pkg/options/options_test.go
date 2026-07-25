package options

import "testing"

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("MACHINE_ID", "ws1")
	for _, k := range []string{
		"ORBSTACK_PATH", "ORBSTACK_DISTRO", "ORBSTACK_VERSION", "ORBSTACK_ARCH",
		"ORBSTACK_CPUS", "ORBSTACK_MEMORY", "ORBSTACK_DISK", "ORBSTACK_ISOLATED",
	} {
		t.Setenv(k, "")
	}

	o, err := FromEnv(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.OrbctlPath != "orbctl" {
		t.Errorf("OrbctlPath = %q, want orbctl", o.OrbctlPath)
	}
	if o.Distro != "ubuntu" {
		t.Errorf("Distro = %q, want ubuntu", o.Distro)
	}
	if !o.Isolated {
		t.Error("Isolated should default to true")
	}
}

func TestIsolatedFalse(t *testing.T) {
	t.Setenv("MACHINE_ID", "ws1")
	t.Setenv("ORBSTACK_ISOLATED", "false")
	o, err := FromEnv(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Isolated {
		t.Error("Isolated should be false when ORBSTACK_ISOLATED=false")
	}
}

func TestRequireMachineID(t *testing.T) {
	t.Setenv("MACHINE_ID", "")
	if _, err := FromEnv(true); err == nil {
		t.Error("expected error when MACHINE_ID is missing")
	}
	if _, err := FromEnv(false); err != nil {
		t.Errorf("unexpected error when MACHINE_ID not required: %v", err)
	}
}
