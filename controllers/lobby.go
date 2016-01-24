package controllers

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gophergala2016/gophertron/models"
	"github.com/gorilla/websocket"
)

func Create(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	height, err := strconv.Atoi(values.Get("height"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
	if needed == 1 || needed == 0 {
		return
	}

	field, err := models.NewField(height, width, needed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/join?id="+field.ID, http.StatusTemporaryRedirect)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func Join(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	//log.Println(id)
	_, ok := models.GetGame(id)
	if !ok {
		http.Error(w, "Couldn't find game", 404)
		return
	}

	cookie := http.Cookie{
		Name:    "game-id",
		Value:   id,
		Expires: time.Now().Add(10 * time.Second),
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/game", http.StatusTemporaryRedirect)
}

func Game(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("game-id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	field, ok := models.GetGame(cookie.Value)
	if !ok {
		http.Error(w, "Couldn't find game", 404)
		return
	}

	templ := template.Must(template.ParseFiles("./views/game.html"))
	templ.Execute(w, map[string]interface{}{
		//multiply by ten to get a bigger canvas
		"height": field.Height * 10,
		"width":  field.Width * 10,
	})
}

func WebSocket(w http.ResponseWriter, r *http.Request) {
	//log.Println("reached")
	cookie, err := r.Cookie("game-id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//log.Println(cookie)
	field, ok := models.GetGame(cookie.Value)
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
	if len(field.Gophers) == field.Needed {
		conn.WriteMessage(websocket.TextMessage, []byte("notification"))
	}

	go listener(conn, index, field)
	go sendPath(conn, index, gopher, field)
}
