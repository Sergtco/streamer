package main

import (
	"fmt"
	"net/http"
	"os"
)

func serve(w http.ResponseWriter, r *http.Request) {
    file, err := os.ReadFile("./sound.mp3")
    if err != nil {
        panic("Fuck")
    }
    fmt.Fprint(w, file)
}

func main() {
    // os.WriteFile("decoded.mp3", file, 0644)
    // fmt.Println("Done!")
    http.HandleFunc("/get", serve)
    if http.ListenAndServe(":8080", nil) == nil {
        fmt.Println("Exited!")
    }
}
