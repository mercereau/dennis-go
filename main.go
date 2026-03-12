package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"

	"github.com/jmercereau/dns/backend"
	"github.com/jmercereau/dns/config"
	"github.com/jmercereau/dns/server"
	"github.com/jmercereau/dns/store"
)

func main() {
	dbPath     := flag.String("db", "dns.db", "path to SQLite database")
	seedPath   := flag.String("seed", "", "seed the database from a YAML config file and exit")
	exportPath := flag.String("export", "", "export the database to a YAML config file and exit")
	apiAddr    := flag.String("api", ":9090", "address for the HTTP API server")
	dnsOnly    := flag.Bool("dns-only", false, "run only the DNS server, not the HTTP API")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Error("failed to open database", slog.Any("err", err))
		os.Exit(1)
	}
	defer st.Close()

	if *seedPath != "" {
		cfg, err := config.Load(*seedPath)
		if err != nil {
			log.Error("failed to load seed config", slog.Any("err", err))
			os.Exit(1)
		}
		if err := st.Seed(cfg); err != nil {
			log.Error("failed to seed database", slog.Any("err", err))
			os.Exit(1)
		}
		log.Info("database seeded", slog.String("from", *seedPath), slog.String("db", *dbPath))
		return
	}

	if *exportPath != "" {
		cfg, err := st.Export()
		if err != nil {
			log.Error("failed to export database", slog.Any("err", err))
			os.Exit(1)
		}
		f, err := os.Create(*exportPath)
		if err != nil {
			log.Error("failed to create export file", slog.Any("err", err))
			os.Exit(1)
		}
		defer f.Close()
		if err := yaml.NewEncoder(f).Encode(cfg); err != nil {
			log.Error("failed to write export", slog.Any("err", err))
			os.Exit(1)
		}
		log.Info("database exported", slog.String("to", *exportPath))
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 2)

	s := server.New(st, log)
	go func() { errCh <- s.Run(ctx) }()

	if !*dnsOnly {
		api := backend.New(st, log)
		go func() { errCh <- api.Run(ctx, *apiAddr) }()
	}

	if err := <-errCh; err != nil {
		log.Error("fatal server error", slog.Any("err", err))
		cancel()
		os.Exit(1)
	}
}
