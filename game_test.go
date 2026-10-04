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

func TestNewQuestion(t *testing.T) {
	deck := &Deck{
		Themes:   []string{"a", "b", "c", "d"},
		Subjects: [][]string{{"1", "2", "3", "4"}, {"5", "6", "7", "8"}},
	}
	g := newTestGame(deck, 0, nil, "", &strings.Builder{})

	askedSeen := map[int]bool{}
	for i := 0; i < 1000; i++ {
		q := g.newQuestion()
		if q.Given == q.Asked {
			t.Fatalf("given and asked themes are the same: %+v", q)
		}
		if q.Given < 0 || q.Given >= 4 || q.Asked < 0 || q.Asked >= 4 {
			t.Fatalf("theme out of range: %+v", q)
		}
		if q.Subject < 0 || q.Subject >= 2 {
			t.Fatalf("subject out of range: %+v", q)
		}
		askedSeen[q.Asked] = true
	}
	if len(askedSeen) != 4 {
		t.Errorf("every theme should be asked at some point, got %v", askedSeen)
	}
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

func TestWinner(t *testing.T) {
	good := func(n int) []Question { return make([]Question, n) }

	tests := []struct {
		name    string
		players []*Player
		want    string
	}{
		{"single player", []*Player{{Name: "alice"}}, "alice"},
		{"best score", []*Player{
			{Name: "alice", GoodAnswered: good(1)},
			{Name: "bob", GoodAnswered: good(3)},
			{Name: "carol", GoodAnswered: good(2)},
		}, "bob"},
		{"tie goes to first", []*Player{
			{Name: "alice", GoodAnswered: good(2)},
			{Name: "bob", GoodAnswered: good(2)},
		}, "alice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{players: tt.players}
			if got := g.Winner().Name; got != tt.want {
				t.Errorf("winner = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestScore(t *testing.T) {
	answers := func(n int) []Question { return make([]Question, n) }

	tests := []struct {
		name      string
		good, bad int
		want      int
	}{
		{"no answer", 0, 0, 0},
		{"all good", 3, 0, 100},
		{"all bad", 0, 3, 0},
		{"rounded down", 1, 2, 33},
		{"rounded up", 2, 1, 67},
		{"half", 1, 1, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{GoodAnswered: answers(tt.good), BadAnswered: answers(tt.bad)}
			if got := p.Score(); got != tt.want {
				t.Errorf("score = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestShowResults(t *testing.T) {
	answers := func(n int) []Question { return make([]Question, n) }
	out := &strings.Builder{}
	g := &Game{out: out, players: []*Player{
		{Name: "alice", GoodAnswered: answers(2), BadAnswered: answers(1)},
		{Name: "bob", GoodAnswered: answers(1), BadAnswered: answers(0)},
	}}

	g.ShowResults()

	want := "\nRésultats :\n" +
		"  alice  score  67 %  ✅ 2 réussites  ❌ 1 échec\n" +
		"  bob    score 100 %  ✅ 1 réussite  ❌ 0 échec\n"
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}
