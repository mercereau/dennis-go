package server

import (
	"log/slog"
	"net"
	"time"

	"github.com/miekg/dns"

	"github.com/jmercereau/dns/arp"
	"github.com/jmercereau/dns/filter"
	"github.com/jmercereau/dns/resolver"
	"github.com/jmercereau/dns/store"
)

type handler struct {
	store *store.Store
	arp   *arp.Table
	log   *slog.Logger
}

func (h *handler) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) == 0 {
		dns.HandleFailed(w, req)
		return
	}

	clientIP := clientIP(w.RemoteAddr())
	domain := req.Question[0].Name
	qtype := dns.TypeToString[req.Question[0].Qtype]

	// Resolve client IP → MAC → device → profile
	mac, _ := h.arp.Lookup(clientIP)
	device := h.store.DeviceFor(mac)
	profile := h.store.ProfileFor(mac)

	deviceName, profileName := "", ""
	if device != nil {
		deviceName = device.Name
	}
	if profile != nil {
		profileName = profile.Name
	}

	attrs := []any{
		slog.String("client", clientIP),
		slog.String("domain", domain),
		slog.String("type", qtype),
	}
	if mac != "" {
		attrs = append(attrs, slog.String("mac", mac))
	}
	if deviceName != "" {
		attrs = append(attrs, slog.String("device", deviceName))
	}
	if profileName != "" {
		attrs = append(attrs, slog.String("profile", profileName))
	}

	entry := store.LogEntry{
		Time:     time.Now(),
		ClientIP: clientIP,
		MAC:      mac,
		Device:   deviceName,
		Profile:  profileName,
		Domain:   domain,
		Type:     qtype,
	}

	// Apply filtering
	if profile != nil {
		action := filter.Decide(domain, profile.Block, profile.AllowOnly)
		if action == filter.Block {
			h.log.Info("BLOCK", attrs...)
			entry.Action = "BLOCK"
			go h.store.WriteLog(entry)
			blocked(w, req)
			return
		}
	}

	// Forward to upstream
	resp, err := resolver.Forward(req, h.store.Upstreams())
	if err != nil {
		h.log.Error("ERROR", append(attrs, slog.Any("err", err))...)
		entry.Action = "ERROR"
		go h.store.WriteLog(entry)
		dns.HandleFailed(w, req)
		return
	}

	rcode := dns.RcodeToString[resp.Rcode]
	h.log.Info("ALLOW", append(attrs, slog.String("rcode", rcode))...)
	entry.Action = "ALLOW"
	entry.RCode = rcode
	go h.store.WriteLog(entry)

	resp.SetReply(req)
	w.WriteMsg(resp)
}

// blocked responds with NXDOMAIN.
func blocked(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeNameError)
	m.Authoritative = true
	w.WriteMsg(m)
}

func clientIP(addr net.Addr) string {
	switch a := addr.(type) {
	case *net.UDPAddr:
		return a.IP.String()
	case *net.TCPAddr:
		return a.IP.String()
	}
	host, _, _ := net.SplitHostPort(addr.String())
	return host
}
