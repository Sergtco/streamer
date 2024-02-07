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
    "stream/pkg"
	_ "github.com/mattn/go-sqlite3"
)

const dataBasePath = "./database.db" // database with music info

var database *sql.DB


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
func getSong(id int) (pkg.Song, error) {
    var song pkg.Song
    statement, err := database.Prepare("SELECT id, name, artist, album, path FROM songs WHERE id = ?")
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
    statement, err := database.Prepare("SELECT id, name FROM artists WHERE id = ?")
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
    statement, err := database.Prepare("SELECT id, name, artist FROM albums WHERE id = ?")
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
    rows, err := database.Query("SELECT id, name, artist, album, path FROM songs")
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
