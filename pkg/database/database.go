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
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const dataBasePath = "./database.db" // database with music info

var database *sql.DB

type Song struct {
	id     int
	name   string
	artist sql.NullString
	album  string
    path string
}

type Artist struct {
	id   int
	name string
}
type Album struct {
	id     int
	name   string
	artist string
	cover  string
}

/*
Function that initializes the database.
*/
func initDatabase(dataBasePath string) error {
	if _, err := os.Stat(dataBasePath); os.IsNotExist(err) {
		os.Create(dataBasePath)
	}
	var err error
	database, err = sql.Open("sqlite3", dataBasePath)
    // defer database.Close()
    data, err := os.ReadFile("./migrations.sql")
    if err != nil {
        return err
    }
    for _, query := range strings.Split(string(data), ";") {
        _, err := database.Exec(query + ";")
        if err != nil {
            return err
        } 
    }
	return err
}

/*
Returns the row with the given id of song.

	`id` - integer id of song
*/
func getSong(id int) {
}

/*
Returns all rows with songs.
*/
func getAll() ([]Song, error) {
    rows, err := database.Query("SELECT id, name, artist, album, path FROM songs")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var songs []Song
    for rows.Next() {
        var song Song
        // it's better to validate artist field later (on client, or while transfering)
        err := rows.Scan(&song.id, &song.name, &song.artist, &song.album, &song.path)
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
func getByArtist(artist string) ([]Song, error) {
    if artist == "Unknown" {
        return nil, errors.New("No such artist")
    }
    // TODO
    return nil, nil
}

/*
Returns all rows with specific album.
`album` - string of album's name.
*/
func getByAlbum(album string) {
}
