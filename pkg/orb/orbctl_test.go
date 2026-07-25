package orb

import "testing"

func TestImage(t *testing.T) {
	const ubuntu = "ubuntu"
	cases := []struct {
		distro, version, want string
	}{
		{ubuntu, "", ubuntu},
		{ubuntu, "24.04", "ubuntu:24.04"},
		{"fedora", "43", "fedora:43"},
	}
	for _, tc := range cases {
		if got := image(tc.distro, tc.version); got != tc.want {
			t.Errorf("image(%q,%q) = %q, want %q", tc.distro, tc.version, got, tc.want)
		}
	}
}

func TestDockerProvisionScriptQuotesUser(t *testing.T) {
	script := dockerProvisionScript("dev user")
	if !contains(script, `usermod -aG docker "dev user"`) {
		t.Errorf("expected quoted usermod for the machine user, got:\n%s", script)
	}
	if !contains(script, "get.docker.com") {
		t.Error("provision script should install docker via get.docker.com")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
