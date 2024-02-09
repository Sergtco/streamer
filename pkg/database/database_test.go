package database

import (
	// "os"
	"fmt"
	"testing"
	"time"
)

func TestInitDatabase(t *testing.T) {
    s := time.Now()
    err := ReinitDatabase(dataBasePath)
    e := time.Now()
    fmt.Println(e.UnixMilli() - s.UnixMilli())
    if err != nil {
        panic(err)
    }
    // os.Remove(dataBasePath)
}
