package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	argsWithoutProg := os.Args[1:]
	var names []string
	if len(argsWithoutProg) == 0 {
		names = append(names, "antoine")
	} else {
		names = argsWithoutProg
	}

	var nbQuestions int
	flag.IntVar(&nbQuestions, "n", 3, "nombre de questions par joueur")

	filepath := "anglais.csv"

	game := NewGame(nbQuestions, names...)
	if err := game.ApplyThemesAndSubjects(filepath); err != nil {
		log.Fatalf("Can't parse file %s: %v", filepath, err)
	}
	game.Run()
	game.ShowWinner()
}
