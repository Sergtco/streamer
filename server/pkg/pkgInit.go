package pkg

import (
	"fmt"
	"log"
	"os"
	"stream/config"
	"stream/pkg/database"
)

var (
	CataloguePath = os.Getenv("CATALOGUE")
	outputPath    = os.Getenv("HLS")
	DataBasePath  = os.Getenv("DB_PATH")
)

// func that calls when package imported
func init() {
	err := os.Mkdir(CataloguePath, os.ModePerm)
	if err != nil {
		log.Println("Catalogue already exists")
	}
	err = os.Mkdir(outputPath, os.ModePerm)
	if err != nil {
		log.Println("Hls directory already exists")
	}
	config.InitEnv()
	if err = database.ReinitDatabase(); err != nil {
		log.Fatalf("Unable to reinitialize database: %v", err)
	}
}
