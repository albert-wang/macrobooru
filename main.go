package main

import (
	"log"
	"net/http"
	"strings"
)

func handleRequest(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request to %s", r.URL.Path)

		path := r.URL.Path
		if path != "/macro" {
			w.WriteHeader(404)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(404)
			return
		}

		uploadedId, err := CreateMacro(cfg, r.FormValue("image"), strings.ToUpper(r.FormValue("top")), strings.ToUpper(r.FormValue("bottom")))
		if err != nil {
			log.Print(err)
			w.WriteHeader(502)
			return
		}

		w.Write([]byte(uploadedId))
		w.WriteHeader(200)
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	config, _ := LoadConfigurationFile("config.json")

	log.Printf("Listening on %s", config.BindAddr)
	err := http.ListenAndServe(config.BindAddr, handleRequest(config))
	if err != nil {
		log.Print(err)
	}
}
