package routes

import (
	"html/template"
	"net/http"

	"github.com/gophergala2016/gophertron/controllers"
	"github.com/gophergala2016/gophertron/models"
	_ "github.com/rakyll/gom/http"
)

func Main(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./views/active.html"))
	tmpl.Execute(w, models.GetGames())
}

func InitRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", Main)
	mux.HandleFunc("/create", controllers.Create)
	mux.HandleFunc("/join", controllers.Join)
	mux.HandleFunc("/game", controllers.Game)
	mux.HandleFunc("/websocket", controllers.WebSocket)
}
