package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gophergala2016/gophertron/routes"
)

var addr = flag.String("http", "localhost:8080", "http service address")
var prof = flag.Bool("prof", false, "profile")

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	if *prof {
		log.Println("Enabling pprof")
		go func() { log.Fatal(http.ListenAndServe("localhost:6060", nil)) }()
	}
	log.Println("done")

	mux := http.NewServeMux()
	routes.InitRoutes(mux)
	log.Printf("Serving on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
