package main

import (
	"log"
	"net/http"
	"os"
	"stream/pkg"
	"stream/pkg/admin"

	"github.com/gorilla/handlers"
)

func main() {

	log.SetFlags(log.LstdFlags)

	router := http.NewServeMux()
	router.HandleFunc("/get/{id}", pkg.ServeSong)
	router.HandleFunc("/segments/{song}/{file}", pkg.ServeTS)
	router.HandleFunc("/getSongData/{song}", pkg.GetSongData)
	router.HandleFunc("DELETE /deleteSong/{song}", pkg.DeleteHandler)
	// Admin for browser
	router.HandleFunc("GET /admin/login", admin.AdminLogin)
	router.HandleFunc("POST /admin/login", admin.CheckAdminLogin)
	router.HandleFunc("GET /admin", admin.ValidateJwt(http.HandlerFunc(admin.AdminIndex)))
	router.HandleFunc("GET /admin/songs", admin.ValidateJwt(http.HandlerFunc(admin.ListSongs)))
	router.HandleFunc("POST /admin/add_user", admin.ValidateJwt(http.HandlerFunc(admin.AddUser)))
	router.HandleFunc("POST /admin/change_user", admin.ValidateJwt(http.HandlerFunc(admin.ChangeUser)))
	router.HandleFunc("POST /admin/delete_user", admin.ValidateJwt(http.HandlerFunc(admin.DeleteUser)))

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
