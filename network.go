package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/mdp/qrterminal/v3"
	"github.com/pkg/errors"
)

// runNetworkGame lets the host choose the mode, serves the game on the local
// network, and shows the results once the game is over.
func runNetworkGame(deck *Deck, nbQuestions, port int) error {
	in := bufio.NewScanner(os.Stdin)
	out := os.Stdout

	mode := chooseMode(in, out)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.Wrapf(err, "listen on port %d", port)
	}
	room := NewRoom(deck, mode, nbQuestions)
	go func() {
		if err := http.Serve(listener, NewServer(room)); err != nil {
			log.Fatalf("Server stopped: %v", err)
		}
	}()

	ip, found := lanIP()
	if !found {
		fmt.Fprintln(out, "⚠️  Impossible de trouver l'adresse de cet ordinateur sur le réseau local.")
	}
	printInvitation(out, mode, fmt.Sprintf("http://%s:%d", ip, port))

	finished := watchRoom(room, out)

	fmt.Fprint(out, "\nAppuie sur Entrée pour lancer la partie quand tout le monde est là.\n\n")
	for {
		if !in.Scan() {
			return errors.New("standard input closed")
		}
		err := room.Start()
		if err == nil {
			break
		}
		fmt.Fprintf(out, "⚠️  %v\n", err)
	}
	fmt.Fprintln(out, "🚀 C'est parti !")

	<-finished
	fmt.Fprintf(out, "\n🏆 Le gagnant est %s\n", winner(room.Players()).Name)
	writeResults(out, room.Players())

	// The server keeps running so that phones can show the results.
	fmt.Fprint(out, "\nAppuie sur Entrée pour quitter.\n")
	in.Scan()
	return nil
}

// chooseMode asks the host for the game mode. An empty answer (or no input)
// picks the first mode.
func chooseMode(in *bufio.Scanner, out io.Writer) Mode {
	modes := []Mode{ModeTurns, ModeRace}
	fmt.Fprintln(out, "Mode de jeu :")
	for i, mode := range modes {
		fmt.Fprintf(out, "  %d) %s\n", i+1, mode)
	}
	for {
		fmt.Fprint(out, "Ton choix [1] : ")
		if !in.Scan() {
			return modes[0]
		}
		switch strings.TrimSpace(in.Text()) {
		case "", "1":
			return modes[0]
		case "2":
			return modes[1]
		}
		fmt.Fprintln(out, "Tape 1 ou 2.")
	}
}

func printInvitation(out io.Writer, mode Mode, url string) {
	fmt.Fprintf(out, "\n🎮 Partie créée (mode : %s) !\n\n", mode)
	fmt.Fprintf(out, "📱 Pour jouer, ouvrez ce lien sur votre téléphone (toi aussi !) :\n\n   %s\n\n", url)
	fmt.Fprintf(out, "💬 Message à copier dans WhatsApp :\n\n   Viens jouer à ask-me avec moi 🎮 👉 %s\n\n", url)
	fmt.Fprintln(out, "📷 Ou scannez ce QR code :")
	qrterminal.GenerateHalfBlock(url, qrterminal.L, out)
	fmt.Fprintln(out, "⚠️  Tout le monde doit être connecté au même Wi-Fi (pas en 4G/5G).")
}

// watchRoom prints players joining the lobby, and closes the returned channel
// when the game is over.
func watchRoom(room *Room, out io.Writer) <-chan struct{} {
	finished := make(chan struct{})
	updates, unsubscribe := room.Subscribe()
	go func() {
		defer unsubscribe()
		defer close(finished)
		nbPlayers := 0
		for snapshot := range updates {
			for i := nbPlayers; i < len(snapshot.Players); i++ {
				fmt.Fprintf(out, "✅ %s a rejoint la partie (%s)\n",
					snapshot.Players[i].Name, plural(i+1, "joueur"))
			}
			nbPlayers = len(snapshot.Players)
			if snapshot.Phase == PhaseFinished {
				return
			}
		}
	}()
	return finished
}

// lanIP returns the address of this computer on the local network, or
// "localhost" and false if there is none.
func lanIP() (string, bool) {
	// Dialing UDP sends no packet: it only selects the outgoing interface.
	if conn, err := net.Dial("udp", "8.8.8.8:80"); err == nil {
		defer conn.Close()
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok && addr.IP.IsPrivate() {
			return addr.IP.String(), true
		}
	}

	// Without internet access, look at the network interfaces.
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost", false
	}
	var ips []net.IP
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			ips = append(ips, ipNet.IP)
		}
	}
	if ip := pickLANIP(ips); ip != nil {
		return ip.String(), true
	}
	return "localhost", false
}

// pickLANIP returns the first private IPv4 address, preferring 192.168.x.x
// which is what home Wi-Fi boxes use (10.x.x.x is often a VPN).
func pickLANIP(ips []net.IP) net.IP {
	var fallback net.IP
	for _, ip := range ips {
		ip4 := ip.To4()
		if ip4 == nil || !ip4.IsPrivate() {
			continue
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return ip4
		}
		if fallback == nil {
			fallback = ip4
		}
	}
	return fallback
}
