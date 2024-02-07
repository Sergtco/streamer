package filesystem


import (
    "testing"
    "stream/pkg"
    "fmt"
)
/*
Should be ok 
*/

func TestGetAllFilePaths(t *testing.T) {
    path := pkg.CataloguePath
    musicPaths := make([]string, 0)
    if err := getAllFilePaths(path, &musicPaths); err != nil {
        panic(fmt.Errorf("Unable to get filepaths %e", err))
    }
}

func TestConvertToSongs(t *testing.T) {
    path := pkg.CataloguePath
    musicPaths := make([]string, 0)
    if err := getAllFilePaths(path, &musicPaths); err != nil {
        panic(fmt.Errorf("Unable to get filepaths %e", err))
    }
    songs, err := convertToSongs(musicPaths)
    if err != nil {
        t.Error(err)
    }
    for _, song := range songs {
        fmt.Printf("%+v\n", song)
    }
}
