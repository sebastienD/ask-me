package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// cloudflaredLogs looks like what cloudflared prints when a quick tunnel starts.
const cloudflaredLogs = `2026-10-12T10:00:00Z INF Requesting new quick Tunnel on trycloudflare.com...
2026-10-12T10:00:01Z INF +--------------------------------------------------------------------------------------------+
2026-10-12T10:00:01Z INF |  Your quick Tunnel has been created! Visit it at (it may take some time to be reachable):  |
2026-10-12T10:00:01Z INF |  https://vert-lapin-42.trycloudflare.com                                                   |
2026-10-12T10:00:01Z INF +--------------------------------------------------------------------------------------------+
`

func TestFindTunnelURL(t *testing.T) {
	url, err := findTunnelURL(strings.NewReader(cloudflaredLogs), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://vert-lapin-42.trycloudflare.com" {
		t.Errorf("url = %q", url)
	}
}

// cloudflared must never block on its logs once the address is found.
func TestFindTunnelURLDrainsLogs(t *testing.T) {
	logs, writer := io.Pipe()
	go io.WriteString(writer, cloudflaredLogs)
	if _, err := findTunnelURL(logs, time.Second); err != nil {
		t.Fatal(err)
	}

	written := make(chan error)
	go func() {
		_, err := io.WriteString(writer, "INF Registered tunnel connection\n")
		written <- err
	}()
	select {
	case err := <-written:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("writing logs blocks after the address was found")
	}
}

func TestFindTunnelURLErrors(t *testing.T) {
	if _, err := findTunnelURL(strings.NewReader("ERR failed to request quick Tunnel\n"), time.Second); err == nil {
		t.Error("expected an error when cloudflared stops without address")
	}

	blocked, _ := io.Pipe()
	if _, err := findTunnelURL(blocked, 50*time.Millisecond); err == nil {
		t.Error("expected an error when cloudflared prints nothing")
	}
}

func TestStartTunnelWithoutCloudflared(t *testing.T) {
	defer func(command string) { cloudflaredCommand = command }(cloudflaredCommand)
	cloudflaredCommand = "cloudflared-does-not-exist"

	if _, err := startTunnel(4242); !errors.Is(err, ErrCloudflaredMissing) {
		t.Errorf("error = %v, want %v", err, ErrCloudflaredMissing)
	}
}

// A fake cloudflared prints the usual logs and records its arguments.
func TestStartTunnel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake cloudflared is a shell script")
	}
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	script := "#!/bin/sh\n" +
		"echo \"$@\" > " + argsFile + "\n" +
		"cat >&2 <<'LOGS'\n" + cloudflaredLogs + "LOGS\n" +
		"sleep 30\n"
	fake := filepath.Join(dir, "cloudflared")
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	defer func(command string) { cloudflaredCommand = command }(cloudflaredCommand)
	cloudflaredCommand = fake

	tunnel, err := startTunnel(4242)
	if err != nil {
		t.Fatal(err)
	}
	tunnel.Stop()

	if tunnel.URL != "https://vert-lapin-42.trycloudflare.com" {
		t.Errorf("url = %q", tunnel.URL)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if want := "tunnel --no-autoupdate --url http://localhost:4242"; strings.TrimSpace(string(args)) != want {
		t.Errorf("cloudflared args = %q, want %q", args, want)
	}
}

func TestWaitReady(t *testing.T) {
	defer func(interval time.Duration) { tunnelPollInterval = interval }(tunnelPollInterval)
	tunnelPollInterval = 10 * time.Millisecond

	// The tunnel answers 502 until Cloudflare has registered it.
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusBadGateway)
		}
	}))
	defer server.Close()
	if !(&Tunnel{URL: server.URL}).WaitReady(time.Second) {
		t.Error("WaitReady = false, want true once the page answers")
	}

	server.Close()
	if (&Tunnel{URL: server.URL}).WaitReady(100 * time.Millisecond) {
		t.Error("WaitReady = true for a closed server")
	}
}
