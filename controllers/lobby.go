package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strconv"
	"sync"

	"github.com/gophergala2016/gophertron/models"
)

var (
	mu    = new(sync.RWMutex)
	games = make(map[string]*models.Field)
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

	field, err := models.NewField(height, width, needed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bytes := make([]byte, 10)
	rand.Read(bytes)
	id := base64.URLEncoding.EncodeToString(bytes)
	mu.Lock()
	games[id] = field
	mu.Unlock()
}

func Join(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	mu.RLock()
	field, ok := games[id]
	mu.RUnlock()
	if !ok {
		http.Error(w, "Couldn't find lobby.", http.StatusNotFound)
	}

	gopher := &models.Gopher{}
	err := field.Add(gopher)
}
