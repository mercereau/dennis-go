package store

import "time"

// LogEntry represents a single DNS query event.
type LogEntry struct {
	ID       int64     `json:"id"`
	Time     time.Time `json:"time"`
	ClientIP string    `json:"client_ip"`
	MAC      string    `json:"mac"`
	Device   string    `json:"device"`
	Profile  string    `json:"profile"`
	Domain   string    `json:"domain"`
	Type     string    `json:"type"`
	Action   string    `json:"action"` // ALLOW, BLOCK, ERROR
	RCode    string    `json:"rcode"`
}

// WriteLog inserts a DNS query log entry. Called asynchronously from the DNS handler.
func (s *Store) WriteLog(e LogEntry) {
	s.db.Exec(`
		INSERT INTO dns_logs(ts, client_ip, mac, device, profile, domain, type, action, rcode)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		e.Time.Unix(), e.ClientIP, e.MAC, e.Device, e.Profile, e.Domain, e.Type, e.Action, e.RCode,
	)
}

// ListLogs returns the most recent DNS log entries, newest first.
func (s *Store) ListLogs(limit int) ([]LogEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, ts, client_ip, mac, device, profile, domain, type, action, rcode
		FROM dns_logs
		ORDER BY id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		var ts int64
		rows.Scan(&e.ID, &ts, &e.ClientIP, &e.MAC, &e.Device, &e.Profile, &e.Domain, &e.Type, &e.Action, &e.RCode)
		e.Time = time.Unix(ts, 0)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// SeenDevice is a unique client seen in DNS query logs.
type SeenDevice struct {
	MAC        string    `json:"mac"`
	ClientIP   string    `json:"client_ip"`
	LastSeen   time.Time `json:"last_seen"`
	QueryCount int       `json:"query_count"`
	Registered bool      `json:"registered"`
	Name       string    `json:"name"`
	Profile    string    `json:"profile"`
}

// SeenDevices returns all unique MACs seen in the log, annotated with registration status.
func (s *Store) SeenDevices() ([]SeenDevice, error) {
	rows, err := s.db.Query(`
		SELECT
			l.mac,
			l.client_ip,
			MAX(l.ts)      AS last_seen,
			COUNT(*)       AS query_count,
			COALESCE(d.name, '')    AS name,
			COALESCE(p.name, '')   AS profile
		FROM dns_logs l
		LEFT JOIN devices d ON d.mac = l.mac AND l.mac != ''
		LEFT JOIN profiles p ON p.id = d.profile_id
		WHERE l.mac != ''
		GROUP BY l.mac
		ORDER BY last_seen DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seen []SeenDevice
	for rows.Next() {
		var sd SeenDevice
		var ts int64
		rows.Scan(&sd.MAC, &sd.ClientIP, &ts, &sd.QueryCount, &sd.Name, &sd.Profile)
		sd.LastSeen = time.Unix(ts, 0)
		sd.Registered = sd.Name != ""
		seen = append(seen, sd)
	}
	return seen, rows.Err()
}
