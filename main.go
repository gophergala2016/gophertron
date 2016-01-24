package main

import (
	"flag"
	"log"
	"net/http"

	db "github.com/gophergala2016/gophertron/models/database"
	"github.com/gophergala2016/gophertron/routes"
)

var addr = flag.String("http", "localhost:8080", "http service address")
var prof = flag.Bool("prof", false, "enable's profiling with gom")

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
	db.SetupDB()

	if *prof {
		log.Println("Enabling pprof")
		go func() { log.Fatal(http.ListenAndServe("localhost:6060", nil)) }()
	}

	mux := http.NewServeMux()
	routes.InitRoutes(mux)
	log.Printf("Serving on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
