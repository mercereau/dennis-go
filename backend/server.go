package backend

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jmercereau/dns/store"
)

type Server struct {
	store *store.Store
	log   *slog.Logger
}

func New(store *store.Store, log *slog.Logger) *Server {
	return &Server{store: store, log: log}
}

// Run starts the HTTP API server.
func (s *Server) Run(ctx context.Context, addr string) error {
	mux := http.NewServeMux()

	// Settings
	mux.HandleFunc("GET /api/settings", s.getSettings)
	mux.HandleFunc("PUT /api/settings", s.putSettings)

	// Upstreams
	mux.HandleFunc("GET /api/upstreams", s.getUpstreams)
	mux.HandleFunc("PUT /api/upstreams", s.putUpstreams)

	// Profiles
	mux.HandleFunc("GET /api/profiles", s.listProfiles)
	mux.HandleFunc("POST /api/profiles", s.createProfile)
	mux.HandleFunc("GET /api/profiles/{name}", s.getProfile)
	mux.HandleFunc("PUT /api/profiles/{name}", s.updateProfile)
	mux.HandleFunc("DELETE /api/profiles/{name}", s.deleteProfile)

	// Devices
	mux.HandleFunc("GET /api/devices", s.listDevices)
	mux.HandleFunc("POST /api/devices", s.createDevice)
	mux.HandleFunc("GET /api/devices/{mac}", s.getDevice)
	mux.HandleFunc("PUT /api/devices/{mac}", s.updateDevice)
	mux.HandleFunc("DELETE /api/devices/{mac}", s.deleteDevice)

	// Device Groups
	mux.HandleFunc("GET /api/device-groups", s.listDeviceGroups)
	mux.HandleFunc("POST /api/device-groups", s.createDeviceGroup)
	mux.HandleFunc("GET /api/device-groups/{name}", s.getDeviceGroup)
	mux.HandleFunc("PUT /api/device-groups/{name}", s.updateDeviceGroup)
	mux.HandleFunc("DELETE /api/device-groups/{name}", s.deleteDeviceGroup)

	// Logs
	mux.HandleFunc("GET /api/logs", s.listLogs)
	mux.HandleFunc("GET /api/seen-devices", s.seenDevices)

	srv := &http.Server{Addr: addr, Handler: cors(mux)}
	s.log.Info("API server started", slog.String("listen", addr))

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// cors adds permissive CORS headers. Only needed when the frontend is
// served from a different origin (e.g. Vite dev server on :5173).
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
