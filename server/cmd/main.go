package main

import (
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
	"stream/pkg"
)

func main() {

	log.SetFlags(log.LstdFlags)

	router := http.NewServeMux()
	router.HandleFunc("/get/{id}", pkg.ServeSong)
	router.HandleFunc("/segments/{song}/{file}", pkg.ServeTS)
	router.HandleFunc("/getSongData/{song}", pkg.GetSongData)
	router.HandleFunc("DELETE /deleteSong/{song}", pkg.DeleteHandler)
	router.HandleFunc("GET /admin", pkg.AdminIndex)
	router.HandleFunc("GET /login", pkg.Login)
	router.HandleFunc("POST /login", pkg.Validate)

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
