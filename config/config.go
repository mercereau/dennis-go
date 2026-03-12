package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig    `yaml:"server"`
	Profiles []Profile       `yaml:"profiles"`
	Devices  []Device        `yaml:"devices"`
}

type ServerConfig struct {
	Listen         string   `yaml:"listen"`
	Upstreams      []string `yaml:"upstreams"`
	DefaultProfile string   `yaml:"default_profile"`
}

// Profile defines filtering rules for a group of devices.
type Profile struct {
	Name      string   `yaml:"name"`
	Block     []string `yaml:"block"`      // domain glob patterns to block
	AllowOnly []string `yaml:"allow_only"` // if set, only these patterns are allowed
}

// Device maps a MAC address to a profile.
type Device struct {
	MAC     string `yaml:"mac"`
	Name    string `yaml:"name"`
	Profile string `yaml:"profile"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.normalize()
	return &cfg, nil
}

func (c *Config) normalize() {
	// Normalize MAC addresses to lowercase
	for i := range c.Devices {
		c.Devices[i].MAC = strings.ToLower(c.Devices[i].MAC)
	}
}

// ProfileFor returns the profile for a given MAC address.
// Falls back to the default profile if the MAC is not found.
func (c *Config) ProfileFor(mac string) *Profile {
	mac = strings.ToLower(mac)
	for _, d := range c.Devices {
		if d.MAC == mac {
			return c.findProfile(d.Profile)
		}
	}
	if c.Server.DefaultProfile != "" {
		return c.findProfile(c.Server.DefaultProfile)
	}
	return nil
}

// DeviceFor returns the device config for a given MAC address, or nil if unknown.
func (c *Config) DeviceFor(mac string) *Device {
	mac = strings.ToLower(mac)
	for i := range c.Devices {
		if c.Devices[i].MAC == mac {
			return &c.Devices[i]
		}
	}
	return nil
}

func (c *Config) findProfile(name string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i]
		}
	}
	return nil
}
