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
	router.HandleFunc("GET /admin", pkg.ValidateJwt(http.HandlerFunc(pkg.AdminIndex)))
	router.HandleFunc("GET /admin/login", pkg.AdminLogin)
	router.HandleFunc("POST /admin/login", pkg.CheckAdminLogin)

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
