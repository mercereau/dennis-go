//go:build linux

package arp

import (
	"bufio"
	"os"
	"strings"
)

// readARPTable reads /proc/net/arp.
// Format:
//   IP address       HW type     Flags       HW address            Mask     Device
//   192.168.1.100    0x1         0x2         aa:bb:cc:dd:ee:ff     *        eth0
func readARPTable() (map[string]string, error) {
	f, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries := make(map[string]string)
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		ip := fields[0]
		mac := normalizeMac(fields[3])
		if mac == "00:00:00:00:00:00" || fields[2] == "0x0" {
			continue
		}
		entries[ip] = mac
	}
	return entries, scanner.Err()
}
