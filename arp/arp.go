// Package arp provides IP-to-MAC address resolution by reading the
// kernel's ARP table. This only works for hosts on the same L2 segment
// (direct LAN, no NAT in between).
package arp

import (
	"strings"
	"sync"
	"time"
)

// Table is a cached view of the system's ARP table.
type Table struct {
	mu      sync.RWMutex
	entries map[string]string // IP → MAC
	ttl     time.Duration
	updated time.Time
}

func NewTable(ttl time.Duration) *Table {
	return &Table{
		entries: make(map[string]string),
		ttl:     ttl,
	}
}

// Lookup returns the MAC address for the given IP, refreshing the cache if stale.
// Returns ("", false) if the IP is not in the ARP table.
func (t *Table) Lookup(ip string) (string, bool) {
	t.mu.Lock()
	if time.Since(t.updated) > t.ttl {
		t.refresh()
	}
	t.mu.Unlock()

	t.mu.RLock()
	mac, ok := t.entries[ip]
	t.mu.RUnlock()
	return mac, ok
}

// refresh reads the platform ARP table and updates the cache. Caller must hold t.mu (write).
func (t *Table) refresh() {
	entries, err := readARPTable()
	if err != nil {
		return
	}
	t.entries = entries
	t.updated = time.Now()
}

func normalizeMac(mac string) string {
	return strings.ToLower(mac)
}
