package controllers

import (
	"encoding/json"
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

func sendPath(conn *websocket.Conn, ID int, gopher *models.Gopher, field *models.Field) {
	send := make(chan []byte)
	wait := new(sync.WaitGroup)

	go func() {
		for {
			msg := <-send
			go func() {
				conn.WriteMessage(websocket.TextMessage, msg)
				wait.Done()
			}()
		}
	}()

	//initial keepalive loop
	go func() {
		_, _, err := conn.ReadMessage()
		if err != nil {
			field.PreGameRemove(gopher)
			return
		}
		listener(conn, ID, field)
	}()

	for {
		select {
		case paths := <-gopher.Paths:
			wait.Add(1)
			bytes, _ := json.Marshal(paths)
			send <- bytes
		case notify := <-gopher.Notify:
			wait.Add(1)
			send <- []byte(notify)
		case victory := <-gopher.Close:
			wait.Wait()
			if victory {
				wait.Add(1)
				send <- []byte("victory")
			}
			conn.Close()
			return
		}
	}
}
