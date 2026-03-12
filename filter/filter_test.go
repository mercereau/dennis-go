package filter

import "testing"

func TestDecide(t *testing.T) {
	tests := []struct {
		name         string
		domain       string
		block        []string
		allowOnly    []string
		wantAction   Action
	}{
		// Block patterns
		{
			name:       "no rules — allow",
			domain:     "example.com",
			wantAction: Allow,
		},
		{
			name:       "exact block match",
			domain:     "tiktok.com",
			block:      []string{"tiktok.com"},
			wantAction: Block,
		},
		{
			name:       "wildcard subdomain block",
			domain:     "www.tiktok.com",
			block:      []string{"*.tiktok.com"},
			wantAction: Block,
		},
		{
			name:       "wildcard does not match apex",
			domain:     "tiktok.com",
			block:      []string{"*.tiktok.com"},
			wantAction: Allow,
		},
		{
			name:       "double wildcard matches apex",
			domain:     "tiktok.com",
			block:      []string{"**.tiktok.com"},
			wantAction: Block,
		},
		{
			name:       "double wildcard matches subdomain",
			domain:     "cdn.tiktok.com",
			block:      []string{"**.tiktok.com"},
			wantAction: Block,
		},
		{
			name:       "double wildcard matches deep subdomain",
			domain:     "a.b.tiktok.com",
			block:      []string{"**.tiktok.com"},
			wantAction: Block,
		},
		{
			name:       "trailing dot in DNS name",
			domain:     "tiktok.com.",
			block:      []string{"**.tiktok.com"},
			wantAction: Block,
		},
		{
			name:       "unrelated domain not blocked",
			domain:     "example.com",
			block:      []string{"**.tiktok.com"},
			wantAction: Allow,
		},
		// Allow-only patterns
		{
			name:       "allow_only — matching domain allowed",
			domain:     "api.amazonaws.com",
			allowOnly:  []string{"**.amazonaws.com"},
			wantAction: Allow,
		},
		{
			name:       "allow_only — non-matching domain blocked",
			domain:     "example.com",
			allowOnly:  []string{"**.amazonaws.com"},
			wantAction: Block,
		},
		{
			name:       "allow_only takes precedence over block",
			domain:     "api.amazonaws.com",
			block:      []string{"**.amazonaws.com"},
			allowOnly:  []string{"**.amazonaws.com"},
			wantAction: Allow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decide(tt.domain, tt.block, tt.allowOnly)
			if got != tt.wantAction {
				t.Errorf("Decide(%q) = %v, want %v", tt.domain, got, tt.wantAction)
			}
		})
	}
}
