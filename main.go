package main

import (
	"flag"
	"log"
)

/*
- solo
 - defininr un nb de tour
 - definir un temps et faire un max de bonnes réponses durant ce temps
 - un mode ou on choisit juste la première lettre
- multi en local au tour par tour
- multi distant
  - id de partie : create nbTurn
  - join player to partie
  - evict player from partie
  - start partie
  - stats
     - durée total : somme des tours des joueurs
     - moyenne des réponses par joueur
     - celui qui a réfléchi le plus longtemps
     - le plus rapide
     - celui qui a fait le moins d'erreur
     - remettre les verbes qui ont été en erreur
  - plusieurs pmode multi
     - le plus rapide à une même question qui tombe en même temps
     - tour par tour
     - on pose une question aux autres joueurs

 - gérer les traductions multiples
 - pas d'erreur si manque accent (à paramétrer)
 - gérer les lignes vides
 - ajouter un moyen de jouer la prononciation
 - gérer le fait qu'un mot peut avoir plusieurs solution
*/
func main() {
	var nbQuestions int
	flag.IntVar(&nbQuestions, "n", 3, "nombre de questions par joueur")
	flag.Parse()

	names := flag.Args()
	if len(names) == 0 {
		names = []string{"antoine"}
	}

	const deckPath = "anglais.csv"
	deck, err := LoadDeck(deckPath)
	if err != nil {
		log.Fatalf("Can't parse file %s: %v", deckPath, err)
	}

	game := NewGame(deck, nbQuestions, names)
	game.Run()
	game.ShowWinner()
	game.ShowResults()
}
