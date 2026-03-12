//go:build darwin

package arp

import (
	"os/exec"
	"strings"
)

// readARPTable parses the output of `arp -a`.
// Example line:
//   ? (192.168.1.5) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]
func readARPTable() (map[string]string, error) {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil, err
	}

	entries := make(map[string]string)
	for _, line := range strings.Split(string(out), "\n") {
		// Extract IP: between '(' and ')'
		ipStart := strings.Index(line, "(")
		ipEnd := strings.Index(line, ")")
		if ipStart < 0 || ipEnd <= ipStart {
			continue
		}
		ip := line[ipStart+1 : ipEnd]

		// Extract MAC: after " at "
		atIdx := strings.Index(line, " at ")
		if atIdx < 0 {
			continue
		}
		rest := line[atIdx+4:]
		mac := strings.Fields(rest)[0]
		if mac == "(incomplete)" || mac == "ff:ff:ff:ff:ff:ff" {
			continue
		}
		entries[ip] = normalizeMac(mac)
	}
	return entries, nil
}
