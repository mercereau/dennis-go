package config

import (
	"os"
	"testing"
)

const testConfig = `
server:
  listen: ":5353"
  upstream: "8.8.8.8:53"
  default_profile: "default"

profiles:
  - name: default
  - name: kids
    block:
      - "**.tiktok.com"
  - name: iot
    allow_only:
      - "**.amazonaws.com"

devices:
  - mac: "AA:BB:CC:DD:EE:FF"
    name: "tablet"
    profile: "kids"
  - mac: "11:22:33:44:55:66"
    name: "sensor"
    profile: "iot"
`

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestLoad(t *testing.T) {
	cfg, err := Load(writeTemp(t, testConfig))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Listen != ":5353" {
		t.Errorf("listen = %q, want :5353", cfg.Server.Listen)
	}
	if len(cfg.Profiles) != 3 {
		t.Errorf("profiles count = %d, want 3", len(cfg.Profiles))
	}
}

func TestProfileFor(t *testing.T) {
	cfg, err := Load(writeTemp(t, testConfig))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		mac         string
		wantProfile string
	}{
		{"aa:bb:cc:dd:ee:ff", "kids"},       // lowercase version of config MAC
		{"AA:BB:CC:DD:EE:FF", "kids"},       // case-insensitive
		{"11:22:33:44:55:66", "iot"},
		{"ff:ff:ff:ff:ff:ff", "default"},    // unknown device → default
	}

	for _, tt := range tests {
		p := cfg.ProfileFor(tt.mac)
		if p == nil {
			t.Errorf("ProfileFor(%q) = nil, want %q", tt.mac, tt.wantProfile)
			continue
		}
		if p.Name != tt.wantProfile {
			t.Errorf("ProfileFor(%q) = %q, want %q", tt.mac, p.Name, tt.wantProfile)
		}
	}
}
