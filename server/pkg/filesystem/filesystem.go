package filesystem

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"stream/pkg/structs"
	"strings"

	"github.com/dhowden/tag"
)

var CataloguePath = os.Getenv("CATALOGUE")
var SupportedFormats = []string{"mp3", "flac", "wav"}

/*
Scans music directory for music files and returns slice with Song structure
*/
func ScanFs() ([]structs.Song, error) {
	path := CataloguePath
	musicPaths := make([]string, 0)
	if err := getAllFilePaths(path, &musicPaths); err != nil {
		return nil, fmt.Errorf("Unable to get filepaths: %v", err)
	}
	songs, err := convertToSongs(musicPaths)
	if err != nil {
		log.Print(err)
	}
	return songs, nil
}

/*
Gets all filepaths in path of catalogue and writes in `output`
*/
func getAllFilePaths(root string, output *[]string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isMusic(info.Name()) {
			*output = append(*output, path)
		}
		return nil
	})
	return err
}

/*
Checks if file is music file.

It checks supported formats from `SupportedFormats` variable.
*/
func isMusic(filename string) bool {
	splitted := strings.Split(filename, ".")
	if format := strings.ToLower(splitted[len(splitted)-1]); slices.Contains(SupportedFormats, format) {
		return true
	}
	return false
}

func convertToSongs(paths []string) ([]structs.Song, error) {
	output := make([]structs.Song, 0)
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("Error opening file %s: %v", path, err)
		}
		data, err := tag.ReadFrom(file)
		if err != nil {
			splitted := strings.Split(path, "/")
			output = append(output, structs.Song{
				Id:     -1,
				Name:   splitted[len(splitted)-1],
				Artist: "Unknown",
				Album:  splitted[len(splitted)-1],
				Path:   path,
			})
			log.Print(fmt.Errorf("Error reading file metadata: %s\n", path))
			continue
		}
		output = append(output, structs.Song{
			Id:     -1,
			Name:   data.Title(),
			Artist: data.Artist(),
			Album:  data.Album(),
			Path:   path,
		})
	}
	return output, nil
}

func ConvertToSong(path string) (structs.Song, error) {
    var song structs.Song
    file, err := os.Open(path)
    if err != nil {
        return song, fmt.Errorf("Error opening file %s: %v", path, err)
    }
    data, err := tag.ReadFrom(file)
    if err != nil {
        splitted := strings.Split(path, "/")
        song = structs.Song{
            Id:     -1,
            Name:   splitted[len(splitted)-1],
            Artist: "Unknown",
            Album:  splitted[len(splitted)-1],
            Path:   path,
        }
        log.Print(fmt.Errorf("Error reading file metadata: %s\n", path))
    } else {
        song = structs.Song{
            Id:     -1,
            Name:   data.Title(),
            Artist: data.Artist(),
            Album:  data.Album(),
            Path:   path,
        }
    }
	return song, nil
}
