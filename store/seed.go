package store

import (
	"fmt"
	"strings"

	"github.com/jmercereau/dns/config"
)

// Seed imports a YAML config into the database, replacing all existing data.
// Use this once to migrate from config.yaml to the database.
func (s *Store) Seed(cfg *config.Config) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing data
	for _, table := range []string{"device_group_schedules", "device_group_members", "device_groups", "devices", "rules", "profiles", "upstreams", "settings"} {
		if _, err := tx.Exec(`DELETE FROM ` + table); err != nil {
			return fmt.Errorf("clearing %s: %w", table, err)
		}
	}

	// Settings
	settings := map[string]string{
		"listen":          cfg.Server.Listen,
		"default_profile": cfg.Server.DefaultProfile,
	}
	for k, v := range settings {
		if _, err := tx.Exec(`INSERT INTO settings(key, value) VALUES(?,?)`, k, v); err != nil {
			return fmt.Errorf("setting %s: %w", k, err)
		}
	}

	// Upstreams
	for i, addr := range cfg.Server.Upstreams {
		if _, err := tx.Exec(`INSERT INTO upstreams(position, address) VALUES(?,?)`, i, addr); err != nil {
			return fmt.Errorf("upstream %s: %w", addr, err)
		}
	}

	// Profiles + rules
	profileIDs := make(map[string]int64)
	for _, p := range cfg.Profiles {
		res, err := tx.Exec(`INSERT INTO profiles(name) VALUES(?)`, p.Name)
		if err != nil {
			return fmt.Errorf("profile %s: %w", p.Name, err)
		}
		id, _ := res.LastInsertId()
		profileIDs[p.Name] = id

		for _, pat := range p.Block {
			if _, err := tx.Exec(`INSERT INTO rules(profile_id, type, pattern) VALUES(?,?,?)`, id, "block", pat); err != nil {
				return err
			}
		}
		for _, pat := range p.AllowOnly {
			if _, err := tx.Exec(`INSERT INTO rules(profile_id, type, pattern) VALUES(?,?,?)`, id, "allow_only", pat); err != nil {
				return err
			}
		}
	}

	// Devices
	for _, d := range cfg.Devices {
		mac := strings.ToLower(d.MAC)
		var profileID *int64
		if d.Profile != "" {
			id, ok := profileIDs[d.Profile]
			if !ok {
				return fmt.Errorf("device %s references unknown profile %q", d.Name, d.Profile)
			}
			profileID = &id
		}
		if _, err := tx.Exec(`INSERT INTO devices(mac, name, profile_id) VALUES(?,?,?)`, mac, d.Name, profileID); err != nil {
			return fmt.Errorf("device %s: %w", d.Name, err)
		}
	}

	// Device Groups
	for _, g := range cfg.DeviceGroups {
		profileID, ok := profileIDs[g.Profile]
		if !ok {
			return fmt.Errorf("device group %s references unknown profile %q", g.Name, g.Profile)
		}
		res, err := tx.Exec(`INSERT INTO device_groups(name, profile_id) VALUES(?,?)`, g.Name, profileID)
		if err != nil {
			return fmt.Errorf("device group %s: %w", g.Name, err)
		}
		groupID, _ := res.LastInsertId()
		for _, mac := range g.Devices {
			mac = strings.ToLower(mac)
			if _, err := tx.Exec(`INSERT INTO device_group_members(group_id, mac) VALUES(?,?)`, groupID, mac); err != nil {
				return fmt.Errorf("device group member %s/%s: %w", g.Name, mac, err)
			}
		}
		for _, sched := range g.Schedules {
			schedProfileID, ok := profileIDs[sched.Profile]
			if !ok {
				return fmt.Errorf("schedule in group %s references unknown profile %q", g.Name, sched.Profile)
			}
			if _, err := tx.Exec(`INSERT INTO device_group_schedules(group_id, profile_id, start_time, end_time) VALUES(?,?,?,?)`,
				groupID, schedProfileID, sched.Start, sched.End); err != nil {
				return fmt.Errorf("schedule %s-%s in group %s: %w", sched.Start, sched.End, g.Name, err)
			}
		}
	}

	return tx.Commit()
}
