package main

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

func TestChooseMode(t *testing.T) {
	tests := []struct {
		input string
		want  Mode
	}{
		{"\n", ModeTurns},
		{"1\n", ModeTurns},
		{"2\n", ModeRace},
		{" 2 \n", ModeRace},
		{"3\nabc\n2\n", ModeRace}, // asks again until the answer is valid
		{"", ModeTurns},           // no input
	}
	for _, tt := range tests {
		out := &strings.Builder{}
		got := chooseMode(bufio.NewScanner(strings.NewReader(tt.input)), out)
		if got != tt.want {
			t.Errorf("chooseMode(%q) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestPickLANIP(t *testing.T) {
	ips := func(addrs ...string) []net.IP {
		var result []net.IP
		for _, addr := range addrs {
			result = append(result, net.ParseIP(addr))
		}
		return result
	}

	tests := []struct {
		name string
		ips  []net.IP
		want string
	}{
		{"home Wi-Fi", ips("127.0.0.1", "192.168.1.23"), "192.168.1.23"},
		{"prefers 192.168 over a VPN", ips("10.8.0.2", "192.168.1.23"), "192.168.1.23"},
		{"other private network", ips("127.0.0.1", "10.0.0.5"), "10.0.0.5"},
		{"ignores public and IPv6", ips("8.8.8.8", "fe80::1", "fd00::1"), "<nil>"},
		{"nothing", nil, "<nil>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickLANIP(tt.ips).String(); got != tt.want {
				t.Errorf("pickLANIP = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestPrintInvitation(t *testing.T) {
	out := &strings.Builder{}
	printInvitation(out, ModeRace, "http://192.168.1.23:4242")

	for _, want := range []string{
		"Le plus rapide",
		"WhatsApp",
		"Viens jouer à ask-me avec moi 🎮 👉 http://192.168.1.23:4242",
		"même Wi-Fi",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("invitation does not contain %q:\n%s", want, out.String())
		}
	}
}

func TestWatchRoom(t *testing.T) {
	room, clock, ids := newTestRoom(t, echoDeck, ModeRace, 1)
	out := &strings.Builder{}
	finished := watchRoom(room, out)

	for _, name := range []string{"Léa", "Tom"} {
		id, err := room.Join(name)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := room.Start(); err != nil {
		t.Fatal(err)
	}
	mustAnswer(t, room, ids[0], "x", true)
	clock.elapse()

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("watchRoom did not notice the end of the game")
	}
	for _, want := range []string{"✅ Léa a rejoint la partie (1 joueur)", "✅ Tom a rejoint la partie (2 joueurs)"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output does not contain %q:\n%s", want, out.String())
		}
	}
}
