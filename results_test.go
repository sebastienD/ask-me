package main

import (
	"strings"
	"testing"
)

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
			if got := winner(tt.players).Name; got != tt.want {
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

func TestWriteResults(t *testing.T) {
	answers := func(n int) []Question { return make([]Question, n) }
	out := &strings.Builder{}
	players := []*Player{
		{Name: "alice", GoodAnswered: answers(2), BadAnswered: answers(1)},
		{Name: "bob", GoodAnswered: answers(1), BadAnswered: answers(0)},
	}

	writeResults(out, players)

	want := "\nRésultats :\n" +
		"  alice  score  67 %  ✅ 2 réussites  ❌ 1 échec\n" +
		"  bob    score 100 %  ✅ 1 réussite  ❌ 0 échec\n"
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}
