package server

import (
	"testing"
	"time"
)

func TestLiveListenerRegistryRegisterAndSnapshot(t *testing.T) {
	r := NewLiveListenerRegistry(nil)
	unreg := r.Register(LiveListener{
		ID:        "http:/live:1",
		Mount:     "/live",
		IP:        "203.0.113.10",
		UserAgent: "VLC/3.0",
		Transport: "http",
	}, false)
	defer unreg()

	r.AddBytes("http:/live:1", 4096)
	listeners := r.Snapshot("", "", false)
	if len(listeners) != 1 {
		t.Fatalf("expected 1 listener, got %d", len(listeners))
	}
	if listeners[0].IP != "203.0.113.10" {
		t.Fatalf("unexpected ip %q", listeners[0].IP)
	}
	if listeners[0].UserAgent != "VLC/3.0" {
		t.Fatalf("unexpected ua %q", listeners[0].UserAgent)
	}
	if listeners[0].BytesSent != 4096 {
		t.Fatalf("expected 4096 bytes, got %d", listeners[0].BytesSent)
	}
}

func TestLiveListenerRegistryHLSExpires(t *testing.T) {
	r := NewLiveListenerRegistry(nil)
	r.TouchHLS("/live", "198.51.100.20", "hls.js")
	if len(r.Snapshot("", "", false)) != 1 {
		t.Fatal("expected active hls listener")
	}

	r.mu.Lock()
	for _, e := range r.sessions {
		if e.ttl {
			e.LastSeen = time.Now().Add(-hlsListenerTTL - time.Second)
		}
	}
	r.mu.Unlock()

	if len(r.Snapshot("", "", false)) != 0 {
		t.Fatal("expected expired hls listener to be pruned")
	}
}

func TestLiveListenerRegistryMountFilter(t *testing.T) {
	r := NewLiveListenerRegistry(nil)
	r.Register(LiveListener{ID: "a", Mount: "/live", IP: "1.1.1.1", Transport: "http"}, false)
	r.Register(LiveListener{ID: "b", Mount: "/other", IP: "2.2.2.2", Transport: "http"}, false)

	filtered := r.Snapshot("/live", "", false)
	if len(filtered) != 1 || filtered[0].Mount != "/live" {
		t.Fatalf("mount filter failed: %+v", filtered)
	}
}
