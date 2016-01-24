package controllers

import (
	"net/http"
	"strconv"

	"github.com/gophergala2016/gophertron/models"
	"github.com/gorilla/websocket"
)

func Create(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	height, err := strconv.Atoi(values.Get("height"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	width, err := strconv.Atoi(values.Get("width"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	needed, err := strconv.Atoi(values.Get("needed"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = models.NewField(height, width, needed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func Join(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	field, ok := models.GetGame(id)
	if !ok {
		http.Error(w, "Couldn't find game", 404)
		return
	}

	gopher := models.NewGopher()
	index, err := field.Add(gopher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	go listener(conn, index, field)
	go sendPath(conn, gopher.Paths, gopher.Close)
}
