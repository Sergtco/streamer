package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"stream/pkg/admin"
	"stream/pkg/database"
	"stream/pkg/filesystem"
	"stream/pkg/structs"
	"strings"
)

var SupportedFormats = []string{"mp3", "flac", "wav"}

// url - /add_playlist
func AddPlaylist(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	claims, err := admin.DecodeLogin(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	user, err := database.GetUser(claims.Login)
	if err != nil || user == nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
	}

	var newPlaylist structs.Playlist
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(bodyBytes, &newPlaylist); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	playListId, err := database.InsertPlaylist(newPlaylist)
	if err != nil {
		http.Error(w, "Unable to create playlist with this name/songs", http.StatusBadRequest)
	}
	newPlaylist.UserId = user.Id
	newPlaylist.Id = playListId
	response, err := json.Marshal(newPlaylist)
	if err != nil {
		http.Error(w, "Unable to create playlist", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// url - /add_to_playlist/playlist_id/song_id
func AddToPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistId, err := strconv.Atoi(r.PathValue("playlist_id"))
	if err != nil {
		http.Error(w, "Invalid playlist id", http.StatusBadRequest)
		return
	}
	songId, err := strconv.Atoi(r.PathValue("song_id"))
	if err != nil {
		http.Error(w, "Invalid song id", http.StatusBadRequest)
		return
	}

	cookie, _ := r.Cookie("token")
	claims, err := admin.DecodeLogin(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	user, err := database.GetUser(claims.Login)
	if err != nil || user == nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	playlistOwner, err := database.GetPlaylistOwner(playlistId)
	if err != nil {
		http.Error(w, "Invalid playlist id", http.StatusBadRequest)
		return
	}

	if playlistOwner != user.Id {
		http.Error(w, "You don't have permission", http.StatusBadRequest)
		return
	}
	database.AddToPlaylist(songId, playlistId)
}

type UserPlaylists struct {
	Playlists []int `json:"playlists"`
}

// url - /get_playlists (must be authorized!)
func GetUserPlaylists(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	claims, err := admin.DecodeLogin(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	user, err := database.GetUser(claims.Login)
	if err != nil || user == nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	playlists, err := database.GetUsersPlaylists(user.Id)
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	userPlaylists := UserPlaylists{Playlists: playlists}
	response, err := json.Marshal(userPlaylists)
	if err != nil {
		http.Error(w, "Unable to serialize playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// To access the song url should look like: http://localhost:8080/play/song_id
//
// 'song_id' - the id of song
//
// Server will respond with m3u8 file.
func Play(w http.ResponseWriter, r *http.Request) {
	songId := r.PathValue("id")

	err := generateHLS(songId)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	http.ServeFile(w, r, outputPath+songId+"/"+songId+".m3u8")
}

// url = /fetch/{type}?id=`id`
// type - [song, album, artist, playlist, all]
// `id` - actual id (not needed if type is all)
func Fetch(w http.ResponseWriter, r *http.Request) {
	t := r.PathValue("type")
	var res []byte
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	switch t {
	case "all":
		songs, err := database.GetAllSongs()
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Internal Server Error, try again later...", http.StatusInternalServerError)
		}
		res, err = json.Marshal(songs)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Internal Server Error, try again later...", http.StatusInternalServerError)
		}
	case "song":
		song, err := database.GetSong(id)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "There's no song with this id", http.StatusBadRequest)
		}
		res, err = json.Marshal(song)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Internal Server Error, try again later...", http.StatusInternalServerError)
		}
	case "album":
		album, err := database.GetAlbum(id)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "There's no song with this id", http.StatusBadRequest)
		}
		res, err = json.Marshal(album)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Internal Server Error, try again later...", http.StatusInternalServerError)
		}
	case "artist":
		artist, err := database.GetArtist(id)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "There's no song with this id", http.StatusBadRequest)
		}
		res, err = json.Marshal(artist)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Internal Server Error, try again later...", http.StatusInternalServerError)
		}
	case "playlist":
		playlist, err := database.GetPlaylist(id)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "There's no playlist with this id", http.StatusBadRequest)
		}
		res, err = json.Marshal(playlist)
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, "Internal Server Error, try again later...", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func generateHLS(songId string) error {
	m3u8Path := filepath.Join(outputPath, songId, songId+".m3u8")
	if _, err := os.Stat(m3u8Path); err == nil {
		// .m3u8 file already exists, return without generating HLS
		return nil
	}

	tsPattern := filepath.Join(outputPath, songId, "segment_*.ts")
	matches, err := filepath.Glob(tsPattern)
	if err != nil {
		return fmt.Errorf("Failed to check existing TS files: %w", err)
	}
	if len(matches) > 0 {
		// TS files already exist, return without generating HLS
		return nil
	}

	err = os.MkdirAll(outputPath+songId, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to create hls directory: %w", err)
	}

	id, err := strconv.Atoi(songId)
	if err != nil {
		return fmt.Errorf("Failed to read song id. %w", err)
	}

	song, err := database.GetSong(id)
	if err != nil {
		return fmt.Errorf("Failed to get song. %w", err)
	}

	inputFile := song.Path
	outputM3U8 := filepath.Join(outputPath+songId+"/", songId+".m3u8")

	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-c:a", "aac",
		"-b:a", "128k",
		"-hls_time", "10",
		"-hls_segment_filename", outputPath+songId+"/"+"segment_%03d.ts",
		"-hls_playlist_type", "event",
		"-hls_list_size", "0",
		"-f", "hls",
		"-hls_base_url", "/segments/"+songId+"/",
		outputM3U8,
	)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("FFmpeg command failed: %w", err)
	}

	return nil
}

// To access a segment of song url should look like: http://localhost:8080/segments/song_id/segment_xxx.ts
//
// 'xxx' - the actual segment number.
// 'song_id' - the actual id of song to stream.
//
// Server will respond with .ts file.
func PlaySegment(w http.ResponseWriter, r *http.Request) {
	songId := r.PathValue("song")
	fileName := r.PathValue("file")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	http.ServeFile(w, r, outputPath+songId+"/"+fileName)
}

/*
Handler for uploading song

It supports only .mp3 file format (for a while)
*/
func UploadSong(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("song")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !isMusic(handler.Filename) {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}

	newFile, err := os.Create(CataloguePath + handler.Filename)
	if err != nil {
		log.Printf("Error creating file: %s \n", err)
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		log.Printf("Error copying file: %s \n", err)
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	// TODO: do not rebuilding database!
	song, err := filesystem.ConvertToSong(CataloguePath + handler.Filename)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}

	songId, err := database.InsertSong(song)
	if err != nil {
		http.Error(w, "Error inserting song in db", http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{"id": songId, "path": CataloguePath + handler.Filename}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error serializing song id: %v", err)
	}

	model_url := os.Getenv("MODEL_URL")
	req, err := http.NewRequest("POST", "http://"+model_url+":6969/mfcc", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error deleting from AI db: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to AI service: %v\n", err)
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	//TODO!!!! ()
	//   if req.Response.StatusCode != 200 {
	//       log.Printf("Expected 200 got %d", req.Response.StatusCode)
	// http.Error(w, fmt.Sprintf("Error fetching song features: %s", err), http.StatusInternalServerError)
	//   }

	w.WriteHeader(http.StatusOK)
}

/*
Handler for deleting song by id
*/
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	songId, err := strconv.Atoi(r.PathValue("song"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	song, err := database.DeleteSong(songId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Can't find song by id: %v", err), http.StatusBadRequest)
		return
	}

	err = filesystem.DeleteFile("./" + song.Path)
	if err != nil {
		log.Printf("Could not delete file: %s", err)
		http.Error(w, fmt.Sprintf("Error deleting song: %s", err), http.StatusInternalServerError)
		return
	}

	data := map[string]int{"id": songId}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error serializing song id: %v", err)
	}

	model_url := os.Getenv("MODEL_URL")
	req, err := http.NewRequest("DELETE", "http://"+model_url+":6969/delete_song", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error deleting from AI db: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to AI service: %v\n", err)
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if req.Response.StatusCode != 200 {
		log.Printf("Expected 200 got %d", req.Response.StatusCode)
		http.Error(w, fmt.Sprintf("Error deleting song: %s", err), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message: Song deleted"}`))
}

/*
Checks if the provided path is considered secure based on specific criteria.

The function returns true if the path is deemed secure, and false otherwise.
*/
func isPathSecure(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	if path[0] == '/' {
		return false
	}
	if !strings.Contains(path, "catalogue") {
		return false
	}
	return true
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
