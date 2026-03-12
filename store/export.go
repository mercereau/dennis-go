package store

import (
	"github.com/jmercereau/dns/config"
)

// Export reads the current database state and returns it as a Config struct,
// which can be marshalled to YAML to produce a config.yaml snapshot.
func (s *Store) Export() (*config.Config, error) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Listen:         s.Listen(),
			DefaultProfile: s.DefaultProfile(),
			Upstreams:      s.Upstreams(),
		},
	}

	profiles, err := s.ListProfiles()
	if err != nil {
		return nil, err
	}
	for _, p := range profiles {
		cfg.Profiles = append(cfg.Profiles, config.Profile{
			Name:      p.Name,
			Block:     p.Block,
			AllowOnly: p.AllowOnly,
		})
	}

	devices, err := s.ListDevices()
	if err != nil {
		return nil, err
	}
	for _, d := range devices {
		cfg.Devices = append(cfg.Devices, config.Device{
			MAC:     d.MAC,
			Name:    d.Name,
			Profile: d.Profile,
		})
	}

	groups, err := s.ListDeviceGroups()
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		dg := config.DeviceGroup{
			Name:    g.Name,
			Profile: g.Profile,
			Devices: g.Devices,
		}
		for _, s := range g.Schedules {
			dg.Schedules = append(dg.Schedules, config.Schedule{
				Profile: s.Profile,
				Start:   s.Start,
				End:     s.End,
			})
		}
		cfg.DeviceGroups = append(cfg.DeviceGroups, dg)
	}

	return cfg, nil
}
