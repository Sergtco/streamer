package pkg

import (
	"fmt"
	"os"
	"stream/config"
	"stream/pkg/database"
)

var (
	CataloguePath string
	outputPath string
    DataBasePath string
)

// func that calls when package imported
func init() {
	CataloguePath = os.Getenv("CATALOGUE")
	outputPath    = os.Getenv("HLS")
    DataBasePath = os.Getenv("DB_PATH")
    fmt.Println(CataloguePath, outputPath, DataBasePath)

    err := os.Mkdir(CataloguePath, os.ModePerm)
    if err != nil {
        fmt.Println("Catalogue already exists")
    }
    err = os.Mkdir(outputPath, os.ModePerm)
    if err != nil {
        fmt.Println("Hls directory already exists")
    }
    config.InitEnv()
    if err = database.InitDatabase(); err != nil {
        panic(fmt.Errorf("Unable to initialize database: %v", err))
    }
}
