package models

import (
	"errors"
	"sync"
	"time"
)

type ChangeDirection struct {
	Gopher    *Gopher
	Direction Direction
	Wait      *sync.WaitGroup
}

type Field struct {
	Height, Width int
	needed        int //number of players needed to start the game

	Gophers []*Gopher
	Board   [][]bool

	Change chan ChangeDirection
	Remove chan *Gopher

	mu         *sync.RWMutex
	inProgress bool

	cycles int
}

var ErrInProgress = errors.New("Game is already in progress.")

func New(height, width int) *Field {
	field := &Field{
		Height: height,
		Width:  width,
		Board:  make([][]bool, height),
		Change: make(chan ChangeDirection),
		Remove: make(chan *Gopher),
		mu:     new(sync.RWMutex),
	}

	for i := range field.Board {
		field.Board[i] = make([]bool, width)
	}

	return field
}

func (f *Field) Add(g *Gopher) error {
	f.mu.RLock()

	if f.inProgress {
		f.mu.RUnlock()
		return ErrInProgress
	}
	f.mu.RUnlock()

	f.Gophers = append(f.Gophers, g)
	if len(f.Gophers) == f.needed {
		f.mu.Lock()
		f.inProgress = true
		f.mu.Unlock()
		go f.start()
	}

	return nil
}

func (f *Field) remove(gopher *Gopher) {
	for i, currGopher := range f.Gophers {
		if currGopher == gopher {
			f.Gophers, f.Gophers[len(f.Gophers)-1] = append(f.Gophers[:i], f.Gophers[i+1:]...), nil
		}
	}
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
	tick := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case dir := <-f.Change:
			dir.Gopher.Direction = dir.Direction
			dir.Wait.Done()
		case gopher := <-f.Remove:
			f.remove(gopher)
			if len(f.Gophers) == 1 {
				f.end()
				return
			}
		case <-tick.C:
			for _, gopher := range f.Gophers {
				if f.increment(gopher) {
					//gopher collided, clear it's path and remove it
					//from the field
					f.clearPath(gopher)
					f.remove(gopher)

					if len(f.Gophers) == 1 {
						f.end()
						return
					}
				}

				f.broadcast()
			}
		}
	}
}

func (f *Field) end() {

}

func (f *Field) broadcast() {

}
