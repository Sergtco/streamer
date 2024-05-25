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

// swagger:route POST /add_playlist playlist addPlaylist
// Adds new playlist for user.
// responses:
//
//	200: playlist
//	400: badRequest
//	500: internalServerError
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

	newPlaylist.UserId = user.Id
	playListId, err := database.InsertPlaylist(newPlaylist)
	if err != nil {
		http.Error(w, "Unable to create playlist with this name/songs", http.StatusBadRequest)
	}
	newPlaylist.Id = playListId
	response, err := json.Marshal(newPlaylist)
	if err != nil {
		http.Error(w, "Unable to create playlist", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// swagger:route DELETE /delete_playlist/{playlist_id} playlist deletePlaylist
// Adds new playlist for user.
// responses:
//
//	200: statsuOk
//	400: badRequest
//	500: internalServerError
func DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	playlistId, err := strconv.Atoi(r.PathValue("playlist_id"))
	if err != nil {
		http.Error(w, "Invalid playlist id", http.StatusBadRequest)
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

	err = database.DeletePlaylist(user.Id, playlistId)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:route DELETE /delete_from_playlist/{playlist_id}/{song_id} playlist deleteFromPlaylist
//
// # Adds song song to playlist
//
// responses:
//
//	200: statusOk
//	400: badRequest
//	500: internalServerError
func DeleteFromPlaylist(w http.ResponseWriter, r *http.Request) {
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
	err = database.DeleteFromPlaylist(playlistId, songId)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:route POST /add_to_playlist/{playlist_id}/{song_id} playlist addToPlaylist
//
// # Adds song song to playlist
//
// responses:
//
//	200: statusOk
//	400: badRequest
//	500: internalServerError
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
	w.WriteHeader(200)
	return
}

//swagger:response playlists
type UserPlaylists struct {
	//User's playlists
	Playlists []int `json:"playlists"`
}

// swagger:route GET /get_playlists playlist getPlaylists
//
// Returns all user's playlists.
//
// responses:
//
//	200: playlists
//	400: badRequest
//	500: internalServerError
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
		log.Printf("Error in database: %v", err)
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(playlists)
	if err != nil {
		http.Error(w, "Unable to serialize playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// swagger:route GET /play/{song_id} song play
//
// Gets m3u8 playlist file
// responses:
//
//	200: playResponse
//	400: badRequest
//	500: internalServerError
func Play(w http.ResponseWriter, r *http.Request) {
	songId := r.PathValue("song_id")

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

// swagger:route GET /fetch/{type}?id=id song album artist playlist fetch
//
// type - [song, album, artist, playlist, all]
// `id` - actual id (not needed if type is all)
// fetches some model by id.
// responses:
//
//	200: fetchResponse
//	400: badRequest
//	500: internalServerError
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
	// swagger:route GET /fetch/{type}?id=id song album artist playlist fetch
	//
	// type - [song, album, artist, playlist, all]
	// `id` - actual id (not needed if type is all)
	// fetches some model by id.
	// responses:
	//
	//	200: fetchResponse
	//	400: badRequest
	//	500: internalServerError
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

// swagger:route GET /segments/{song_id}/{file} song file play
//
// Fetches song's segment by its' id.
// Server responds with .ts file.
// responses:
//
//	200: segmentResponse
//	400: badRequest
//	500: internalServerError
func PlaySegment(w http.ResponseWriter, r *http.Request) {
	songId := r.PathValue("song_id")
	fileName := r.PathValue("file")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	http.ServeFile(w, r, outputPath+songId+"/"+fileName)
}

// swagger:route POST /upload_song song file uploadSong
//
// Handler for uploading song
// It supports only .mp3 file format (for a while)
// responses:
//
//	200: statusOk
//	400: badRequest
//	500: internalServerError
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

	if resp.StatusCode != 200 {
		log.Printf("Expected 200 got %d", resp.StatusCode)
		http.Error(w, fmt.Sprintf("Error fetching song features: %s", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// swagger:route DELETE /delete_song/{song_id} file uploadSong
//
// Handler for deleting song by id
// responses:
//
//	200: statusOk
//	400: badRequest
//	500: internalServerError
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	songId, err := strconv.Atoi(r.PathValue("song_id"))
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
		http.Error(w, fmt.Sprintf("Error serializing song id: %s", err), http.StatusInternalServerError)
		return
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
	if resp.StatusCode != 200 {
		log.Printf("Expected 200 got %d", resp.StatusCode)
		http.Error(w, fmt.Sprintf("Error deleting song: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message: Song deleted"}`))
}

//swagger:response radioResponse
type RadioResponse struct {
	//json array of indices of next songs
	//in:body
	Data []byte `json:"data"`
}

// swagger:route GET /radio/{song_id} file uploadSong
//
// Handler for radio
// responses:
//
//	200: radioResponse
//	400: badRequest
//	500: internalServerError
func Radio(w http.ResponseWriter, r *http.Request) {
	songId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	data := map[string]int{"id": songId}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error serializing song id: %v", err)
	}

	model_url := os.Getenv("MODEL_URL")
	req, err := http.NewRequest("POST", "http://"+model_url+":6969/rank_tracks", bytes.NewBuffer(jsonData))
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
	if resp.StatusCode != 200 {
		log.Printf("Expected 200 got %d", resp.StatusCode)
		http.Error(w, fmt.Sprintf("Error ranking songs: %s", err), http.StatusInternalServerError)
		return
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bodyBytes)
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
