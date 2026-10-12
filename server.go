package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

//go:embed web/index.html
var indexHTML []byte

// longPollTimeout is how long /api/state waits for a change. It stays well
// below the 100 s after which Cloudflare cuts a request without response.
var longPollTimeout = 25 * time.Second

// NewServer serves the game page and its API:
//   - POST /api/join   {"name"}         → {"id"}
//   - POST /api/answer {"id", "answer"} → {"correct"}
//   - GET  /api/state?after=<version>   → Snapshot, once its version differs
//
// The state is sent by long polling rather than Server-Sent Events, because
// Cloudflare quick tunnels (-remote) do not support SSE.
func NewServer(room *Room) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/join", post(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if !decode(w, r, &req) {
			return
		}
		id, err := room.Join(req.Name)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"id": id})
	}))
	mux.HandleFunc("/api/answer", post(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string `json:"id"`
			Answer string `json:"answer"`
		}
		if !decode(w, r, &req) {
			return
		}
		correct, err := room.Answer(req.ID, req.Answer)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"correct": correct})
	}))
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		handleState(w, r, room)
	})
	return mux
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

// handleState answers as soon as the game differs from the version the phone
// already has, or after longPollTimeout with the same state.
func handleState(w http.ResponseWriter, r *http.Request, room *Room) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	after, err := strconv.Atoi(r.URL.Query().Get("after"))
	if err != nil {
		after = -1 // no known version: answer right away
	}

	ctx, cancel := context.WithTimeout(r.Context(), longPollTimeout)
	defer cancel()
	snapshot := room.WaitChange(ctx, after)

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, snapshot)
}

func post(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "requête invalide"})
		return false
	}
	return true
}

// writeError sends the room error message, which is meant for the players.
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, ErrUnknownPlayer) {
		status = http.StatusNotFound
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
