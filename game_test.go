package main

import (
	"bufio"
	"math/rand"
	"strings"
	"testing"
)

// newTestGame builds a game reading answers from input and writing to out.
func newTestGame(deck *Deck, nbTurns int, names []string, input string, out *strings.Builder) *Game {
	g := NewGame(deck, nbTurns, names)
	g.in = bufio.NewScanner(strings.NewReader(input))
	g.out = out
	g.rng = rand.New(rand.NewSource(1))
	return g
}

// With two identical columns, the expected answer is always the given word.
var echoDeck = &Deck{
	Themes:   []string{"a", "b"},
	Subjects: [][]string{{"x", "x"}},
}

func TestPlayTurn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantGood bool
	}{
		{"good answer", "x\n", true},
		{"good answer with spaces", "  x \n", true},
		{"bad answer", "y\n", false},
		{"wrong case", "X\n", false},
		{"no input", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &strings.Builder{}
			g := newTestGame(echoDeck, 1, []string{"alice"}, tt.input, out)
			player := g.players[0]

			g.playTurn(player)

			if got := player.Successes() == 1; got != tt.wantGood {
				t.Errorf("good answer = %v, want %v", got, tt.wantGood)
			}
			if len(player.GoodAnswered)+len(player.BadAnswered) != 1 {
				t.Errorf("expected exactly one recorded question")
			}
			if !tt.wantGood && !strings.Contains(out.String(), "la bonne réponse était") {
				t.Errorf("the correct answer should be shown, got %q", out.String())
			}
		})
	}
}

func TestRun(t *testing.T) {
	out := &strings.Builder{}
	// Answers alternate between alice and bob, turn after turn.
	g := newTestGame(echoDeck, 2, []string{"alice", "bob"}, "x\nwrong\nx\nx\n", out)

	g.Run()

	alice, bob := g.players[0], g.players[1]
	if alice.Successes() != 2 || alice.Failures() != 0 {
		t.Errorf("alice: %d good, %d bad; want 2 good, 0 bad", alice.Successes(), alice.Failures())
	}
	if bob.Successes() != 1 || bob.Failures() != 1 {
		t.Errorf("bob: %d good, %d bad; want 1 good, 1 bad", bob.Successes(), bob.Failures())
	}
	if n := strings.Count(out.String(), "alors que vaut"); n != 4 {
		t.Errorf("asked %d questions, want 4", n)
	}
}
