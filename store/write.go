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

// --- Device Groups ---

// ScheduleRow is a single time-window schedule entry within a device group.
type ScheduleRow struct {
	Profile string `json:"profile"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

// DeviceGroupRow is the full device group data returned by list/get operations.
type DeviceGroupRow struct {
	Name      string        `json:"name"`
	Profile   string        `json:"profile"`
	Devices   []string      `json:"devices"`
	Schedules []ScheduleRow `json:"schedules"`
}

func (s *Store) ListDeviceGroups() ([]DeviceGroupRow, error) {
	rows, err := s.db.Query(`
		SELECT g.name, COALESCE(p.name, '')
		FROM device_groups g
		LEFT JOIN profiles p ON p.id = g.profile_id
		ORDER BY g.name`)
	if err != nil {
		return nil, err
	}
	type meta struct{ name, profile string }
	var metas []meta
	for rows.Next() {
		var m meta
		rows.Scan(&m.name, &m.profile)
		metas = append(metas, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var groups []DeviceGroupRow
	for _, m := range metas {
		g, err := s.deviceGroupFull(m.name, m.profile)
		if err != nil {
			return nil, err
		}
		groups = append(groups, *g)
	}
	return groups, nil
}

func (s *Store) GetDeviceGroup(name string) (*DeviceGroupRow, error) {
	row := s.db.QueryRow(`
		SELECT g.name, COALESCE(p.name, '')
		FROM device_groups g
		LEFT JOIN profiles p ON p.id = g.profile_id
		WHERE g.name = ?`, name)
	var gname, profile string
	if err := row.Scan(&gname, &profile); err != nil {
		return nil, err
	}
	return s.deviceGroupFull(gname, profile)
}

func (s *Store) deviceGroupFull(name, profile string) (*DeviceGroupRow, error) {
	// Fetch members — close rows before opening the next query (MaxOpenConns=1).
	rows, err := s.db.Query(`
		SELECT m.mac
		FROM device_group_members m
		JOIN device_groups g ON g.id = m.group_id
		WHERE g.name = ?
		ORDER BY m.mac`, name)
	if err != nil {
		return nil, err
	}
	g := &DeviceGroupRow{Name: name, Profile: profile, Devices: []string{}, Schedules: []ScheduleRow{}}
	for rows.Next() {
		var mac string
		rows.Scan(&mac)
		g.Devices = append(g.Devices, mac)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch schedules.
	rows, err = s.db.Query(`
		SELECT p.name, s.start_time, s.end_time
		FROM device_group_schedules s
		JOIN device_groups g ON g.id = s.group_id
		JOIN profiles p ON p.id = s.profile_id
		WHERE g.name = ?
		ORDER BY s.start_time`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sr ScheduleRow
		rows.Scan(&sr.Profile, &sr.Start, &sr.End)
		g.Schedules = append(g.Schedules, sr)
	}
	return g, rows.Err()
}

func (s *Store) CreateDeviceGroup(name, profile string, devices []string, schedules []ScheduleRow) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var profileID int64
	if err := tx.QueryRow(`SELECT id FROM profiles WHERE name = ?`, profile).Scan(&profileID); err != nil {
		return fmt.Errorf("profile %q not found: %w", profile, err)
	}
	res, err := tx.Exec(`INSERT INTO device_groups(name, profile_id) VALUES(?,?)`, name, profileID)
	if err != nil {
		return fmt.Errorf("create device group: %w", err)
	}
	groupID, _ := res.LastInsertId()
	for _, mac := range devices {
		mac = strings.ToLower(mac)
		if _, err := tx.Exec(`INSERT INTO device_group_members(group_id, mac) VALUES(?,?)`, groupID, mac); err != nil {
			return fmt.Errorf("adding device %s: %w", mac, err)
		}
	}
	if err := insertSchedules(tx, groupID, schedules); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpdateDeviceGroup(name, profile string, devices []string, schedules []ScheduleRow) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var groupID int64
	if err := tx.QueryRow(`SELECT id FROM device_groups WHERE name = ?`, name).Scan(&groupID); err != nil {
		return fmt.Errorf("device group %q not found: %w", name, err)
	}
	var profileID int64
	if err := tx.QueryRow(`SELECT id FROM profiles WHERE name = ?`, profile).Scan(&profileID); err != nil {
		return fmt.Errorf("profile %q not found: %w", profile, err)
	}
	if _, err := tx.Exec(`UPDATE device_groups SET profile_id = ? WHERE id = ?`, profileID, groupID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM device_group_members WHERE group_id = ?`, groupID); err != nil {
		return err
	}
	for _, mac := range devices {
		mac = strings.ToLower(mac)
		if _, err := tx.Exec(`INSERT INTO device_group_members(group_id, mac) VALUES(?,?)`, groupID, mac); err != nil {
			return fmt.Errorf("adding device %s: %w", mac, err)
		}
	}
	if _, err := tx.Exec(`DELETE FROM device_group_schedules WHERE group_id = ?`, groupID); err != nil {
		return err
	}
	if err := insertSchedules(tx, groupID, schedules); err != nil {
		return err
	}
	return tx.Commit()
}

func insertSchedules(tx *sql.Tx, groupID int64, schedules []ScheduleRow) error {
	for _, s := range schedules {
		var profileID int64
		if err := tx.QueryRow(`SELECT id FROM profiles WHERE name = ?`, s.Profile).Scan(&profileID); err != nil {
			return fmt.Errorf("schedule profile %q not found: %w", s.Profile, err)
		}
		if _, err := tx.Exec(`INSERT INTO device_group_schedules(group_id, profile_id, start_time, end_time) VALUES(?,?,?,?)`,
			groupID, profileID, s.Start, s.End); err != nil {
			return fmt.Errorf("adding schedule %s-%s: %w", s.Start, s.End, err)
		}
	}
	return nil
}

func (s *Store) DeleteDeviceGroup(name string) error {
	res, err := s.db.Exec(`DELETE FROM device_groups WHERE name = ?`, name)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("device group %q not found", name)
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
