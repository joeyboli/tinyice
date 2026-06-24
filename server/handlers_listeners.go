package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DatanoiseTV/tinyice/config"
)

func (s *Server) apiGetListeners(w http.ResponseWriter, r *http.Request) {
	user, ok := s.checkAuth(r)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if user.Role == config.RoleDJ {
		jsonError(w, "Forbidden", http.StatusForbidden)
		return
	}

	mount := strings.TrimSpace(r.URL.Query().Get("mount"))
	transport := strings.TrimSpace(r.URL.Query().Get("transport"))
	includeInternal := r.URL.Query().Get("include_internal") == "1"

	listeners := s.ListenerRegistry.Snapshot(mount, transport, includeInternal)
	jsonResponse(w, map[string]interface{}{
		"listeners": listeners,
		"total":     len(listeners),
	})
}

func (s *Server) handleAdminListeners(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.checkAuth(r); !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	mount := strings.TrimSpace(r.URL.Query().Get("mount"))
	transport := strings.TrimSpace(r.URL.Query().Get("transport"))
	includeInternal := r.URL.Query().Get("include_internal") == "1"

	listeners := s.ListenerRegistry.Snapshot(mount, transport, includeInternal)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"listeners": listeners,
		"total":     len(listeners),
	})
}

func (s *Server) touchHLSListener(mount string, r *http.Request) {
	if s.ListenerRegistry == nil {
		return
	}
	ip := parseClientIP(r.RemoteAddr)
	s.ListenerRegistry.TouchHLS(mount, ip, r.Header.Get("User-Agent"))
}

func (s *Server) registerHTTPListener(id, mount string, r *http.Request) func() {
	if s.ListenerRegistry == nil {
		return func() {}
	}
	ip := parseClientIP(r.RemoteAddr)
	return s.ListenerRegistry.Register(LiveListener{
		ID:        id,
		Mount:     mount,
		IP:        ip,
		UserAgent: r.Header.Get("User-Agent"),
		Transport: "http",
	}, false)
}

func (s *Server) registerWebRTCListener(mount, transport string, r *http.Request) func() {
	if s.ListenerRegistry == nil {
		return func() {}
	}
	ip := parseClientIP(r.RemoteAddr)
	id := transport + ":" + mount + ":" + ip + ":" + time.Now().Format("150405.000000")
	return s.ListenerRegistry.Register(LiveListener{
		ID:        id,
		Mount:     mount,
		IP:        ip,
		UserAgent: r.Header.Get("User-Agent"),
		Transport: transport,
	}, false)
}
