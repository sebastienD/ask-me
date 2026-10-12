package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func postJSON(t *testing.T, handler http.Handler, url, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var data map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil {
		t.Fatalf("POST %s: invalid JSON %q", url, rec.Body.String())
	}
	return rec.Code, data
}

func TestServeIndex(t *testing.T) {
	handler := NewServer(NewRoom(echoDeck, ModeTurns, 1))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "<title>ask-me</title>") {
		t.Errorf("GET / = %d, want the game page", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /nope = %d, want 404", rec.Code)
	}
}

func TestJoinAPI(t *testing.T) {
	room := NewRoom(echoDeck, ModeTurns, 1)
	handler := NewServer(room)

	code, data := postJSON(t, handler, "/api/join", `{"name":"Léa"}`)
	if code != http.StatusOK || data["id"] == "" {
		t.Fatalf("join = %d %v, want an id", code, data)
	}

	code, data = postJSON(t, handler, "/api/join", `{"name":"léa"}`)
	if code != http.StatusBadRequest || data["error"] != ErrNameTaken.Error() {
		t.Errorf("duplicate join = %d %v, want 400 %q", code, data, ErrNameTaken)
	}

	code, _ = postJSON(t, handler, "/api/join", `not json`)
	if code != http.StatusBadRequest {
		t.Errorf("invalid body = %d, want 400", code)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/join", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /api/join = %d, want 405", rec.Code)
	}
}

func TestAnswerAPI(t *testing.T) {
	room := NewRoom(echoDeck, ModeTurns, 1)
	room.schedule = (&fakeClock{}).schedule
	handler := NewServer(room)
	id, err := room.Join("Léa")
	if err != nil {
		t.Fatal(err)
	}
	if err := room.Start(); err != nil {
		t.Fatal(err)
	}

	code, data := postJSON(t, handler, "/api/answer", `{"id":"unknown","answer":"x"}`)
	if code != http.StatusNotFound {
		t.Errorf("unknown player = %d %v, want 404", code, data)
	}

	code, data = postJSON(t, handler, "/api/answer", `{"id":"`+id+`","answer":"x"}`)
	if code != http.StatusOK || data["correct"] != true {
		t.Errorf("answer = %d %v, want correct", code, data)
	}
}

func getState(t *testing.T, url string) Snapshot {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d", url, resp.StatusCode)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	var s Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestStateAPI(t *testing.T) {
	room := NewRoom(echoDeck, ModeTurns, 1)
	server := httptest.NewServer(NewServer(room))
	defer server.Close()

	// Without a known version, the state comes right away.
	first := getState(t, server.URL+"/api/state")
	if first.Phase != PhaseLobby || len(first.Players) != 0 {
		t.Fatalf("state = %+v, want an empty lobby", first)
	}

	// With the current version, the request waits for the next change.
	changed := make(chan Snapshot)
	go func() {
		changed <- getState(t, fmt.Sprintf("%s/api/state?after=%d", server.URL, first.Version))
	}()
	select {
	case s := <-changed:
		t.Fatalf("answered before any change: %+v", s)
	case <-time.After(50 * time.Millisecond):
	}
	if _, err := room.Join("Léa"); err != nil {
		t.Fatal(err)
	}
	select {
	case s := <-changed:
		if s.Version <= first.Version || len(s.Players) != 1 {
			t.Errorf("state after join = %+v, want Léa and a new version", s)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no answer after the change")
	}
}

func TestStateAPITimeout(t *testing.T) {
	defer func(timeout time.Duration) { longPollTimeout = timeout }(longPollTimeout)
	longPollTimeout = 50 * time.Millisecond

	room := NewRoom(echoDeck, ModeTurns, 1)
	server := httptest.NewServer(NewServer(room))
	defer server.Close()

	version := room.Snapshot().Version
	s := getState(t, fmt.Sprintf("%s/api/state?after=%d", server.URL, version))
	if s.Version != version {
		t.Errorf("version = %d, want the unchanged %d", s.Version, version)
	}

	resp, err := http.Post(server.URL+"/api/state", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/state = %d, want 405", resp.StatusCode)
	}
}
