package routes

import (
	"net/http"

	"github.com/gophergala2016/gophertron/controllers"
)

func Main(w http.ResponseWriter, r *http.Request) {

}

func InitRoutes() {
	http.HandleFunc("/", Main)
	http.HandleFunc("/create", controllers.Create)
	http.HandleFunc("/join", controllers.Join)
}
