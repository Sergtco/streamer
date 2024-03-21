package main

import (
	"fmt"
	"net/http"
	"stream/pkg"
)

func main() {
	addr, port := "127.0.0.1", "8080"
	http.HandleFunc("/get", pkg.ServeSong)
	http.HandleFunc("/fetch", pkg.FetchDB)
	http.HandleFunc("/segments/", pkg.ServeTS)
	http.HandleFunc("/getSongData", pkg.GetSongData)
	http.HandleFunc("/deleteSong", pkg.DeleteHandler)
	fmt.Printf("Listening on %s:%s\n", addr, port)
	if http.ListenAndServe(fmt.Sprintf("%s:%s", addr, port), nil) == nil {
		fmt.Println("Exited")
	}
}
