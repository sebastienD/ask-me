package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	blue  = color.New(color.FgBlue).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
)

// Game is a local, turn-based quiz: each turn, every player answers one question.
// in, out and rng are fields so that tests can replace them.
type Game struct {
	deck    *Deck
	players []*Player
	nbTurns int

	in  *bufio.Scanner
	out io.Writer
	rng *rand.Rand
}

// Player keeps track of the questions a player answered right and wrong.
type Player struct {
	Name         string
	GoodAnswered []Question
	BadAnswered  []Question
}

func (p *Player) Score() int {
	return len(p.GoodAnswered)
}

// Question asks for the Asked theme of a subject, knowing its Given theme.
type Question struct {
	Subject int
	Given   int
	Asked   int
}

func NewGame(deck *Deck, nbTurns int, names []string) *Game {
	players := make([]*Player, len(names))
	for i, name := range names {
		players[i] = &Player{Name: name}
	}

	return &Game{
		deck:    deck,
		players: players,
		nbTurns: nbTurns,
		in:      bufio.NewScanner(os.Stdin),
		out:     os.Stdout,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Game) Run() {
	for i := 0; i < g.nbTurns; i++ {
		for _, player := range g.players {
			g.playTurn(player)
		}
	}
}

func (g *Game) playTurn(player *Player) {
	q := g.newQuestion()
	subject := g.deck.Subjects[q.Subject]
	expected := subject[q.Asked]

	fmt.Fprintf(g.out, "%s, si %s vaut %s, alors que vaut %s ?\n 👉 ",
		player.Name, g.deck.Themes[q.Given], blue(subject[q.Given]), g.deck.Themes[q.Asked])

	if g.readAnswer() == expected {
		fmt.Fprint(g.out, "👍  \n\n")
		player.GoodAnswered = append(player.GoodAnswered, q)
	} else {
		fmt.Fprintf(g.out, "🥲 la bonne réponse était %s\n\n", green(expected))
		player.BadAnswered = append(player.BadAnswered, q)
	}
}

// newQuestion picks a random subject and two distinct random themes.
func (g *Game) newQuestion() Question {
	nbThemes := len(g.deck.Themes)
	given := g.rng.Intn(nbThemes)
	asked := g.rng.Intn(nbThemes - 1)
	if asked >= given {
		asked++
	}
	return Question{
		Subject: g.rng.Intn(len(g.deck.Subjects)),
		Given:   given,
		Asked:   asked,
	}
}

// readAnswer returns the next input line, or "" when the input is exhausted.
func (g *Game) readAnswer() string {
	if !g.in.Scan() {
		return ""
	}
	return strings.TrimSpace(g.in.Text())
}

// Winner returns the player with the most good answers; the first one wins ties.
func (g *Game) Winner() *Player {
	var winner *Player
	for _, player := range g.players {
		if winner == nil || player.Score() > winner.Score() {
			winner = player
		}
	}
	return winner
}

func (g *Game) ShowWinner() {
	fmt.Fprint(g.out, "Le gagnant est")
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		fmt.Fprint(g.out, ".")
	}
	time.Sleep(time.Second)
	fmt.Fprintf(g.out, "   %s \n", g.Winner().Name)
}
