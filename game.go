package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"math/rand"
	"os"
	"strings"
)

type Theme string

type Subject []string

type Game struct {
	themes   []Theme
	subjects []Subject
	players  []Player
	nbTurn   int
	turn     int
}

type Player struct {
	name        string
	goodAnswers []Question
	badAnswers  []Question
}

type Question struct {
	header int
	line   []string
}

func NewGame(nbTurn int, names ...string) *Game {
	players := make([]Player, len(names))
	for i, name := range names {
		players[i] = Player{name: name}
	}
	return &Game{
		nbTurn:  nbTurn,
		players: players,
	}
}

func (g *Game) ApplyThemesAndSubjects(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "open file path %s", path)
	}
	defer f.Close()

	head := true
	var nbThemes int
	var numLine int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		numLine++
		switch {
		case isComment(line):
			continue
		case head:
			head = false
			headers := strings.Split(line, ";")
			for _, h := range headers {
				t := Theme(strings.Trim(h, " "))
				g.themes = append(g.themes, t)
			}
			nbThemes = len(g.themes)
		default:
			subject := strings.Split(line, ";")
			if len(subject) < nbThemes {
				log.Println("Skip the line %d (%s), too short.", numLine, line)
				continue
			}
			for i, elem := range subject {
				subject[i] = strings.Trim(elem, " ")
			}
			g.subjects = append(g.subjects, subject)
		}
	}
	return nil
}

func (g *Game) Run() {
	for i := 0; i < g.nbTurn; i++ {
		give := rand.Intn(len(g.themes))
		ask := rand.Intn(len(g.subjects))
		fmt.Printf("Voici %s: %s\n 👉 ", g.themes[give], g.subjects[ask][give])
		answer := bufio.NewScanner(os.Stdin)
		good := g.subjects[ask][len(g.themes)-1]
		if answer.Scan() && answer.Text() == good {
			fmt.Println("👍")
		} else {
			fmt.Printf("🥲 the good answer was %s\n", good)
		}
	}
}

func isComment(text string) bool {
	return strings.HasPrefix(text, "#")
}
