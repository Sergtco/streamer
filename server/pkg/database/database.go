package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"stream/pkg/filesystem"
	"stream/pkg/structs"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrorUserExists error = fmt.Errorf("User already exist!")
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
	os.RemoveAll(os.Getenv("HLS"))
	err := InitDatabase()
	return err
}

/*
Function that initializes the database.
Returns fs.ErrExist if dtatabase exists
*/
func InitDatabase() error {
	if _, err := os.Stat(DataBasePath); os.IsNotExist(err) {
		os.Create(DataBasePath)
	} else {
		return fs.ErrExist
	}
	var err error
	Database, err = sql.Open("sqlite3", DataBasePath)
	data, err := os.ReadFile("./config/migrations.sql")
	if err != nil {
		return err
	}
	for _, query := range strings.Split(string(data), "/sp") {
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
	for _, song := range songs {
		err := InsertSong(song)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

/*
Returns the row with the given id of artist.

	`id` - integer id of artist.
*/
func GetArtist(id int) (structs.Artist, error) {
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

func InsertArtist(name string) (int, error) {
	insArtistQuery := "INSERT OR IGNORE INTO artists (name) VALUES (?);"
	selectArtistId := "SELECT id FROM artists WHERE name = ?;"

	_, err := Database.Exec(insArtistQuery, name)
	if err != nil {
		return 0, fmt.Errorf("Errror while inserting artist: %s", err)
	}
	var artId int
	query, err := Database.Query(selectArtistId, name)
	if err != nil {
		return 0, fmt.Errorf("Errror opening query: %s", err)
	}
	defer query.Close()
	query.Next()
	query.Scan(&artId)
	return artId, nil
}

/*
Returns the row with the given id of album.

	`id` - integer id of album.
*/
func GetAlbum(id int) (structs.Album, error) {
	var album structs.Album
	statement, err := Database.Prepare(
		`
		SELECT albums.id, albums.name, artists.name 
		FROM albums 
		LEFT JOIN artists ON albums.artist_id = artists.id
		WHERE albums.id = ?;
		`)
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

func InsertAlbum(name string, artistId int) (int, error) {
	insAlbumQuery := "INSERT OR IGNORE INTO albums (name, artist_id) VALUES (?, ?);"
	selectAlbumId := "SELECT id FROM albums WHERE (name, artist_id) = (?, ?)"
	_, err := Database.Exec(insAlbumQuery, name, artistId)
	if err != nil {
		return 0, fmt.Errorf("Error while inserting Album: %s", err)
	}
	query, err := Database.Query(selectAlbumId, name, artistId)
	if err != nil {
		return 0, fmt.Errorf("Error while selcting album: %s", err)
	}
	defer query.Close()
	var albId int
	query.Next()
	query.Scan(&albId)
	return albId, nil
}

func InsertSong(song structs.Song) error {
	querySongs := "INSERT OR IGNORE INTO songs (name, artist_id, album_id, path) VALUES  (?, ?, ?, ?);"
	artId, err := InsertArtist(song.Artist)
	if err != nil {
		return fmt.Errorf("Error inserting song: %s", err)
	}
	albId, err := InsertAlbum(song.Album, artId)
	if err != nil {
		return fmt.Errorf("Error inserting song: %s", err)
	}
	_, err = Database.Exec(querySongs, song.Name, artId, albId, song.Path)
	if err != nil {
		return fmt.Errorf("Error isnerting song: %s", err)
	}
	return nil
}

/*
Returns the row with the given id of song.

	`id` - integer id of song
*/
func GetSong(id int) (structs.Song, error) {
	var song structs.Song
	statement, err := Database.Prepare(
		`SELECT songs.id, songs.name, artists.name, albums.name, path 
		FROM songs 
		LEFT JOIN  artists ON artists.id = songs.artist_id
		LEFT JOIN  albums ON albums.id = songs.artist_id
		WHERE songs.id = ?;
		`)
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

/*
Returns all rows with songs.
*/
func GetAllSongs() ([]structs.Song, error) {
	rows, err := Database.Query(`SELECT songs.id, songs.name, artists.name, albums.name, path 
		FROM songs
		LEFT JOIN artists ON  artists.id = songs.artist_id
		LEFT JOIN albums ON albums.id = songs.album_id`)
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
/* func getByArtist(id int) ([]structs.Song, error) {
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
} */

/* Returns nil if user does not exist */
func GetUser(login string) (*structs.User, error) {
	rows, err := Database.Query("SELECT id, name, login, password, is_admin FROM users WHERE login = ?", login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	exist := rows.Next()
	if !exist {
		return nil, nil
	}
	user := &structs.User{}
	rows.Scan(&user.Id, &user.Name, &user.Login, &user.Password, &user.IsAdmin)
	return user, nil
}

/*
Inserts user, if user exists returns ErrorUserExists error, else return user Id.
*/
func InsertUser(name, login, password string, isAdmin int) (int, error) {
	exists, err := GetUser(login)
	if err != nil {
		return 0, err
	}
	if exists != nil {
		return 0, ErrorUserExists
	}

	_, err = Database.Exec(`INSERT OR IGNORE INTO users (name, login, password, is_admin) 
		VALUES (?, ?, ?, ?)`, name, login, password, isAdmin)
	if err != nil {
		return 0, err
	}
	rows, err := Database.Query("SELECT id from users WHERE login = ?", login)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	rows.Next()
	var userId int
	rows.Scan(&userId)
	return userId, nil
}

func GetAllUsers() ([]structs.User, error) {
	res := []structs.User{}
	rows, err := Database.Query(`SELECT id, name, login, password, is_admin FROM users;`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		user := structs.User{}
		rows.Scan(&user.Id, &user.Name, &user.Login, &user.Password, &user.IsAdmin)
		res = append(res, user)
	}
	return res, nil
}

func UpdateUser(name, login, password string, IsAdmin bool) error {
	_, err := Database.Exec(`UPDATE users
		SET name = ?, password = ?, is_admin = ?
		WHERE login = ?`, name, password, IsAdmin, login)
	if err != nil {
		return err
	}
	return nil
}

func DeleteUser(login string) error {
	_, err := Database.Exec(`DELETE FROM users WHERE login = ?`, login)
	if err != nil {
		return err
	}
	return nil
}
