package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	mathrand "math/rand"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Mode is the way questions are asked in a network game.
type Mode int

const (
	// ModeTurns asks each player a question in turn, like the terminal game.
	ModeTurns Mode = iota
	// ModeRace asks the same question to everybody: the first good answer wins.
	ModeRace
)

func (m Mode) String() string {
	if m == ModeRace {
		return "Le plus rapide"
	}
	return "Chacun son tour"
}

// Phase is the state of a room.
type Phase string

const (
	PhaseLobby    Phase = "lobby"    // players are joining
	PhaseQuestion Phase = "question" // a question is waiting for answers
	PhaseReveal   Phase = "reveal"   // the answer is shown before the next question
	PhaseFinished Phase = "finished" // results are shown
)

const (
	maxNameLength       = 20
	defaultQuestionTime = 30 * time.Second
	defaultRevealTime   = 3 * time.Second
)

// Errors are shown as is to the players.
var (
	ErrEmptyName       = errors.New("il faut un prénom")
	ErrNameTooLong     = errors.New("ce prénom est trop long")
	ErrNameTaken       = errors.New("ce prénom est déjà pris")
	ErrAlreadyStarted  = errors.New("la partie a déjà commencé")
	ErrNoPlayer        = errors.New("personne n'a encore rejoint la partie")
	ErrUnknownPlayer   = errors.New("joueur inconnu")
	ErrNoQuestion      = errors.New("pas de question en cours")
	ErrNotYourTurn     = errors.New("ce n'est pas ton tour")
	ErrAlreadyAnswered = errors.New("tu as déjà répondu")
)

// Room is a network game. It is safe for concurrent use; every change is
// broadcast to subscribers as a Snapshot.
type Room struct {
	mu sync.Mutex

	deck         *Deck
	mode         Mode
	nbQuestions  int // per player
	rng          *mathrand.Rand
	questionTime time.Duration
	revealTime   time.Duration
	// schedule runs f after d; tests replace it to control time.
	schedule func(d time.Duration, f func())

	phase    Phase
	players  []*Player
	ids      map[string]*Player
	round    int // index of the current question
	total    int // number of questions in the game
	question Question
	turn     *Player // ModeTurns: the player who must answer
	answered map[*Player]bool
	deadline time.Time
	reveal   *Reveal

	subscribers map[chan Snapshot]struct{}
}

func NewRoom(deck *Deck, mode Mode, nbQuestions int) *Room {
	return &Room{
		deck:         deck,
		mode:         mode,
		nbQuestions:  nbQuestions,
		rng:          mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		questionTime: defaultQuestionTime,
		revealTime:   defaultRevealTime,
		schedule:     func(d time.Duration, f func()) { time.AfterFunc(d, f) },
		phase:        PhaseLobby,
		ids:          map[string]*Player{},
		subscribers:  map[chan Snapshot]struct{}{},
	}
}

// Join adds a player to the lobby and returns the id they must use to answer.
func (r *Room) Join(name string) (string, error) {
	name = strings.TrimSpace(name)
	switch {
	case name == "":
		return "", ErrEmptyName
	case utf8.RuneCountInString(name) > maxNameLength:
		return "", ErrNameTooLong
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.phase != PhaseLobby {
		return "", ErrAlreadyStarted
	}
	for _, player := range r.players {
		if strings.EqualFold(player.Name, name) {
			return "", ErrNameTaken
		}
	}

	id := newID()
	player := &Player{Name: name}
	r.players = append(r.players, player)
	r.ids[id] = player
	r.broadcast()
	return id, nil
}

func (r *Room) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.phase != PhaseLobby {
		return ErrAlreadyStarted
	}
	if len(r.players) == 0 {
		return ErrNoPlayer
	}

	r.total = r.nbQuestions
	if r.mode == ModeTurns {
		r.total *= len(r.players)
	}
	if r.total == 0 {
		r.phase = PhaseFinished
	} else {
		r.askQuestion()
	}
	r.broadcast()
	return nil
}

// Answer records the answer of a player and tells whether it is right.
func (r *Room) Answer(id, answer string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	player, ok := r.ids[id]
	switch {
	case !ok:
		return false, ErrUnknownPlayer
	case r.phase != PhaseQuestion:
		return false, ErrNoQuestion
	case r.mode == ModeTurns && player != r.turn:
		return false, ErrNotYourTurn
	case r.answered[player]:
		return false, ErrAlreadyAnswered
	}

	r.answered[player] = true
	correct := strings.TrimSpace(answer) == r.deck.Answer(r.question)
	if correct {
		player.GoodAnswered = append(player.GoodAnswered, r.question)
		r.endQuestion(player)
	} else {
		player.BadAnswered = append(player.BadAnswered, r.question)
		if r.mode == ModeTurns || len(r.answered) == len(r.players) {
			r.endQuestion(nil)
		}
	}
	r.broadcast()
	return correct, nil
}

// askQuestion draws the next question. The caller must hold the lock.
func (r *Room) askQuestion() {
	r.phase = PhaseQuestion
	r.question = r.deck.RandomQuestion(r.rng)
	r.answered = map[*Player]bool{}
	r.reveal = nil
	if r.mode == ModeTurns {
		r.turn = r.players[r.round%len(r.players)]
	}
	r.deadline = time.Now().Add(r.questionTime)

	round := r.round
	r.schedule(r.questionTime, func() { r.timeout(round) })
}

// timeout ends the question if nobody found the answer in time. In ModeTurns,
// not answering counts as a failure.
func (r *Room) timeout(round int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.phase != PhaseQuestion || r.round != round {
		return
	}
	if r.mode == ModeTurns {
		r.turn.BadAnswered = append(r.turn.BadAnswered, r.question)
	}
	r.endQuestion(nil)
	r.broadcast()
}

// endQuestion shows the answer and who found it. The caller must hold the lock.
func (r *Room) endQuestion(finder *Player) {
	r.phase = PhaseReveal
	r.reveal = &Reveal{Answer: r.deck.Answer(r.question)}
	if finder != nil {
		r.reveal.Finder = finder.Name
	}

	round := r.round
	r.schedule(r.revealTime, func() { r.next(round) })
}

// next moves on to the following question, or ends the game.
func (r *Room) next(round int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.phase != PhaseReveal || r.round != round {
		return
	}
	r.round++
	if r.round >= r.total {
		r.phase = PhaseFinished
	} else {
		r.askQuestion()
	}
	r.broadcast()
}

// Subscribe returns a channel receiving the current snapshot, then a new one
// after each change. Only the latest snapshot is kept if the reader is slow.
func (r *Room) Subscribe() (<-chan Snapshot, func()) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan Snapshot, 1)
	ch <- r.snapshot()
	r.subscribers[ch] = struct{}{}

	unsubscribe := func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.subscribers, ch)
	}
	return ch, unsubscribe
}

func (r *Room) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.snapshot()
}

// Players returns the players, for the final results.
func (r *Room) Players() []*Player {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*Player(nil), r.players...)
}

// broadcast sends the current snapshot to every subscriber, replacing any
// snapshot not read yet. The caller must hold the lock.
func (r *Room) broadcast() {
	snapshot := r.snapshot()
	for ch := range r.subscribers {
		select {
		case <-ch:
		default:
		}
		ch <- snapshot
	}
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
