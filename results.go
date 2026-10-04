package main

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// Player keeps track of the questions a player answered right and wrong.
type Player struct {
	Name         string
	GoodAnswered []Question
	BadAnswered  []Question
}

func (p *Player) Successes() int {
	return len(p.GoodAnswered)
}

func (p *Player) Failures() int {
	return len(p.BadAnswered)
}

// Score is the percentage of good answers, rounded to the nearest integer.
func (p *Player) Score() int {
	total := p.Successes() + p.Failures()
	if total == 0 {
		return 0
	}
	return (100*p.Successes() + total/2) / total
}

// winner returns the player with the most good answers; the first one wins ties.
func winner(players []*Player) *Player {
	var best *Player
	for _, player := range players {
		if best == nil || player.Successes() > best.Successes() {
			best = player
		}
	}
	return best
}

// writeResults writes one aligned line per player with score, successes and failures.
func writeResults(w io.Writer, players []*Player) {
	nameWidth := 0
	for _, player := range players {
		nameWidth = max(nameWidth, utf8.RuneCountInString(player.Name))
	}

	fmt.Fprintln(w, "\nRésultats :")
	for _, player := range players {
		fmt.Fprintf(w, "  %-*s  score %3d %%  ✅ %s  ❌ %s\n",
			nameWidth, player.Name, player.Score(),
			plural(player.Successes(), "réussite"), plural(player.Failures(), "échec"))
	}
}

// plural formats n followed by word, with an "s" when n > 1.
func plural(n int, word string) string {
	if n > 1 {
		word += "s"
	}
	return fmt.Sprintf("%d %s", n, word)
}
