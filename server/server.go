package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/miekg/dns"

	"github.com/jmercereau/dns/arp"
	"github.com/jmercereau/dns/store"
)

const arpCacheTTL = 30 * time.Second

type Server struct {
	store *store.Store
	log   *slog.Logger
}

func New(store *store.Store, log *slog.Logger) *Server {
	return &Server{store: store, log: log}
}

func (s *Server) Run(ctx context.Context) error {
	arpTable := arp.NewTable(arpCacheTTL)

	h := &handler{
		store: s.store,
		arp:   arpTable,
		log:   s.log,
	}

	mux := dns.NewServeMux()
	mux.HandleFunc(".", h.ServeDNS)

	listen := s.store.Listen()
	udp := &dns.Server{Addr: listen, Net: "udp", Handler: mux}
	tcp := &dns.Server{Addr: listen, Net: "tcp", Handler: mux}

	errCh := make(chan error, 2)
	go func() { errCh <- udp.ListenAndServe() }()
	go func() { errCh <- tcp.ListenAndServe() }()

	s.log.Info("DNS server started",
		slog.String("listen", listen),
		slog.Any("upstreams", s.store.Upstreams()),
	)

	select {
	case <-ctx.Done():
		udp.Shutdown()
		tcp.Shutdown()
		return nil
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
}
