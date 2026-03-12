// Package store provides persistent storage for DNS server configuration
// using SQLite. All server settings, profiles, and devices are stored here
// instead of a flat config file.
package store

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// Store is the SQLite-backed configuration and device store.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite database at path and runs migrations.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite allows one writer at a time; WAL enables concurrent reads
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA foreign_keys=ON;
		PRAGMA busy_timeout=5000;

		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS upstreams (
			position INTEGER PRIMARY KEY,
			address  TEXT NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS profiles (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS rules (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
			type       TEXT NOT NULL CHECK(type IN ('block', 'allow_only')),
			pattern    TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS devices (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			mac        TEXT NOT NULL UNIQUE,
			name       TEXT NOT NULL DEFAULT '',
			profile_id INTEGER REFERENCES profiles(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS dns_logs (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			ts        INTEGER NOT NULL DEFAULT (unixepoch()),
			client_ip TEXT NOT NULL DEFAULT '',
			mac       TEXT NOT NULL DEFAULT '',
			device    TEXT NOT NULL DEFAULT '',
			profile   TEXT NOT NULL DEFAULT '',
			domain    TEXT NOT NULL DEFAULT '',
			type      TEXT NOT NULL DEFAULT '',
			action    TEXT NOT NULL DEFAULT '',
			rcode     TEXT NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS dns_logs_ts  ON dns_logs(ts);
		CREATE INDEX IF NOT EXISTS dns_logs_mac ON dns_logs(mac);
	`)
	return err
}

// Listen returns the listen address (default ":53").
func (s *Store) Listen() string {
	return s.setting("listen", ":53")
}

// Upstreams returns the list of upstream DNS servers in priority order.
func (s *Store) Upstreams() []string {
	rows, err := s.db.Query(`SELECT address FROM upstreams ORDER BY position`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var addrs []string
	for rows.Next() {
		var addr string
		rows.Scan(&addr)
		addrs = append(addrs, addr)
	}
	return addrs
}

// DefaultProfile returns the profile name applied to unknown devices.
func (s *Store) DefaultProfile() string {
	return s.setting("default_profile", "")
}

// Device holds the resolved device info for a MAC address.
type Device struct {
	MAC         string
	Name        string
	ProfileName string
}

// Profile holds the filtering rules for a profile.
type Profile struct {
	Name      string
	Block     []string
	AllowOnly []string
}

// DeviceFor returns the device record for a MAC address, or nil if unknown.
func (s *Store) DeviceFor(mac string) *Device {
	mac = strings.ToLower(mac)
	row := s.db.QueryRow(`
		SELECT d.mac, d.name, COALESCE(p.name, '')
		FROM devices d
		LEFT JOIN profiles p ON p.id = d.profile_id
		WHERE d.mac = ?`, mac)
	var d Device
	if err := row.Scan(&d.MAC, &d.Name, &d.ProfileName); err != nil {
		return nil
	}
	return &d
}

// ProfileFor returns the filtering profile for a MAC address.
// Falls back to the default profile for unknown devices.
func (s *Store) ProfileFor(mac string) *Profile {
	mac = strings.ToLower(mac)

	// Try to find the device's assigned profile, then fall back to default.
	row := s.db.QueryRow(`
		SELECT p.id, p.name
		FROM devices d
		JOIN profiles p ON p.id = d.profile_id
		WHERE d.mac = ?`, mac)

	var profileID int64
	var profileName string
	if err := row.Scan(&profileID, &profileName); err != nil {
		// Unknown device: use default profile
		defaultName := s.DefaultProfile()
		if defaultName == "" {
			return nil
		}
		return s.profileByName(defaultName)
	}
	return s.profileByID(profileID, profileName)
}

func (s *Store) profileByName(name string) *Profile {
	row := s.db.QueryRow(`SELECT id FROM profiles WHERE name = ?`, name)
	var id int64
	if err := row.Scan(&id); err != nil {
		return nil
	}
	return s.profileByID(id, name)
}

func (s *Store) profileByID(id int64, name string) *Profile {
	rows, err := s.db.Query(`SELECT type, pattern FROM rules WHERE profile_id = ?`, id)
	if err != nil {
		return nil
	}
	defer rows.Close()

	p := &Profile{Name: name}
	for rows.Next() {
		var ruleType, pattern string
		rows.Scan(&ruleType, &pattern)
		switch ruleType {
		case "block":
			p.Block = append(p.Block, pattern)
		case "allow_only":
			p.AllowOnly = append(p.AllowOnly, pattern)
		}
	}
	return p
}

func (s *Store) setting(key, defaultVal string) string {
	var val string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&val)
	if err != nil {
		return defaultVal
	}
	return val
}
