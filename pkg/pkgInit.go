package pkg

import (
	"fmt"
	"os"
	"stream/config"
	"stream/pkg/database"
)

var (
	CataloguePath = os.Getenv("CATALOGUE")
	outputPath    = os.Getenv("HLS")
    DataBasePath = os.Getenv("DB_PATH")
)

// func that calls when package imported
func init() {
    err := os.Mkdir(CataloguePath, os.ModePerm)
    if err != nil {
        fmt.Println("Catalogue already exists")
    }
    err = os.Mkdir(outputPath, os.ModePerm)
    if err != nil {
        fmt.Println("Hls directory already exists")
    }
    config.InitEnv()
    if err = database.ReinitDatabase(); err != nil {
        panic(fmt.Errorf("Unable to reinitialize database: %v", err))
    }
}
