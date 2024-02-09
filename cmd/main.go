package main

import (
	"fmt"
	"net/http"
	"stream/config"
	"stream/pkg"
)

func init() {
    config.InitEnv()
}

func main() {
	http.HandleFunc("/get", pkg.ServeSong)
	http.HandleFunc("/segments/", pkg.ServeTS)
	if http.ListenAndServe(":8080", nil) == nil {
		fmt.Println("Exited")
	}
}
