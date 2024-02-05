package database

import (
	// "os"
	"testing"
)

func TestInitDatabase(t *testing.T) {
    err := initDatabase(dataBasePath)
    if err != nil {
        panic(err)
    }
    // os.Remove(dataBasePath)
}
