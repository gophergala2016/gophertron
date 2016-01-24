package routes

import (
	"net/http"
	_ "net/http/pprof"

	"html/template"

	"github.com/gophergala2016/gophertron/controllers"
	"github.com/gophergala2016/gophertron/models"
)

func Main(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./views/active.html"))
	tmpl.Execute(w, models.GetGames())
}

func InitRoutes() {
	http.HandleFunc("/", Main)
	http.HandleFunc("/create", controllers.Create)
	http.HandleFunc("/join", controllers.Join)
	http.HandleFunc("/game", controllers.Game)
	http.HandleFunc("/websocket", controllers.WebSocket)
}
