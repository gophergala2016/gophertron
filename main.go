package main

import (
	"flag"
	"net/http"

	"log"

	"github.com/gophergala2016/gophertron/routes"
)

var addr = flag.String("http", "localhost:8080", "http service address")

func main() {
	routes.InitRoutes()
	log.SetFlags(log.Lshortfile)
	log.Println("Serving on ", *addr)
	http.ListenAndServe(*addr, http.DefaultServeMux)
}
