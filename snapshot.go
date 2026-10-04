package main

// Snapshot is the state of a room as seen by the players. It never contains
// the answer of the current question before the reveal.
type Snapshot struct {
	Phase    Phase         `json:"phase"`
	Mode     string        `json:"mode"`
	Players  []PlayerView  `json:"players"`
	Question *QuestionView `json:"question,omitempty"`
	Reveal   *Reveal       `json:"reveal,omitempty"`
	Winner   string        `json:"winner,omitempty"`
}

type PlayerView struct {
	Name      string `json:"name"`
	Successes int    `json:"successes"`
	Failures  int    `json:"failures"`
	Score     int    `json:"score"`
}

type QuestionView struct {
	Number     int      `json:"number"` // starts at 1
	Total      int      `json:"total"`
	GivenTheme string   `json:"givenTheme"`
	Clue       string   `json:"clue"`
	AskedTheme string   `json:"askedTheme"`
	Turn       string   `json:"turn,omitempty"` // ModeTurns: who must answer
	Answered   []string `json:"answered"`
	Deadline   int64    `json:"deadline"` // Unix time in milliseconds
}

// Reveal is the answer of the question that just ended.
type Reveal struct {
	Answer string `json:"answer"`
	Finder string `json:"finder,omitempty"` // empty when nobody found it
}

// snapshot builds the current state. The caller must hold the lock.
func (r *Room) snapshot() Snapshot {
	s := Snapshot{
		Phase:   r.phase,
		Mode:    r.mode.String(),
		Players: make([]PlayerView, len(r.players)),
		Reveal:  r.reveal,
	}
	for i, player := range r.players {
		s.Players[i] = PlayerView{
			Name:      player.Name,
			Successes: player.Successes(),
			Failures:  player.Failures(),
			Score:     player.Score(),
		}
	}

	if r.phase == PhaseQuestion || r.phase == PhaseReveal {
		q := &QuestionView{
			Number:     r.round + 1,
			Total:      r.total,
			GivenTheme: r.deck.Themes[r.question.Given],
			Clue:       r.deck.Clue(r.question),
			AskedTheme: r.deck.Themes[r.question.Asked],
			Answered:   []string{},
			Deadline:   r.deadline.UnixMilli(),
		}
		if r.turn != nil && r.mode == ModeTurns {
			q.Turn = r.turn.Name
		}
		for _, player := range r.players {
			if r.answered[player] {
				q.Answered = append(q.Answered, player.Name)
			}
		}
		s.Question = q
	}

	if r.phase == PhaseFinished && len(r.players) > 0 {
		s.Winner = winner(r.players).Name
	}
	return s
}
