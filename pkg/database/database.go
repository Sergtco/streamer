package database

/*
song
|id|name|artist|album|path|
artist
|id|name|
album
|id|name|artist|cover|
./artist/album/(song, cover.jpg)
*/

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"stream/pkg"
	"stream/pkg/filesystem"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const dataBasePath = "./database.db" // database with music info

var Database *sql.DB
var mutex *sync.Mutex

/*
For reinitialization of database
*/
func ReinitDatabase(dataBasePath string) error {
    os.Remove(dataBasePath)
    err := InitDatabase(dataBasePath)
    return err
}
/*
Function that initializes the database.
*/
func InitDatabase(dataBasePath string) error {
	if _, err := os.Stat(dataBasePath); os.IsNotExist(err) {
		os.Create(dataBasePath)
	}
	var err error
	Database, err = sql.Open("sqlite3", dataBasePath)
	data, err := os.ReadFile("./migrations.sql")
	if err != nil {
		return err
	}
	for _, query := range strings.Split(string(data), "жопа") {
		_, err := Database.Exec(query)
		if err != nil {
			return err
		}
	}
    songs, err := filesystem.ScanFs()
    if err != nil {
        return err
    }
    err = fillDatabase(songs)
    if err != nil {
        return err
    }
	return nil
}

func fillDatabase(songs []pkg.Song) error {
	if Database == nil {
		return fmt.Errorf("Database is closed.")
	}
	querySongs := "INSERT INTO songs (name, artist, album, path) VALUES  (?, ?, ?, ?);"
	for _, song := range songs {
		Database.Exec(querySongs, song.Name, song.Artist, song.Album, song.Path)
	}
    return nil
}

/*
Returns the row with the given id of song.

	`id` - integer id of song
*/
func getSong(id int) (pkg.Song, error) {
	var song pkg.Song
	statement, err := Database.Prepare("SELECT id, name, artist, album, path FROM songs WHERE id = ?;")
	if err != nil {
		return song, err
	}
	defer statement.Close()

	err = statement.QueryRow(id).Scan(&song.Id, &song.Name, &song.Artist, &song.Album, &song.Path)
	if err != nil {
		return song, err
	}

	return song, nil
}

/*
Returns the row with the given id of artist.

	`id` - integer id of artist.
*/
func getArtist(id int) (pkg.Artist, error) {
	var artist pkg.Artist
	statement, err := Database.Prepare("SELECT id, name FROM artists WHERE id = ?;")
	if err != nil {
		return artist, err
	}
	defer statement.Close()

	err = statement.QueryRow(id).Scan(&artist.Id, &artist.Name)
	if err != nil {
		return artist, err
	}

	return artist, nil
}

/*
Returns the row with the given id of album.

	`id` - integer id of album.
*/
func getAlbum(id int) (pkg.Album, error) {
	var album pkg.Album
	statement, err := Database.Prepare("SELECT id, name, artist FROM albums WHERE id = ?;")
	if err != nil {
		return album, err
	}
	defer statement.Close()

	err = statement.QueryRow(id).Scan(&album.Id, &album.Name, &album.Artist)
	if err != nil {
		return album, err
	}

	return album, nil
}

/*
Returns all rows with songs.
*/
func getAllSongs() ([]pkg.Song, error) {
	rows, err := Database.Query("SELECT id, name, artist, album, path FROM songs;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []pkg.Song
	for rows.Next() {
		var song pkg.Song
		// it's better to validate artist field later (on client, or while transfering)
		err := rows.Scan(&song.Id, &song.Name, &song.Artist, &song.Album, &song.Path)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}
	return songs, nil
}

/*
Returns all rows with specific artist.
`artist` - string of artist's name.
*/
func getByArtist(artist string) ([]pkg.Song, error) {
	if artist == "Unknown" {
		return nil, errors.New("No such artist")
	}

	return nil, nil
}

/*
Returns all rows with specific album.
`album` - string of album's name.
*/
func getByAlbum(album string) {
}
