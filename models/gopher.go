package models

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
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
	Paths     chan map[string]GopherInfo
	Notify    chan string
	Close     chan bool
}

func NewGopher() *Gopher {
	return &Gopher{
		Paths:  make(chan map[string]GopherInfo),
		Close:  make(chan bool),
		Notify: make(chan string),
	}
}
