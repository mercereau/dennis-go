package store

import (
	"testing"

	"github.com/jmercereau/dns/config"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

var testCfg = &config.Config{
	Server: config.ServerConfig{
		Listen:         ":5353",
		Upstreams:      []string{"1.1.1.1:53", "8.8.8.8:53"},
		DefaultProfile: "default",
	},
	Profiles: []config.Profile{
		{Name: "default"},
		{Name: "kids", Block: []string{"**.tiktok.com", "**.youtube.com"}},
		{Name: "iot", AllowOnly: []string{"**.amazonaws.com"}},
	},
	Devices: []config.Device{
		{MAC: "AA:BB:CC:DD:EE:FF", Name: "tablet", Profile: "kids"},
		{MAC: "11:22:33:44:55:66", Name: "sensor", Profile: "iot"},
	},
}

func TestSeedAndQuery(t *testing.T) {
	s := openTestStore(t)
	if err := s.Seed(testCfg); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if got := s.Listen(); got != ":5353" {
		t.Errorf("Listen = %q, want :5353", got)
	}

	ups := s.Upstreams()
	if len(ups) != 2 || ups[0] != "1.1.1.1:53" {
		t.Errorf("Upstreams = %v", ups)
	}
}

func TestDeviceFor(t *testing.T) {
	s := openTestStore(t)
	s.Seed(testCfg)

	d := s.DeviceFor("aa:bb:cc:dd:ee:ff") // lowercase
	if d == nil || d.Name != "tablet" {
		t.Fatalf("DeviceFor tablet: %+v", d)
	}

	if s.DeviceFor("ff:ff:ff:ff:ff:ff") != nil {
		t.Error("expected nil for unknown MAC")
	}
}

func TestProfileFor(t *testing.T) {
	s := openTestStore(t)
	s.Seed(testCfg)

	tests := []struct {
		mac         string
		wantProfile string
		wantBlock   []string
	}{
		{"aa:bb:cc:dd:ee:ff", "kids", []string{"**.tiktok.com", "**.youtube.com"}},
		{"11:22:33:44:55:66", "iot", nil},
		{"ff:ff:ff:ff:ff:ff", "default", nil}, // unknown → default
	}

	for _, tt := range tests {
		p := s.ProfileFor(tt.mac)
		if p == nil {
			t.Errorf("ProfileFor(%q) = nil, want %q", tt.mac, tt.wantProfile)
			continue
		}
		if p.Name != tt.wantProfile {
			t.Errorf("ProfileFor(%q).Name = %q, want %q", tt.mac, p.Name, tt.wantProfile)
		}
		if tt.wantBlock != nil && len(p.Block) != len(tt.wantBlock) {
			t.Errorf("ProfileFor(%q).Block = %v, want %v", tt.mac, p.Block, tt.wantBlock)
		}
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	s := openTestStore(t)
	s.Seed(testCfg)
	if err := s.Seed(testCfg); err != nil {
		t.Errorf("second seed failed: %v", err)
	}
	ups := s.Upstreams()
	if len(ups) != 2 {
		t.Errorf("expected 2 upstreams after re-seed, got %d", len(ups))
	}
}
