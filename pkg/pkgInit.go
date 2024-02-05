package pkg

import (
	"fmt"
	"os"
)

const (
	cataloguePath = "./catalogue/" // mp3 file storage
	outputPath    = "./hls/" // directory for segmented music
)

// func that calls when package imported
func init() {
    err := os.Mkdir(cataloguePath, os.ModePerm)
    if err != nil {
        fmt.Println("Catalogue already exists")
    }
    err = os.Mkdir(outputPath, os.ModePerm)
    if err != nil {
        fmt.Println("Hls directory already exists")
    }
}
