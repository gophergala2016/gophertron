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
	Needed        int //number of players needed to start the game

	Gophers []*Gopher
	Board   [][]bool

	Change chan ChangeDirection
	Remove chan int

	State  int
	cycles int
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
		ID:     id,
		Height: height,
		Needed: needed,
		Width:  width,
		Board:  make([][]bool, height),
		Change: make(chan ChangeDirection),
		Remove: make(chan int),
		mu:     new(sync.Mutex),
		State:  Initializing,
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
		{f.Width / 2, 0, Down},
		{0, f.Height / 2, Right},
		{f.Width / 2, f.Height - 1, Up},
		{f.Width - 1, f.Height / 2, Left},
	}
	for _, pos := range initPositions {
		if !f.Board[pos.X][pos.Y] {
			f.setPos(g, pos.X, pos.Y)
			g.Direction = pos.D
			break
		}
	}

	if len(f.Gophers) == f.Needed {
		f.State = InProgress
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
	gopher.Close <- false
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
	case Up:
		g.Y--
	case Down:
		g.Y++
	case Left:
		g.X--
	case Right:
		g.X++
	}

	if (g.X == -1 || g.X == f.Width) || (g.Y == -1 || g.Y == f.Height) {
		//collision detected
		return true
	}

	if f.Board[g.X][g.Y] {
		return true
	}

	f.Board[g.X][g.Y] = true
	g.Score++

	if f.cycles > 100 {
		f.Board[g.Path[0].X][g.Path[0].Y] = false
		g.Path = append(g.Path[:0], g.Path[1:]...)
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
	mapMu.Lock()
	delete(activeFields, f.ID)
	mapMu.Unlock()

	for {
		select {
		case dir := <-f.Change:
			if f.State != InProgress {
				dir.Wait.Done()
				continue
			}
			currDir := f.Gophers[dir.Index].Direction
			if (currDir == Up || currDir == Down) && (dir.Direction == Up || dir.Direction == Down) {
				dir.Wait.Done()
				continue
			}
			if (currDir == Left || currDir == Right) && (dir.Direction == Left || dir.Direction == Right) {
				dir.Wait.Done()
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

	if len(f.Gophers) != 0 {
		f.Gophers[0].Close <- true
	}
}

var colors = []string{"#b71c1c", "#880E4F", "#4A148C", "#1A237E"}

type GopherInfo struct {
	Coordinate []Coordinate `json:"coordinate"`
	Color      string       `json:"color"`
}

func (f *Field) broadcast() {
	paths := make(map[string]GopherInfo)
	for i, gopher := range f.Gophers {
		index := strconv.Itoa(i)
		var coordinates []Coordinate
		for _, c := range gopher.Path {
			coordinates = append(coordinates, c)
		}

		paths[index] = GopherInfo{coordinates, colors[i]}
		gopher.Paths <- paths
	}
}
