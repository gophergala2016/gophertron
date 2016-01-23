package models

import (
	"sync"
)

type ChangeDirection struct {
	Gopher    *Gopher
	Direction Direction
	Wait      *sync.WaitGroup
}

type Field struct {
	Gophers []*Gopher
	Board   [100][100]bool
	Change  chan ChangeDirection
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

	return nil
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

	if g.X == 100 || g.Y == 100 {
		return true
	}
	if f.Board[g.X][g.Y] {
		return true
	}

	f.Board[g.X][g.Y] = true
	return false
}

func (f *Field) clearPath(g *Gopher) {
	for _, c := range g.Path {
		f.Board[c.X][c.Y] = false
	}
}

func (f *Field) Start() {
	for {
		select {
		case dir := <-f.Change:
			dir.Gopher.Direction = dir.Direction
			dir.Wait.Done()
		default:
			for i, gopher := range f.Gophers {
				if f.increment(gopher) {
					//gopher collided, clear it's path and remove it
					//from the field
					f.clearPath(gopher)
					f.Gophers, f.Gophers[len(f.Gophers)-1] = append(f.Gophers[:i], f.Gophers[i+1:]...), nil
					if len(f.Gophers) == 1 {
						f.End()
						return
					}
				}
				f.Broadcast()
			}
		}
	}
}

func (f *Field) End() {

}

func (f *Field) Broadcast() {

}
