package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

//go:embed web/index.html
var indexHTML []byte

// keepAliveInterval keeps idle event streams open on phones and proxies.
const keepAliveInterval = 15 * time.Second

// NewServer serves the game page and its API:
//   - POST /api/join   {"name"}         → {"id"}
//   - POST /api/answer {"id", "answer"} → {"correct"}
//   - GET  /api/events                  → Server-Sent Events of Snapshot
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
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		streamEvents(w, r, room)
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

func streamEvents(w http.ResponseWriter, r *http.Request, room *Room) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	updates, unsubscribe := room.Subscribe()
	defer unsubscribe()
	keepAlive := time.NewTicker(keepAliveInterval)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepAlive.C:
			fmt.Fprint(w, ": keep-alive\n\n")
		case snapshot := <-updates:
			data, err := json.Marshal(snapshot)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
		}
		flusher.Flush()
	}
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
