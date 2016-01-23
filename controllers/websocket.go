package controllers

import (
	"log"
	"sync"

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
	for {
		var req struct {
			Request string
			Param   string
		}

		err := conn.ReadJSON(&req)
		if err != nil {
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
