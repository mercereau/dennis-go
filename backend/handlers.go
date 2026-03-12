package backend

import (
	"encoding/json"
	"net/http"

	"github.com/jmercereau/dns/store"
)

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func errJSON(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// --- Settings ---

type settingsBody struct {
	Listen         string `json:"listen"`
	DefaultProfile string `json:"default_profile"`
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, settingsBody{
		Listen:         s.store.Listen(),
		DefaultProfile: s.store.DefaultProfile(),
	})
}

func (s *Server) putSettings(w http.ResponseWriter, r *http.Request) {
	var body settingsBody
	if err := readJSON(r, &body); err != nil {
		errJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Listen != "" {
		if err := s.store.SetSetting("listen", body.Listen); err != nil {
			errJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if body.DefaultProfile != "" {
		if err := s.store.SetSetting("default_profile", body.DefaultProfile); err != nil {
			errJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	s.getSettings(w, r)
}

// --- Upstreams ---

func (s *Server) getUpstreams(w http.ResponseWriter, r *http.Request) {
	ups := s.store.Upstreams()
	if ups == nil {
		ups = []string{}
	}
	writeJSON(w, http.StatusOK, ups)
}

func (s *Server) putUpstreams(w http.ResponseWriter, r *http.Request) {
	var addrs []string
	if err := readJSON(r, &addrs); err != nil {
		errJSON(w, http.StatusBadRequest, "expected a JSON array of address strings")
		return
	}
	if err := s.store.SetUpstreams(addrs); err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.getUpstreams(w, r)
}

// --- Profiles ---

func (s *Server) listProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := s.store.ListProfiles()
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if profiles == nil {
		profiles = []store.ProfileRow{}
	}
	writeJSON(w, http.StatusOK, profiles)
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	p, err := s.store.GetProfile(name)
	if err != nil {
		errJSON(w, http.StatusNotFound, "profile not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type profileBody struct {
	Name      string   `json:"name"`
	Block     []string `json:"block"`
	AllowOnly []string `json:"allow_only"`
}

func (s *Server) createProfile(w http.ResponseWriter, r *http.Request) {
	var body profileBody
	if err := readJSON(r, &body); err != nil || body.Name == "" {
		errJSON(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := s.store.CreateProfile(body.Name, body.Block, body.AllowOnly); err != nil {
		errJSON(w, http.StatusConflict, err.Error())
		return
	}
	p, _ := s.store.GetProfile(body.Name)
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var body profileBody
	if err := readJSON(r, &body); err != nil {
		errJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := s.store.UpdateProfile(name, body.Block, body.AllowOnly); err != nil {
		errJSON(w, http.StatusNotFound, err.Error())
		return
	}
	p, _ := s.store.GetProfile(name)
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) deleteProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.store.DeleteProfile(name); err != nil {
		errJSON(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Devices ---

func (s *Server) listDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.store.ListDevices()
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if devices == nil {
		devices = []store.DeviceRow{}
	}
	writeJSON(w, http.StatusOK, devices)
}

func (s *Server) getDevice(w http.ResponseWriter, r *http.Request) {
	mac := r.PathValue("mac")
	d, err := s.store.GetDevice(mac)
	if err != nil {
		errJSON(w, http.StatusNotFound, "device not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

type deviceBody struct {
	MAC     string `json:"mac"`
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

func (s *Server) createDevice(w http.ResponseWriter, r *http.Request) {
	var body deviceBody
	if err := readJSON(r, &body); err != nil || body.MAC == "" {
		errJSON(w, http.StatusBadRequest, "mac is required")
		return
	}
	if err := s.store.CreateDevice(body.MAC, body.Name, body.Profile); err != nil {
		errJSON(w, http.StatusConflict, err.Error())
		return
	}
	d, _ := s.store.GetDevice(body.MAC)
	writeJSON(w, http.StatusCreated, d)
}

func (s *Server) updateDevice(w http.ResponseWriter, r *http.Request) {
	mac := r.PathValue("mac")
	var body deviceBody
	if err := readJSON(r, &body); err != nil {
		errJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := s.store.UpdateDevice(mac, body.Name, body.Profile); err != nil {
		errJSON(w, http.StatusNotFound, err.Error())
		return
	}
	d, _ := s.store.GetDevice(mac)
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) deleteDevice(w http.ResponseWriter, r *http.Request) {
	mac := r.PathValue("mac")
	if err := s.store.DeleteDevice(mac); err != nil {
		errJSON(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Device Groups ---

func (s *Server) listDeviceGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := s.store.ListDeviceGroups()
	if err != nil {
		errJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if groups == nil {
		groups = []store.DeviceGroupRow{}
	}
	writeJSON(w, http.StatusOK, groups)
}

func (s *Server) getDeviceGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	g, err := s.store.GetDeviceGroup(name)
	if err != nil {
		errJSON(w, http.StatusNotFound, "device group not found")
		return
	}
	writeJSON(w, http.StatusOK, g)
}

type deviceGroupBody struct {
	Name      string              `json:"name"`
	Profile   string              `json:"profile"`
	Devices   []string            `json:"devices"`
	Schedules []store.ScheduleRow `json:"schedules"`
}

func (s *Server) createDeviceGroup(w http.ResponseWriter, r *http.Request) {
	var body deviceGroupBody
	if err := readJSON(r, &body); err != nil || body.Name == "" || body.Profile == "" {
		errJSON(w, http.StatusBadRequest, "name and profile are required")
		return
	}
	if body.Devices == nil {
		body.Devices = []string{}
	}
	if body.Schedules == nil {
		body.Schedules = []store.ScheduleRow{}
	}
	if err := s.store.CreateDeviceGroup(body.Name, body.Profile, body.Devices, body.Schedules); err != nil {
		errJSON(w, http.StatusConflict, err.Error())
		return
	}
	g, _ := s.store.GetDeviceGroup(body.Name)
	writeJSON(w, http.StatusCreated, g)
}

func (s *Server) updateDeviceGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var body deviceGroupBody
	if err := readJSON(r, &body); err != nil || body.Profile == "" {
		errJSON(w, http.StatusBadRequest, "profile is required")
		return
	}
	if body.Devices == nil {
		body.Devices = []string{}
	}
	if body.Schedules == nil {
		body.Schedules = []store.ScheduleRow{}
	}
	if err := s.store.UpdateDeviceGroup(name, body.Profile, body.Devices, body.Schedules); err != nil {
		errJSON(w, http.StatusNotFound, err.Error())
		return
	}
	g, _ := s.store.GetDeviceGroup(name)
	writeJSON(w, http.StatusOK, g)
}

func (s *Server) deleteDeviceGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.store.DeleteDeviceGroup(name); err != nil {
		errJSON(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
