package models

type Direction int

const (
	North Direction = 1 << iota
	South
	East
	West
)

type Coordinate struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type Gopher struct {
	// Current direction
	Direction Direction
	X, Y      int
	Path      []Coordinate
	Score     int
	Paths     chan map[string][]Coordinate
	Close     chan struct{}
}

func NewGopher() *Gopher {
	return &Gopher{
		Paths: make(chan map[string][]Coordinate),
		Close: make(chan struct{}),
	}
}
