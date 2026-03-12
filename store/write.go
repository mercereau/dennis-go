package store

import (
	"database/sql"
	"fmt"
	"strings"
)

// --- Settings ---

func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(`INSERT INTO settings(key, value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// --- Upstreams ---

func (s *Store) SetUpstreams(addrs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM upstreams`); err != nil {
		return err
	}
	for i, addr := range addrs {
		if _, err := tx.Exec(`INSERT INTO upstreams(position, address) VALUES(?,?)`, i, addr); err != nil {
			return fmt.Errorf("inserting upstream %s: %w", addr, err)
		}
	}
	return tx.Commit()
}

// --- Profiles ---

// ProfileRow is the full profile data returned by list/get operations.
type ProfileRow struct {
	Name      string   `json:"name"`
	Block     []string `json:"block"`
	AllowOnly []string `json:"allow_only"`
}

func (s *Store) ListProfiles() ([]ProfileRow, error) {
	rows, err := s.db.Query(`SELECT id, name FROM profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}

	// Collect IDs before closing rows — opening a second query while rows is
	// still open deadlocks when MaxOpenConns=1 (the single connection is held
	// by the open cursor).
	type meta struct{ id int64; name string }
	var metas []meta
	for rows.Next() {
		var m meta
		rows.Scan(&m.id, &m.name)
		metas = append(metas, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var profiles []ProfileRow
	for _, m := range metas {
		p, err := s.profileByIDFull(m.id, m.name)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, *p)
	}
	return profiles, nil
}

func (s *Store) GetProfile(name string) (*ProfileRow, error) {
	row := s.db.QueryRow(`SELECT id FROM profiles WHERE name = ?`, name)
	var id int64
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return s.profileByIDFull(id, name)
}

func (s *Store) CreateProfile(name string, block, allowOnly []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO profiles(name) VALUES(?)`, name)
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	id, _ := res.LastInsertId()
	if err := insertRules(tx, id, block, allowOnly); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpdateProfile(name string, block, allowOnly []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int64
	if err := tx.QueryRow(`SELECT id FROM profiles WHERE name = ?`, name).Scan(&id); err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM rules WHERE profile_id = ?`, id); err != nil {
		return err
	}
	if err := insertRules(tx, id, block, allowOnly); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) DeleteProfile(name string) error {
	res, err := s.db.Exec(`DELETE FROM profiles WHERE name = ?`, name)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("profile %q not found", name)
	}
	return nil
}

// --- Devices ---

// DeviceRow is the full device data returned by list/get operations.
type DeviceRow struct {
	MAC     string `json:"mac"`
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

func (s *Store) ListDevices() ([]DeviceRow, error) {
	rows, err := s.db.Query(`
		SELECT d.mac, d.name, COALESCE(p.name, '')
		FROM devices d
		LEFT JOIN profiles p ON p.id = d.profile_id
		ORDER BY d.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []DeviceRow
	for rows.Next() {
		var d DeviceRow
		rows.Scan(&d.MAC, &d.Name, &d.Profile)
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *Store) GetDevice(mac string) (*DeviceRow, error) {
	mac = strings.ToLower(mac)
	row := s.db.QueryRow(`
		SELECT d.mac, d.name, COALESCE(p.name, '')
		FROM devices d
		LEFT JOIN profiles p ON p.id = d.profile_id
		WHERE d.mac = ?`, mac)
	var d DeviceRow
	if err := row.Scan(&d.MAC, &d.Name, &d.Profile); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) CreateDevice(mac, name, profileName string) error {
	mac = strings.ToLower(mac)
	var profileID *int64
	if profileName != "" {
		var id int64
		if err := s.db.QueryRow(`SELECT id FROM profiles WHERE name = ?`, profileName).Scan(&id); err != nil {
			return fmt.Errorf("profile %q not found: %w", profileName, err)
		}
		profileID = &id
	}
	_, err := s.db.Exec(`INSERT INTO devices(mac, name, profile_id) VALUES(?,?,?)`, mac, name, profileID)
	return err
}

func (s *Store) UpdateDevice(mac, name, profileName string) error {
	mac = strings.ToLower(mac)
	var profileID *int64
	if profileName != "" {
		var id int64
		if err := s.db.QueryRow(`SELECT id FROM profiles WHERE name = ?`, profileName).Scan(&id); err != nil {
			return fmt.Errorf("profile %q not found: %w", profileName, err)
		}
		profileID = &id
	}
	res, err := s.db.Exec(`UPDATE devices SET name=?, profile_id=? WHERE mac=?`, name, profileID, mac)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("device %q not found", mac)
	}
	return nil
}

func (s *Store) DeleteDevice(mac string) error {
	mac = strings.ToLower(mac)
	res, err := s.db.Exec(`DELETE FROM devices WHERE mac = ?`, mac)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("device %q not found", mac)
	}
	return nil
}

// --- helpers ---

type execer interface {
	Exec(string, ...any) (interface{ LastInsertId() (int64, error); RowsAffected() (int64, error) }, error)
}

func insertRules(tx *sql.Tx, profileID int64, block, allowOnly []string) error {
	for _, pat := range block {
		if _, err := tx.Exec(`INSERT INTO rules(profile_id, type, pattern) VALUES(?,?,?)`, profileID, "block", pat); err != nil {
			return err
		}
	}
	for _, pat := range allowOnly {
		if _, err := tx.Exec(`INSERT INTO rules(profile_id, type, pattern) VALUES(?,?,?)`, profileID, "allow_only", pat); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) profileByIDFull(id int64, name string) (*ProfileRow, error) {
	rows, err := s.db.Query(`SELECT type, pattern FROM rules WHERE profile_id = ? ORDER BY type, pattern`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	p := &ProfileRow{Name: name, Block: []string{}, AllowOnly: []string{}}
	for rows.Next() {
		var rtype, pattern string
		rows.Scan(&rtype, &pattern)
		switch rtype {
		case "block":
			p.Block = append(p.Block, pattern)
		case "allow_only":
			p.AllowOnly = append(p.AllowOnly, pattern)
		}
	}
	return p, rows.Err()
}
