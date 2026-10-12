package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

// cloudflaredCommand is the Cloudflare tunnel client; tests replace it.
var cloudflaredCommand = "cloudflared"

const (
	tunnelStartTimeout = 30 * time.Second
	tunnelReadyTimeout = 20 * time.Second
)

// tunnelPollInterval is the delay between two checks of the public address.
var tunnelPollInterval = time.Second

// tunnelURLPattern matches the public address printed by a quick tunnel.
var tunnelURLPattern = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

// ErrCloudflaredMissing is returned when cloudflared is not installed.
var ErrCloudflaredMissing = errors.New(`cloudflared est introuvable. Pour l'installer :
  macOS   : brew install cloudflared
  Windows : winget install --id Cloudflare.cloudflared
  Linux   : https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/`)

// Tunnel is a Cloudflare quick tunnel: a public https address forwarding to
// a local port, without account nor router configuration.
type Tunnel struct {
	URL string
	cmd *exec.Cmd
}

// checkCloudflared returns the path of cloudflared, or ErrCloudflaredMissing.
func checkCloudflared() (string, error) {
	path, err := exec.LookPath(cloudflaredCommand)
	if err != nil {
		return "", ErrCloudflaredMissing
	}
	return path, nil
}

// startTunnel runs cloudflared and waits until it prints its public address.
func startTunnel(port int) (*Tunnel, error) {
	path, err := checkCloudflared()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(path, "tunnel", "--no-autoupdate", "--url", fmt.Sprintf("http://localhost:%d", port))
	logs, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "pipe cloudflared logs")
	}
	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "start cloudflared")
	}

	url, err := findTunnelURL(logs, tunnelStartTimeout)
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, err
	}
	return &Tunnel{URL: url, cmd: cmd}, nil
}

// findTunnelURL reads cloudflared logs until they contain the public address.
// The logs keep being drained afterwards so that cloudflared never blocks.
func findTunnelURL(logs io.Reader, timeout time.Duration) (string, error) {
	found := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(logs)
		for scanner.Scan() {
			if url := tunnelURLPattern.FindString(scanner.Text()); url != "" {
				found <- url
				io.Copy(io.Discard, logs)
				return
			}
		}
		close(found)
	}()

	select {
	case url, ok := <-found:
		if !ok {
			return "", errors.New("cloudflared stopped without giving a public address")
		}
		return url, nil
	case <-time.After(timeout):
		return "", errors.Errorf("cloudflared gave no public address after %s", timeout)
	}
}

// publicResolver asks Cloudflare DNS directly. Asking the system resolver
// before the tunnel's DNS record exists would make the computer cache the
// failure, and the host could not open the link for a while.
var publicResolver = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		var dialer net.Dialer
		return dialer.DialContext(ctx, "udp", "1.1.1.1:53")
	},
}

// WaitReady polls the public address until the game page answers: a new
// quick tunnel takes a few seconds to be reachable.
func (t *Tunnel) WaitReady(timeout time.Duration) bool {
	dialer := &net.Dialer{Resolver: publicResolver}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DialContext: dialer.DialContext},
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if resp, err := client.Get(t.URL); err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(tunnelPollInterval)
	}
	return false
}

func (t *Tunnel) Stop() {
	t.cmd.Process.Kill()
	t.cmd.Wait()
}
