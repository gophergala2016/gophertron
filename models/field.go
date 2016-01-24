package models

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

type ChangeDirection struct {
	Index     int
	Direction Direction
	Wait      *sync.WaitGroup
}

const (
	Initializing = iota
	InProgress
	Ended
)

type Field struct {
	mu *sync.Mutex //initially used while adding players

	ID            string
	Height, Width int
	needed        int //number of players needed to start the game

	Gophers []*Gopher
	Board   [][]bool

	Change chan ChangeDirection
	Remove chan int

	State     int
	cycles    int
	nextCycle interface{}
}

var (
	ErrInProgress = errors.New("game is already in progress")
	ErrMaxPlayers = errors.New("cannot add more than 4 players")
	mapMu         = new(sync.RWMutex)
	activeFields  = make(map[string]*Field)
)

func GetGames() map[string]*Field {
	games := make(map[string]*Field)
	mapMu.RLock()
	defer mapMu.RUnlock()

	for id, field := range activeFields {
		games[id] = field
	}

	return games
}

func GetGame(id string) (*Field, bool) {
	mapMu.RLock()
	defer mapMu.RUnlock()
	field, ok := activeFields[id]

	return field, ok
}

func NewField(height, width int, needed int) (*Field, error) {
	if needed > 4 {
		return nil, ErrMaxPlayers
	}

	bytes := make([]byte, 10)
	rand.Read(bytes)
	id := base64.URLEncoding.EncodeToString(bytes)

	field := &Field{
		ID:        id,
		Height:    height,
		needed:    needed,
		Width:     width,
		Board:     make([][]bool, height),
		Change:    make(chan ChangeDirection),
		Remove:    make(chan int),
		mu:        new(sync.Mutex),
		State:     Initializing,
		nextCycle: time.NewTicker(100 * time.Millisecond),
	}

	for i := range field.Board {
		field.Board[i] = make([]bool, width)
	}

	mapMu.Lock()
	defer mapMu.Unlock()
	activeFields[id] = field

	return field, nil
}

func (f *Field) Add(g *Gopher) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.State == InProgress {
		return 0, ErrInProgress
	}

	f.Gophers = append(f.Gophers, g)
	//set gopher's initial position and direction

	var initPositions = []struct {
		X int
		Y int
		D Direction
	}{
		{f.Width / 2, 0, North},
		{0, f.Height / 2, East},
		{f.Width / 2, f.Height - 1, South},
		{f.Width - 1, f.Height / 2, West},
	}
	for _, pos := range initPositions {
		if !f.Board[pos.X][pos.Y] {
			f.setPos(g, pos.X, pos.Y)
			g.Direction = pos.D
			break
		}
	}

	if len(f.Gophers) == f.needed {
		f.State = Ended
		go f.start()
	}

	return len(f.Gophers) - 1, nil
}

func (f *Field) setPos(g *Gopher, X, Y int) {
	f.Board[X][Y] = true
	g.X = X
	g.Y = Y
}

func (f *Field) remove(gopher *Gopher) {
	for i, currGopher := range f.Gophers {
		if currGopher == gopher {
			f.Gophers, f.Gophers[len(f.Gophers)-1] = append(f.Gophers[:i], f.Gophers[i+1:]...), nil
		}
	}
}

func (f *Field) PrintBoard() {
	for y := f.Height - 1; y >= 0; y-- {
		for x := 0; x < f.Width; x++ {
			if f.Board[x][y] {
				fmt.Print(1)
			} else {
				fmt.Print(0)
			}
			//fmt.Printf("(%d,%d)", x, y)
		}
		fmt.Println()
	}
	fmt.Println("-----")
}

func (f *Field) increment(g *Gopher) bool {
	g.Path = append(g.Path, Coordinate{g.X, g.Y})
	switch g.Direction {
	case North:
		g.Y++
	case South:
		g.Y--
	case East:
		g.X++
	case West:
		g.X--
	}

	if g.X == f.Height || g.Y == f.Width {
		return true
	}
	if f.Board[g.X][g.Y] {
		return true
	}

	f.Board[g.X][g.Y] = true
	g.Score++

	if f.cycles > 10 {
		f.Board[g.Path[0].X][g.Path[0].Y] = false
	} else {
		f.cycles++
	}

	return false
}

func (f *Field) clearPath(g *Gopher) {
	for _, c := range g.Path {
		f.Board[c.X][c.Y] = false
	}
}

func (f *Field) start() {
	log.Printf("%s: Starting main game loop.", f.ID)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case dir := <-f.Change:
			if f.State != InProgress {
				continue
			}
			f.Gophers[dir.Index].Direction = dir.Direction
			dir.Wait.Done()
		case index := <-f.Remove:
			f.remove(f.Gophers[index])
			if len(f.Gophers) == 1 || len(f.Gophers) == 0 {
				f.end()
				return
			}
		case <-tick.C:
			for _, gopher := range f.Gophers {
				if f.increment(gopher) {
					//fmt.Println(i, " collided")
					//gopher collided, clear it's path and remove it
					//from the field
					f.clearPath(gopher)
					f.remove(gopher)

					if len(f.Gophers) == 1 {
						f.end()
						return
					}
				}
			}
			f.broadcast()
			//f.PrintBoard()
		}
	}
}

func (f *Field) end() {
	log.Printf("%s: ended", f.ID)
	f.mu.Lock()
	f.State = Ended
	f.mu.Unlock()
	if _, ok := f.nextCycle.(*time.Ticker); ok {
		f.nextCycle.(*time.Ticker).Stop()
	}
	mapMu.Lock()
	delete(activeFields, f.ID)
	mapMu.Unlock()
}

func (f *Field) broadcast() {
	paths := make(map[string][]Coordinate)
	for i, gopher := range f.Gophers {
		index := strconv.Itoa(i)
		for _, c := range gopher.Path {
			paths[index] = append(paths[index], c)
		}

		gopher.Paths <- paths
	}
}
