package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func randExcept(themes []string, indexExcept int) (index int, theme string) {
	chooseIndex := func() int {
		return rand.Intn(len(themes))
	}
	index = chooseIndex()
	for index == indexExcept {
		index = chooseIndex()
	}
	theme = themes[index]
	return
}

type Subject []string

type Game struct {
	themes   []string
	subjects []Subject
	players  []*Player
	nbTurn   int
}

type Player struct {
	name         string
	goodAnswered []Question
	badAnswered  []Question
}

type Question struct {
	themeGiven string
	themeAsked string
	line       int
}

func NewGame(nbTurn int, names ...string) *Game {
	players := make([]*Player, len(names))
	for i, name := range names {
		players[i] = &Player{name: name}
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
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("Could not close file", err)
		}
	}()

	head := true
	var nbThemes, numLine int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		numLine++
		line := scanner.Text()
		switch {
		case isComment(line):
			continue
		case head:
			head = false
			headers := strings.Split(line, ";")
			for _, h := range headers {
				t := strings.Trim(h, " ")
				g.themes = append(g.themes, t)
			}
			nbThemes = len(g.themes)
		default:
			subjects := strings.Split(line, ";")
			if len(subjects) < nbThemes {
				log.Printf("Skip the line %d (%s), too short.\n", numLine, line)
				continue
			}
			for i, elem := range subjects {
				subjects[i] = strings.Trim(elem, " ")
			}
			g.subjects = append(g.subjects, subjects)
		}
	}
	return nil
}

func (g *Game) Run() {
	for i := 0; i < g.nbTurn; i++ {
		for _, player := range g.players {
			g.playTurn(player)
		}
	}
}

func (g *Game) playTurn(player *Player) {
	themeIndexGiven, themeGiven := randExcept(g.themes, -1)
	themeIndexAsked, themeAsked := randExcept(g.themes, themeIndexGiven)
	subjectIndexAsked := rand.Intn(len(g.subjects))
	subjectAsked := g.subjects[subjectIndexAsked]
	question := Question{
		themeGiven: themeGiven,
		themeAsked: themeAsked,
		line:       subjectIndexAsked,
	}

	blue := color.New(color.FgBlue).SprintFunc()
	fmt.Printf("%s, si %s vaut %s, alors que vaut %s ?\n 👉 ", player.name, themeGiven, blue(subjectAsked[themeIndexGiven]), themeAsked)

	goodAnswer := subjectAsked[themeIndexAsked]
	answer := bufio.NewScanner(os.Stdin)
	if answer.Scan() && answer.Text() == goodAnswer {
		fmt.Print("👍  \n\n")
		player.goodAnswered = append(player.goodAnswered, question)
	} else {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("🥲 la bonne réponse était %s\n\n", green(goodAnswer))
		player.badAnswered = append(player.badAnswered, question)
	}
}

func (g *Game) ShowWinner() {
	fmt.Print("Le gagnant est")

	done := make(chan struct{})
	ticker := time.NewTicker(time.Second)
	go func() {
		for i := 0; i < 3; i++ {
			select {
			case <-ticker.C:
				fmt.Print(".")
			}
		}
		ticker.Stop()
		done <- struct{}{}
	}()
	<-done

	time.Sleep(time.Second)
	winner := g.players[0]
	for _, player := range g.players {
		if len(player.goodAnswered) > len(winner.goodAnswered) {
			winner = player
		}
	}
	fmt.Printf("   %s \n", winner.name)
}

func isComment(text string) bool {
	return strings.HasPrefix(text, "#")
}
