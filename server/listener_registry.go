package server

import (
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DatanoiseTV/tinyice/relay"
)

const hlsListenerTTL = 30 * time.Second

// LiveListener is one active playback session (HTTP stream, HLS poll,
// or WebRTC/WHEP).
type LiveListener struct {
	ID               string    `json:"id"`
	Mount            string    `json:"mount"`
	IP               string    `json:"ip"`
	UserAgent        string    `json:"user_agent"`
	Transport        string    `json:"transport"`
	ConnectedAt      time.Time `json:"connected_at"`
	LastSeen         time.Time `json:"last_seen"`
	ConnectedSeconds int       `json:"connected_seconds"`
	BytesSent        int64     `json:"bytes_sent"`
	CountryISO       string    `json:"country_iso,omitempty"`
	Country          string    `json:"country,omitempty"`
	City             string    `json:"city,omitempty"`
}

type liveListenerEntry struct {
	LiveListener
	internal bool
	ttl      bool // short-lived sessions (HLS polling)
}

// LiveListenerRegistry tracks connected listeners for the admin UI and API.
type LiveListenerRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*liveListenerEntry
	geo      *relay.GeoLookup
}

func NewLiveListenerRegistry(geo *relay.GeoLookup) *LiveListenerRegistry {
	return &LiveListenerRegistry{
		sessions: make(map[string]*liveListenerEntry),
		geo:      geo,
	}
}

func parseClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func (r *LiveListenerRegistry) enrichGeo(ip string) (iso, country, city string) {
	if r.geo == nil || ip == "" {
		return "", "", ""
	}
	info := r.geo.Lookup(ip)
	iso = info.ISO
	city = info.City
	if iso != "" {
		if meta, ok := relay.CountryCentroids()[iso]; ok {
			country = meta.Name
		}
	}
	return iso, country, city
}

// Register adds or replaces a long-lived listener session. Returns unregister.
func (r *LiveListenerRegistry) Register(l LiveListener, internal bool) func() {
	if r == nil || l.ID == "" {
		return func() {}
	}
	now := time.Now()
	if l.ConnectedAt.IsZero() {
		l.ConnectedAt = now
	}
	l.LastSeen = now
	if l.CountryISO == "" && l.IP != "" {
		l.CountryISO, l.Country, l.City = r.enrichGeo(l.IP)
	}

	r.mu.Lock()
	r.sessions[l.ID] = &liveListenerEntry{LiveListener: l, internal: internal}
	r.mu.Unlock()

	return func() { r.Unregister(l.ID) }
}

// TouchHLS registers or refreshes an HLS viewer keyed by mount+IP.
func (r *LiveListenerRegistry) TouchHLS(mount, ip, ua string) {
	if r == nil || mount == "" || ip == "" {
		return
	}
	id := "hls:" + mount + ":" + ip
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.sessions[id]; ok {
		e.LastSeen = now
		if ua != "" {
			e.UserAgent = ua
		}
		return
	}
	iso, country, city := r.enrichGeo(ip)
	r.sessions[id] = &liveListenerEntry{
		LiveListener: LiveListener{
			ID:          id,
			Mount:       mount,
			IP:          ip,
			UserAgent:   ua,
			Transport:   "hls",
			ConnectedAt: now,
			LastSeen:    now,
			CountryISO:  iso,
			Country:     country,
			City:        city,
		},
		ttl: true,
	}
}

// AddBytes accumulates outbound bytes for a session.
func (r *LiveListenerRegistry) AddBytes(id string, n int) {
	if r == nil || id == "" || n <= 0 {
		return
	}
	r.mu.RLock()
	e, ok := r.sessions[id]
	r.mu.RUnlock()
	if !ok {
		return
	}
	atomic.AddInt64(&e.BytesSent, int64(n))
}

func (r *LiveListenerRegistry) Unregister(id string) {
	if r == nil || id == "" {
		return
	}
	r.mu.Lock()
	delete(r.sessions, id)
	r.mu.Unlock()
}

func (r *LiveListenerRegistry) pruneExpiredLocked(now time.Time) {
	for id, e := range r.sessions {
		if !e.ttl {
			continue
		}
		if now.Sub(e.LastSeen) > hlsListenerTTL {
			delete(r.sessions, id)
		}
	}
}

// Snapshot returns active listeners, optionally filtered.
func (r *LiveListenerRegistry) Snapshot(mount, transport string, includeInternal bool) []LiveListener {
	if r == nil {
		return nil
	}
	now := time.Now()
	mount = strings.TrimSpace(mount)
	transport = strings.ToLower(strings.TrimSpace(transport))

	r.mu.Lock()
	r.pruneExpiredLocked(now)
	out := make([]LiveListener, 0, len(r.sessions))
	for _, e := range r.sessions {
		if e.internal && !includeInternal {
			continue
		}
		if mount != "" && e.Mount != mount {
			continue
		}
		if transport != "" && strings.ToLower(e.Transport) != transport {
			continue
		}
		l := e.LiveListener
		l.BytesSent = atomic.LoadInt64(&e.BytesSent)
		l.ConnectedSeconds = int(now.Sub(l.ConnectedAt).Seconds())
		if l.ConnectedSeconds < 0 {
			l.ConnectedSeconds = 0
		}
		out = append(out, l)
	}
	r.mu.Unlock()

	sort.Slice(out, func(i, j int) bool {
		if out[i].Mount != out[j].Mount {
			return out[i].Mount < out[j].Mount
		}
		if out[i].ConnectedAt.Equal(out[j].ConnectedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].ConnectedAt.Before(out[j].ConnectedAt)
	})
	return out
}

func (r *LiveListenerRegistry) Count() int {
	if r == nil {
		return 0
	}
	now := time.Now()
	r.mu.Lock()
	r.pruneExpiredLocked(now)
	n := len(r.sessions)
	r.mu.Unlock()
	return n
}
