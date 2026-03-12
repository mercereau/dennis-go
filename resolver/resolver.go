// Package resolver forwards DNS queries to an upstream server.
package resolver

import (
	"errors"
	"fmt"

	"github.com/miekg/dns"
)

// Forward tries each upstream in order, returning the first successful response.
func Forward(req *dns.Msg, upstreams []string) (*dns.Msg, error) {
	c := &dns.Client{}
	var errs []error
	for _, upstream := range upstreams {
		resp, _, err := c.Exchange(req, upstream)
		if err == nil {
			return resp, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", upstream, err))
	}
	return nil, fmt.Errorf("all upstreams failed: %w", errors.Join(errs...))
}
