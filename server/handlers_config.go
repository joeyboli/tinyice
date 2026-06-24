package server

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DatanoiseTV/tinyice/config"
)

func (s *Server) apiExportConfig(w http.ResponseWriter, r *http.Request) {
	user, ok := s.checkAuth(r)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if user.Role != config.RoleSuperAdmin {
		jsonError(w, "Forbidden", http.StatusForbidden)
		return
	}

	data, err := s.Config.Export()
	if err != nil {
		jsonError(w, "Export failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="tinyice-config.json"`)
	w.Write(data)
	s.Audit(r, "config_exported", "settings", "", "")
}

func (s *Server) apiImportConfig(w http.ResponseWriter, r *http.Request) {
	if !s.isCSRFSafe(r) {
		jsonError(w, "Forbidden", http.StatusForbidden)
		return
	}
	user, ok := s.checkAuth(r)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if user.Role != config.RoleSuperAdmin {
		jsonError(w, "Forbidden", http.StatusForbidden)
		return
	}

	var incoming config.Config
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		jsonError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := s.Config.MergeImport(&incoming); err != nil {
		jsonError(w, "Invalid config", http.StatusBadRequest)
		return
	}
	if err := s.Config.SaveConfig(); err != nil {
		jsonError(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	s.Relay.LowLatency = s.Config.LowLatencyMode
	jsonResponse(w, map[string]string{"status": "imported"})
	s.Audit(r, "config_imported", "settings", "", "")
}

func (s *Server) handleInsightsExport(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.checkAuth(r); !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if s.Relay.History == nil {
		http.Error(w, "History disabled", http.StatusServiceUnavailable)
		return
	}

	hours := 24
	if v := r.URL.Query().Get("hours"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 24*365*5 {
			hours = n
		}
	}
	mountFilter := strings.TrimSpace(r.URL.Query().Get("mount"))

	stats := s.Relay.History.GetAllHistoricalStats(time.Duration(hours) * time.Hour)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="tinyice-listeners-%dh.csv"`, hours))

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"mount", "timestamp", "listeners", "bytes_in", "bytes_out"})
	for mount, series := range stats {
		if mountFilter != "" && mount != mountFilter {
			continue
		}
		for _, pt := range series {
			_ = cw.Write([]string{
				mount,
				pt.Timestamp.UTC().Format(time.RFC3339),
				strconv.Itoa(pt.Listeners),
				strconv.FormatInt(pt.BytesIn, 10),
				strconv.FormatInt(pt.BytesOut, 10),
			})
		}
	}
	cw.Flush()
}

type podcastFeed struct {
	XMLName xml.Name       `xml:"rss"`
	Version string         `xml:"version,attr"`
	Channel podcastChannel `xml:"channel"`
}

type podcastChannel struct {
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Description string        `xml:"description"`
	Language    string        `xml:"language"`
	LastBuild   string        `xml:"lastBuildDate"`
	Items       []podcastItem `xml:"item"`
}

type podcastItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

func (s *Server) handlePodcastFeed(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasSuffix(path, "/feed.xml") {
		http.NotFound(w, r)
		return
	}
	mount := strings.TrimSuffix(path, "/feed.xml")
	if mount == "" {
		http.NotFound(w, r)
		return
	}
	if mount[0] != '/' {
		mount = "/" + mount
	}

	if _, ok := s.Relay.GetStream(mount); !ok {
		if fb, has := s.Config.FallbackMounts[mount]; has {
			mount = fb
		} else {
			http.NotFound(w, r)
			return
		}
	}

	stationName := s.Config.PageTitle
	if stream, ok := s.Relay.GetStream(mount); ok {
		snap := stream.Snapshot()
		if snap.Name != "" {
			stationName = snap.Name
		}
	}

	baseURL := strings.TrimRight(s.Config.BaseURL, "/")
	if baseURL == "" {
		baseURL = "http://" + r.Host
	}

	var items []podcastItem
	if s.Relay.History != nil {
		for _, h := range s.Relay.History.Get(mount) {
			items = append(items, podcastItem{
				Title:       h.Song,
				Description: h.Song,
				GUID:        fmt.Sprintf("%s-%d", mount, h.ID),
				PubDate:     h.Timestamp.UTC().Format(time.RFC1123Z),
			})
		}
	}

	feed := podcastFeed{
		Version: "2.0",
		Channel: podcastChannel{
			Title:       stationName,
			Link:        baseURL + mount,
			Description: fmt.Sprintf("Recent tracks on %s", mount),
			Language:    "en-us",
			LastBuild:   time.Now().UTC().Format(time.RFC1123Z),
			Items:       items,
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	_ = enc.Encode(feed)
}

func playerOptionsFromRequest(r *http.Request) map[string]interface{} {
	q := r.URL.Query()
	opts := map[string]interface{}{}
	if q.Get("stats") == "0" {
		opts["hideStats"] = true
	}
	if q.Get("autoplay") == "1" {
		opts["autoplay"] = true
	}
	if v := strings.TrimSpace(q.Get("accent")); v != "" {
		opts["accent"] = v
	}
	if q.Get("visualizer") == "0" {
		opts["hideVisualizer"] = true
	}
	if q.Get("webrtc") == "1" || strings.EqualFold(q.Get("mode"), "webrtc") {
		opts["webrtc"] = true
	}
	return opts
}
