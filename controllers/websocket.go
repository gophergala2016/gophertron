package controllers

import (
	"log"
	"sync"
	"time"

	"github.com/gophergala2016/gophertron/models"
	"github.com/gorilla/websocket"
)

var dirMap = map[string]models.Direction{
	"north": models.North,
	"south": models.South,
	"east":  models.East,
	"west":  models.West,
}

func listener(conn *websocket.Conn, ID int, field *models.Field) {
	tick := time.NewTicker(5 * time.Millisecond)
	for {
		select {
		case <-tick.C:
			var req struct {
				Request string
				Param   string
			}

			err := conn.ReadJSON(&req)
			if err != nil {
				field.Remove <- ID
				log.Println(err)
				return
			}

			switch req.Request {
			case "move":
				dir, ok := dirMap[req.Param]
				if !ok {
					continue
				}

				wait := new(sync.WaitGroup)
				wait.Add(1)
				field.Change <- models.ChangeDirection{
					Index:     ID,
					Direction: dir,
					Wait:      wait,
				}
				wait.Wait()
			}
		}
	}
}

func sendPath(conn *websocket.Conn, paths chan map[int][]models.Coordinate, close chan struct{}) {
	mu := new(sync.Mutex)

	for {
		select {
		case paths := <-paths:
			go func() {
				mu.Lock()
				conn.WriteJSON(paths)
				mu.Unlock()
			}()
		case <-close:
			return
		}
	}
}
