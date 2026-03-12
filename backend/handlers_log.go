package backend

import (
	"net/http"
	"strconv"

	"github.com/jmercereau/dns/store"
)

func (s *Server) listLogs(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	logs, err := s.store.ListLogs(limit)
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if logs == nil {
		logs = []store.LogEntry{}
	}
	writeJSON(w, http.StatusOK, logs)
}

func (s *Server) seenDevices(w http.ResponseWriter, r *http.Request) {
	seen, err := s.store.SeenDevices()
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if seen == nil {
		seen = []store.SeenDevice{}
	}
	writeJSON(w, http.StatusOK, seen)
}
