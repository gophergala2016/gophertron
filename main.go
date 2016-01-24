package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gophergala2016/gophertron/routes"
	"github.com/pkg/profile"
)

var addr = flag.String("http", "localhost:8080", "http service address")
var cpuprof = flag.Bool("cpuprofile", false, "cpu profile")

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	if *cpuprof {
		log.Println("Profiling CPU usage")
		defer profile.Start().Stop()
	}

	routes.InitRoutes()
	log.Printf("Serving on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, http.DefaultServeMux))
}
