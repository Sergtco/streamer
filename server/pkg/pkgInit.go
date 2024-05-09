package pkg

import (
	"crypto/sha256"
	"log"
	"os"
	"stream/config"
	"stream/pkg/database"
	"strings"
)

var (
	CataloguePath = os.Getenv("CATALOGUE")
	outputPath    = os.Getenv("HLS")
	DataBasePath  = os.Getenv("DB_PATH")
)

// func that is called when package imported
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
	if err = database.InitDatabase(); err != nil {
		log.Fatalf("Unable to initialize database: %v", err)
	}

	password := sha256.Sum256([]byte("password"))

	var passwordString strings.Builder
	for i := range password {
		passwordString.WriteByte(password[i])
	}
	_, err = database.InsertUser("Sergtco", "Sergtco", passwordString.String(), 1)
}
