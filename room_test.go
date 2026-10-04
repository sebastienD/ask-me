package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

// fakeClock records scheduled functions so that tests decide when time passes.
type fakeClock struct {
	pending []func()
}

func (c *fakeClock) schedule(_ time.Duration, f func()) {
	c.pending = append(c.pending, f)
}

// elapse runs every function scheduled so far, as if all delays had expired.
func (c *fakeClock) elapse() {
	pending := c.pending
	c.pending = nil
	for _, f := range pending {
		f()
	}
}

func newTestRoom(t *testing.T, deck *Deck, mode Mode, nbQuestions int, names ...string) (*Room, *fakeClock, []string) {
	t.Helper()
	clock := &fakeClock{}
	r := NewRoom(deck, mode, nbQuestions)
	r.rng = rand.New(rand.NewSource(1))
	r.schedule = clock.schedule

	ids := make([]string, len(names))
	for i, name := range names {
		id, err := r.Join(name)
		if err != nil {
			t.Fatalf("join %s: %v", name, err)
		}
		ids[i] = id
	}
	return r, clock, ids
}

func mustAnswer(t *testing.T, r *Room, id, answer string, wantCorrect bool) {
	t.Helper()
	correct, err := r.Answer(id, answer)
	if err != nil {
		t.Fatalf("answer %q: %v", answer, err)
	}
	if correct != wantCorrect {
		t.Fatalf("answer %q: correct = %v, want %v", answer, correct, wantCorrect)
	}
}

func assertPhase(t *testing.T, r *Room, want Phase) {
	t.Helper()
	if got := r.Snapshot().Phase; got != want {
		t.Fatalf("phase = %s, want %s", got, want)
	}
}

func assertScores(t *testing.T, r *Room, want map[string][2]int) {
	t.Helper()
	for _, p := range r.Snapshot().Players {
		if got := [2]int{p.Successes, p.Failures}; got != want[p.Name] {
			t.Errorf("%s: [successes failures] = %v, want %v", p.Name, got, want[p.Name])
		}
	}
}

func TestJoin(t *testing.T) {
	r, _, _ := newTestRoom(t, echoDeck, ModeTurns, 1, "Léa")

	tests := []struct {
		name string
		want error
	}{
		{"", ErrEmptyName},
		{"   ", ErrEmptyName},
		{"léa", ErrNameTaken},
		{" Léa ", ErrNameTaken},
		{strings.Repeat("a", maxNameLength+1), ErrNameTooLong},
	}
	for _, tt := range tests {
		if _, err := r.Join(tt.name); !errors.Is(err, tt.want) {
			t.Errorf("Join(%q) error = %v, want %v", tt.name, err, tt.want)
		}
	}

	id, err := r.Join(" Tom ")
	if err != nil || id == "" {
		t.Fatalf("Join(Tom) = %q, %v", id, err)
	}
	if got := r.Snapshot().Players[1].Name; got != "Tom" {
		t.Errorf("name = %q, want it trimmed", got)
	}
}

func TestJoinAfterStart(t *testing.T) {
	r, _, _ := newTestRoom(t, echoDeck, ModeTurns, 1, "Léa")
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Join("Tom"); !errors.Is(err, ErrAlreadyStarted) {
		t.Errorf("error = %v, want %v", err, ErrAlreadyStarted)
	}
	if err := r.Start(); !errors.Is(err, ErrAlreadyStarted) {
		t.Errorf("second start error = %v, want %v", err, ErrAlreadyStarted)
	}
}

func TestStartWithoutPlayer(t *testing.T) {
	r, _, _ := newTestRoom(t, echoDeck, ModeTurns, 1)
	if err := r.Start(); !errors.Is(err, ErrNoPlayer) {
		t.Errorf("error = %v, want %v", err, ErrNoPlayer)
	}
}

func TestAnswerErrors(t *testing.T) {
	r, _, ids := newTestRoom(t, echoDeck, ModeTurns, 1, "Léa", "Tom")
	if _, err := r.Answer(ids[0], "x"); !errors.Is(err, ErrNoQuestion) {
		t.Errorf("before start: error = %v, want %v", err, ErrNoQuestion)
	}
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Answer("unknown", "x"); !errors.Is(err, ErrUnknownPlayer) {
		t.Errorf("unknown id: error = %v, want %v", err, ErrUnknownPlayer)
	}
	if _, err := r.Answer(ids[1], "x"); !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("not your turn: error = %v, want %v", err, ErrNotYourTurn)
	}
}

func TestTurnsMode(t *testing.T) {
	r, clock, ids := newTestRoom(t, echoDeck, ModeTurns, 1, "Léa", "Tom")
	lea, tom := ids[0], ids[1]
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}

	q := r.Snapshot().Question
	if q.Turn != "Léa" || q.Number != 1 || q.Total != 2 {
		t.Fatalf("first question = %+v, want Léa's turn, 1/2", q)
	}
	mustAnswer(t, r, lea, "x", true)
	assertPhase(t, r, PhaseReveal)
	if reveal := r.Snapshot().Reveal; reveal.Finder != "Léa" || reveal.Answer != "x" {
		t.Errorf("reveal = %+v, want Léa found x", reveal)
	}

	clock.elapse()
	if q := r.Snapshot().Question; q.Turn != "Tom" || q.Number != 2 {
		t.Fatalf("second question = %+v, want Tom's turn, 2/2", q)
	}
	mustAnswer(t, r, tom, "y", false)
	if reveal := r.Snapshot().Reveal; reveal.Finder != "" {
		t.Errorf("reveal = %+v, want nobody found it", reveal)
	}

	clock.elapse()
	assertPhase(t, r, PhaseFinished)
	assertScores(t, r, map[string][2]int{"Léa": {1, 0}, "Tom": {0, 1}})
	if got := r.Snapshot().Winner; got != "Léa" {
		t.Errorf("winner = %s, want Léa", got)
	}
}

func TestRaceMode(t *testing.T) {
	r, clock, ids := newTestRoom(t, echoDeck, ModeRace, 2, "Léa", "Tom", "Zoé")
	lea, tom, zoe := ids[0], ids[1], ids[2]
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	if q := r.Snapshot().Question; q.Turn != "" || q.Total != 2 {
		t.Fatalf("question = %+v, want no turn and 2 questions", q)
	}

	// A wrong answer does not end the question, but the player can't retry.
	mustAnswer(t, r, tom, "y", false)
	assertPhase(t, r, PhaseQuestion)
	if got := r.Snapshot().Question.Answered; !reflect.DeepEqual(got, []string{"Tom"}) {
		t.Errorf("answered = %v, want [Tom]", got)
	}
	if _, err := r.Answer(tom, "x"); !errors.Is(err, ErrAlreadyAnswered) {
		t.Errorf("retry error = %v, want %v", err, ErrAlreadyAnswered)
	}

	// The first good answer ends the question.
	mustAnswer(t, r, zoe, "x", true)
	assertPhase(t, r, PhaseReveal)
	if got := r.Snapshot().Reveal.Finder; got != "Zoé" {
		t.Errorf("finder = %s, want Zoé", got)
	}

	// The question ends when everybody answered wrong.
	clock.elapse()
	mustAnswer(t, r, lea, "y", false)
	mustAnswer(t, r, tom, "y", false)
	assertPhase(t, r, PhaseQuestion)
	mustAnswer(t, r, zoe, "y", false)
	assertPhase(t, r, PhaseReveal)

	clock.elapse()
	assertPhase(t, r, PhaseFinished)
	assertScores(t, r, map[string][2]int{"Léa": {0, 1}, "Tom": {0, 2}, "Zoé": {1, 1}})
	if got := r.Snapshot().Winner; got != "Zoé" {
		t.Errorf("winner = %s, want Zoé", got)
	}
}

func TestTimeout(t *testing.T) {
	tests := []struct {
		mode Mode
		want [2]int
	}{
		{ModeTurns, [2]int{0, 1}}, // not answering in time is a failure
		{ModeRace, [2]int{0, 0}},  // nobody is penalized
	}
	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			r, clock, _ := newTestRoom(t, echoDeck, tt.mode, 1, "Léa")
			if err := r.Start(); err != nil {
				t.Fatal(err)
			}

			clock.elapse()
			assertPhase(t, r, PhaseReveal)
			if reveal := r.Snapshot().Reveal; reveal.Finder != "" || reveal.Answer != "x" {
				t.Errorf("reveal = %+v, want nobody found x", reveal)
			}
			assertScores(t, r, map[string][2]int{"Léa": tt.want})
		})
	}
}

// A timer from a question that already ended must not end the next one.
func TestStaleTimeout(t *testing.T) {
	r, clock, ids := newTestRoom(t, echoDeck, ModeTurns, 2, "Léa")
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	staleTimeout := clock.pending[0]
	mustAnswer(t, r, ids[0], "x", true)
	clock.elapse() // timeout of question 1 (ignored) then next question

	staleTimeout()
	assertPhase(t, r, PhaseQuestion)
	if got := r.Snapshot().Question.Number; got != 2 {
		t.Errorf("question number = %d, want 2", got)
	}
}

func TestSnapshotHidesAnswer(t *testing.T) {
	deck := &Deck{Themes: []string{"a", "b"}, Subjects: [][]string{{"indice", "secret"}}}
	r, _, _ := newTestRoom(t, deck, ModeRace, 1, "Léa")
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}

	snapshot := r.Snapshot()
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	answer := deck.Answer(r.question)
	if snapshot.Reveal != nil || strings.Contains(string(data), `"`+answer+`"`) {
		t.Errorf("snapshot leaks the answer %q: %s", answer, data)
	}
}

func TestSubscribe(t *testing.T) {
	r, _, _ := newTestRoom(t, echoDeck, ModeTurns, 1)
	updates, unsubscribe := r.Subscribe()

	if s := <-updates; s.Phase != PhaseLobby || len(s.Players) != 0 {
		t.Errorf("first snapshot = %+v, want an empty lobby", s)
	}
	if _, err := r.Join("Léa"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Join("Tom"); err != nil {
		t.Fatal(err)
	}
	// Only the latest snapshot is kept.
	if s := <-updates; len(s.Players) != 2 {
		t.Errorf("snapshot has %d players, want 2", len(s.Players))
	}

	unsubscribe()
	if _, err := r.Join("Zoé"); err != nil {
		t.Fatal(err)
	}
	select {
	case s := <-updates:
		t.Errorf("received %+v after unsubscribe", s)
	default:
	}
}
