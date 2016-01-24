package controllers

import (
	"log"
	"sync"
	"time"

	"github.com/gophergala2016/gophertron/models"
	"github.com/gorilla/websocket"
)

var dirMap = map[string]models.Direction{
	"up":    models.Up,
	"down":  models.Down,
	"left":  models.Left,
	"right": models.Right,
}

func listener(conn *websocket.Conn, ID int, field *models.Field) {
	tick := time.NewTicker(5 * time.Millisecond)
	for {
		select {
		case <-tick.C:
			var req struct {
				Request string `json:"request"`
				Param   string `json:"param"`
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

func sendPath(conn *websocket.Conn, paths chan map[string]models.GopherInfo, close chan bool) {
	mu := new(sync.Mutex)

	for {
		select {
		case paths := <-paths:
			go func() {
				mu.Lock()
				err := conn.WriteJSON(paths)
				if err != nil {
					log.Println(err)
				}
				mu.Unlock()
			}()
		case <-close:
			conn.Close()
			return
		}
	}
}
