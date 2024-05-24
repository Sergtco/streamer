//go:generate swagger generate spec
package main

import (
	"log"
	"net/http"
	"os"
	"stream/pkg"
	"stream/pkg/admin"

	"github.com/gorilla/handlers"
	_ "stream/docs"
)

func main() {

	log.SetFlags(log.LstdFlags)
	router := http.NewServeMux()
	router.HandleFunc("/play/{id}", pkg.Play)
	router.HandleFunc("/segments/{song}/{file}", pkg.PlaySegment)
	router.HandleFunc("/fetch/{type}", admin.ValidateJwt(http.HandlerFunc(pkg.Fetch)))
	router.HandleFunc("DELETE /deleteSong/{song}", admin.ValidateJwt(http.HandlerFunc(pkg.DeleteHandler)))
	router.HandleFunc("POST /add_playlist", admin.ValidateJwt(http.HandlerFunc(pkg.AddPlaylist)))
	router.HandleFunc("POST /add_to_playlist/{playlist_id}/{song_id}", admin.ValidateJwt(http.HandlerFunc(pkg.AddToPlaylist)))
	router.HandleFunc("GET /get_playlists", admin.ValidateJwt(http.HandlerFunc(pkg.GetUserPlaylists)))
	router.HandleFunc("POST /login", admin.UserLogin)
    router.HandleFunc("POST /upload_song", admin.ValidateJwt(http.HandlerFunc(pkg.UploadSong)))
    router.HandleFunc("POST /radio/{id}", admin.ValidateJwt(http.HandlerFunc(pkg.Radio)))
	router.HandleFunc("DELETE /delete_playlist/{id}", admin.ValidateJwt(http.HandlerFunc(pkg.DeletePlaylist)))
	router.HandleFunc("DELETE /delete_from_playlist/{playlist_id}/{song_id}", admin.ValidateJwt(http.HandlerFunc(pkg.DeleteFromPlaylist)))
	// Admin for browser
	router.HandleFunc("GET /admin/login", admin.AdminLogin)
	router.HandleFunc("POST /admin/login", admin.CheckAdminLogin)
	router.HandleFunc("GET /admin", admin.ValidateJwt(http.HandlerFunc(admin.AdminIndex)))
	router.HandleFunc("GET /admin/songs", admin.ValidateJwt(http.HandlerFunc(admin.ListSongs)))
	router.HandleFunc("POST /admin/add_user", admin.ValidateJwt(http.HandlerFunc(admin.AddUser)))
	router.HandleFunc("POST /admin/change_user", admin.ValidateJwt(http.HandlerFunc(admin.ChangeUser)))
	router.HandleFunc("POST /admin/delete_user", admin.ValidateJwt(http.HandlerFunc(admin.DeleteUser)))
	// router.HandleFunc("POST /uploadSong", admin.ValidateJwt(http.HandlerFunc(pkg.UploadSong)))
	router.HandleFunc("POST /upload_song", pkg.UploadSong)

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
