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
	"fmt"
	"os"
	"stream/pkg/filesystem"
	"stream/pkg/structs"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var DataBasePath = os.Getenv("DB_PATH")

var Database *sql.DB
var mutex sync.Mutex

/*
For reinitialization of database
*/
func ReinitDatabase() error {
	mutex.Lock()
	defer mutex.Unlock()

	os.Remove(DataBasePath)
	os.Remove(os.Getenv("HLS") + "*") // to prevent id collisions (TODO: more efficient way)
	err := InitDatabase()
	return err
}

/*
Function that initializes the database.
*/
func InitDatabase() error {
	if _, err := os.Stat(DataBasePath); os.IsNotExist(err) {
		os.Create(DataBasePath)
	}
	var err error
	Database, err = sql.Open("sqlite3", DataBasePath)
	data, err := os.ReadFile("./config/migrations.sql")
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

func fillDatabase(songs []structs.Song) error {
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
func GetSong(id int) (structs.Song, error) {
	var song structs.Song
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
func getArtist(id int) (structs.Artist, error) {
	var artist structs.Artist
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
func getAlbum(id int) (structs.Album, error) {
	var album structs.Album
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
func GetAllSongs() ([]structs.Song, error) {
	rows, err := Database.Query("SELECT id, name, artist, album, path FROM songs;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []structs.Song
	for rows.Next() {
		var song structs.Song
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
func getByArtist(id int) ([]structs.Song, error) {
	artist, err := getArtist(id)
	if err != nil {
		return nil, err
	}

	var songs []structs.Song = make([]structs.Song, 0)
	rows, err := Database.Query("SELECT id, name, artist, album, path FROM songs WHERE artist = ?", artist.Name)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var song structs.Song
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
Returns all rows with specific album.
`album` - string of album's name.
*/
func getByAlbum(album string) {
}

func DeleteSong(id int) (structs.Song, error) {
	song, err := GetSong(id)
	if err != nil {
		return song, err
	}
	mutex.Lock()
	defer mutex.Unlock()
	_, err = Database.Exec("DELETE FROM songs WHERE id = ?", id)
	if err != nil {
		return song, err
	}
	return song, nil
}
