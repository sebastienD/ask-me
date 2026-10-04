package main

import (
	"bufio"
	"encoding/json"
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

func TestEventsAPI(t *testing.T) {
	room := NewRoom(echoDeck, ModeTurns, 1)
	server := httptest.NewServer(NewServer(room))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}

	events := make(chan Snapshot)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			data, ok := strings.CutPrefix(scanner.Text(), "data: ")
			if !ok {
				continue
			}
			var s Snapshot
			if err := json.Unmarshal([]byte(data), &s); err == nil {
				events <- s
			}
		}
		close(events)
	}()
	next := func() Snapshot {
		t.Helper()
		select {
		case s := <-events:
			return s
		case <-time.After(2 * time.Second):
			t.Fatal("no event received")
			return Snapshot{}
		}
	}

	if s := next(); s.Phase != PhaseLobby || len(s.Players) != 0 {
		t.Errorf("first event = %+v, want an empty lobby", s)
	}
	if _, err := room.Join("Léa"); err != nil {
		t.Fatal(err)
	}
	if s := next(); len(s.Players) != 1 || s.Players[0].Name != "Léa" {
		t.Errorf("event after join = %+v, want Léa in the lobby", s)
	}
}
