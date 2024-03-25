package main

import (
	"fmt"
	"net/http"
	"stream/pkg"
)

func main() {
    router := http.NewServeMux()
    router.HandleFunc("/get/{id}", pkg.ServeSong)
    router.HandleFunc("/segments/{song}/{file}", pkg.ServeTS)
    router.HandleFunc("/getSongData/{song}", pkg.GetSongData)
    router.HandleFunc("DELETE /deleteSong/{song}", pkg.DeleteHandler)

    server := http.Server {
        Addr: ":8080",
        Handler: router,
    }

    fmt.Printf("Listening on %s\n", server.Addr)
    err := server.ListenAndServe()
    if err != nil {
        fmt.Printf("Error listening on %s", server.Addr)
    }
}
