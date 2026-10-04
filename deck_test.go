package main

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestParseDeck(t *testing.T) {
	input := `
# un commentaire avant l'en-tête
 base ; prétérit ; traduction

arise;arose;survenir
  # commentaire indenté
 be ; was/were ; être
go;went
`
	deck, err := ParseDeck(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantThemes := []string{"base", "prétérit", "traduction"}
	if !reflect.DeepEqual(deck.Themes, wantThemes) {
		t.Errorf("themes = %q, want %q", deck.Themes, wantThemes)
	}
	wantSubjects := [][]string{
		{"arise", "arose", "survenir"},
		{"be", "was/were", "être"},
	}
	if !reflect.DeepEqual(deck.Subjects, wantSubjects) {
		t.Errorf("subjects = %q, want %q", deck.Subjects, wantSubjects)
	}
}

func TestParseDeckErrors(t *testing.T) {
	tests := map[string]string{
		"empty":         "",
		"only comment":  "# rien\n",
		"one theme":     "base\narise\n",
		"no subject":    "base;prétérit\n\n# rien\n",
		"all too short": "base;prétérit;traduction\narise;arose\n",
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseDeck(strings.NewReader(input)); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

func TestLoadDeckMissingFile(t *testing.T) {
	if _, err := LoadDeck("does-not-exist.csv"); err == nil {
		t.Error("expected an error, got nil")
	}
}

// The shipped file must be complete: every verb has a value for every theme.
func TestLoadDeckAnglais(t *testing.T) {
	deck, err := LoadDeck("anglais.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, subject := range deck.Subjects {
		for j, value := range subject[:len(deck.Themes)] {
			if value == "" {
				t.Errorf("subject %d (%s) has no %s", i, subject[0], deck.Themes[j])
			}
		}
	}
}

func TestRandomQuestion(t *testing.T) {
	deck := &Deck{
		Themes:   []string{"a", "b", "c", "d"},
		Subjects: [][]string{{"1", "2", "3", "4"}, {"5", "6", "7", "8"}},
	}
	rng := rand.New(rand.NewSource(1))

	askedSeen := map[int]bool{}
	for i := 0; i < 1000; i++ {
		q := deck.RandomQuestion(rng)
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

func TestClueAndAnswer(t *testing.T) {
	deck := &Deck{
		Themes:   []string{"base", "prétérit", "traduction"},
		Subjects: [][]string{{"arise", "arose", "survenir"}, {"be", "was/were", "être"}},
	}
	q := Question{Subject: 1, Given: 2, Asked: 0}
	if got := deck.Clue(q); got != "être" {
		t.Errorf("clue = %q, want %q", got, "être")
	}
	if got := deck.Answer(q); got != "be" {
		t.Errorf("answer = %q, want %q", got, "be")
	}
}
