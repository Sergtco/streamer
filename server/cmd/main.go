package main

import (
	"log"
	"net/http"
	"os"
	"stream/pkg"

	"github.com/gorilla/handlers"
)

func main() {

	log.SetFlags(log.LstdFlags)

	router := http.NewServeMux()
	router.HandleFunc("/get/{id}", pkg.ServeSong)
	router.HandleFunc("/segments/{song}/{file}", pkg.ServeTS)
	router.HandleFunc("/getSongData/{song}", pkg.GetSongData)
	router.HandleFunc("DELETE /deleteSong/{song}", pkg.DeleteHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: handlers.LoggingHandler(os.Stdout, router),
	}

	log.Printf("Listening on %s \n", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Errror on server %w", err)
	}
}
