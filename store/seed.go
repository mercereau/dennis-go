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
	for _, table := range []string{"devices", "rules", "profiles", "upstreams", "settings"} {
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
		profileID, ok := profileIDs[d.Profile]
		if !ok {
			return fmt.Errorf("device %s references unknown profile %q", d.Name, d.Profile)
		}
		if _, err := tx.Exec(`INSERT INTO devices(mac, name, profile_id) VALUES(?,?,?)`, mac, d.Name, profileID); err != nil {
			return fmt.Errorf("device %s: %w", d.Name, err)
		}
	}

	return tx.Commit()
}
