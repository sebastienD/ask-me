package main

import (
	"log"
)

func main() {
	filepath := "anglais.csv"
	game := NewGame(2, "sebastien")
	if err := game.ApplyThemesAndSubjects(filepath); err != nil {
		log.Fatalf("Can't parse file %s: %v", filepath, err)
	}
	game.Run()
}
