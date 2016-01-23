package models

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type ChangeDirection struct {
	Gopher    *Gopher
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

	Height, Width int
	needed        int //number of players needed to start the game

	Gophers []*Gopher
	Board   [][]bool

	Change chan ChangeDirection
	Remove chan *Gopher

	State     int
	cycles    int
	nextCycle interface{}
}

var (
	ErrInProgress = errors.New("game is already in progress")
	ErrMaxPlayers = errors.New("cannot add more than 4 players")
)

func New(height, width int, needed int) (*Field, error) {
	if needed > 4 {
		return nil, ErrMaxPlayers
	}

	field := &Field{
		Height:    height,
		needed:    needed,
		Width:     width,
		Board:     make([][]bool, height),
		Change:    make(chan ChangeDirection),
		Remove:    make(chan *Gopher),
		mu:        new(sync.Mutex),
		State:     Initializing,
		nextCycle: time.NewTicker(100 * time.Millisecond),
	}

	for i := range field.Board {
		field.Board[i] = make([]bool, width)
	}

	return field, nil
}

func (f *Field) Add(g *Gopher) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.State == InProgress {
		return ErrInProgress
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

	return nil
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
	channel := make(chan struct{})

	switch f.nextCycle.(type) {
	case time.Ticker:
		go func() {
			for {
				<-f.nextCycle.(time.Ticker).C
				channel <- struct{}{}
			}
		}()
	default:
		//testing
		go func() {
			for {
				channel <- <-f.nextCycle.(chan struct{})
			}
		}()
	}

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
		case <-channel:
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
			//f.PrintBoard()
		}
	}
}

func (f *Field) end() {
	f.mu.Lock()
	f.State = Ended
	f.mu.Unlock()
	if _, ok := f.nextCycle.(*time.Ticker); ok {
		f.nextCycle.(*time.Ticker).Stop()
	}
}

func (f *Field) broadcast() {

}
