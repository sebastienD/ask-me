package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const separator = ";"

// Deck holds the questions material: each theme is a column of the CSV file
// (e.g. "le prétérit") and each subject is a row (e.g. a verb).
type Deck struct {
	Themes   []string
	Subjects [][]string
}

// LoadDeck reads and validates the deck stored in the CSV file at path.
func LoadDeck(path string) (*Deck, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "open file path %s", path)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("Could not close file", err)
		}
	}()

	return ParseDeck(f)
}

// ParseDeck reads a deck from r. The first meaningful line is the header;
// blank lines and lines starting with '#' are ignored, and rows with fewer
// columns than the header are skipped.
func ParseDeck(r io.Reader) (*Deck, error) {
	deck := &Deck{}
	numLine := 0
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		numLine++
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "", isComment(line):
			continue
		case deck.Themes == nil:
			deck.Themes = splitLine(line)
		default:
			subject := splitLine(line)
			if len(subject) < len(deck.Themes) {
				log.Printf("Skip the line %d (%s), too short.\n", numLine, line)
				continue
			}
			deck.Subjects = append(deck.Subjects, subject)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "read deck")
	}

	if len(deck.Themes) < 2 {
		return nil, errors.Errorf("need at least 2 themes, got %d", len(deck.Themes))
	}
	if len(deck.Subjects) == 0 {
		return nil, errors.New("no subject found")
	}
	return deck, nil
}

func splitLine(line string) []string {
	fields := strings.Split(line, separator)
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}
	return fields
}

func isComment(text string) bool {
	return strings.HasPrefix(text, "#")
}
