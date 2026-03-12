// Package filter implements domain filtering based on glob-style patterns.
package filter

import (
	"path"
	"strings"
)

// Action represents what to do with a DNS query.
type Action int

const (
	Allow Action = iota
	Block
)

// Decide returns whether the domain should be blocked or allowed given a profile's rules.
// Profile can be nil (meaning no filtering — allow everything).
func Decide(domain string, blockPatterns, allowOnlyPatterns []string) Action {
	domain = strings.TrimSuffix(strings.ToLower(domain), ".")

	if len(allowOnlyPatterns) > 0 {
		for _, p := range allowOnlyPatterns {
			if matchPattern(p, domain) {
				return Allow
			}
		}
		return Block
	}

	for _, p := range blockPatterns {
		if matchPattern(p, domain) {
			return Block
		}
	}
	return Allow
}

// matchPattern checks if domain matches a glob-style pattern.
// Patterns support:
//   - Exact match:     "example.com"
//   - Wildcard prefix: "*.example.com" (matches sub.example.com but not example.com itself)
//   - Double wildcard: "**.example.com" (matches example.com and any subdomain)
func matchPattern(pattern, domain string) bool {
	pattern = strings.ToLower(strings.TrimSuffix(pattern, "."))
	if strings.HasPrefix(pattern, "**.") {
		// Match the domain itself and any subdomain
		suffix := pattern[3:]
		return domain == suffix || strings.HasSuffix(domain, "."+suffix)
	}
	ok, _ := path.Match(pattern, domain)
	return ok
}
