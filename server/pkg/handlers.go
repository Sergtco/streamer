package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"stream/pkg/database"
	"stream/pkg/filesystem"
	"strings"
)

type Response struct {
	Message string `json:"message"`
}

// To access the song url should look like: http://localhost:8080/get?song=song_id
//
// 'song_id' - the id of song (TODO)
//
// Server will response with m3u8 file.
func ServeSong(w http.ResponseWriter, r *http.Request) {
	songId := r.URL.Query().Get("song")

	err := generateHLS(songId)
	if err != nil {
		fmt.Println("Error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	http.ServeFile(w, r, outputPath+songId+"/"+songId+".m3u8")
}

func FetchDB(w http.ResponseWriter, r *http.Request) {
	songs, err := database.GetAllSongs()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print("Unable to get songs: ", err)
	}
	res, _ := json.Marshal(songs)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}


func generateHLS(songId string) error {
	err := os.MkdirAll(outputPath+songId, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to create hls directory")
	}
	id, err := strconv.Atoi(songId)
	if err != nil {
		return fmt.Errorf("Failed to read song id.")
	}
	song, err := database.GetSong(id)
	if err != nil {
		return fmt.Errorf("Failed to get song.")
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
		"-hls_base_url", "/segments/?song="+songId+"&"+"file=",
		outputM3U8,
	)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("FFmpeg command failed: %w", err)
	}

	return nil
}

// To access a segment of song url should look like: http://localhost:8080/segments/segment_xxx.ts?song=song_id
//
// 'xxx' - the actual segment number.
// 'song_id' - the actual id of song to stream.
//
// Server will response with .ts file.
func ServeTS(w http.ResponseWriter, r *http.Request) {
	songId := r.URL.Query().Get("song")
	fileName := r.URL.Query().Get("file")
	fmt.Println(r.URL.Query().Encode())
	fmt.Println(r.URL.Path)
	fmt.Println(fileName)
	http.ServeFile(w, r, outputPath+songId+"/"+fileName)
}

// Handler to get song data by id
//
// To get a song, url should look like: http://localhost:8080/getSongData?song=song_id
//
// 'song_id' - the actual id of song.
//
// Server will response with JSON.
func GetSongData(w http.ResponseWriter, r *http.Request) {
	songId, err := strconv.Atoi(r.URL.Query().Get("song"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	song, err := database.GetSong(songId)
	json, err := json.Marshal(song)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

/*
Handler for uploading song

It supports only .mp3 file format (for a while)
*/
func UploadHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	// TODO: do not rebuilding database!
	err = database.ReinitDatabase()
	if err != nil {
		http.Error(w, "Eroor rebuilding database", http.StatusInternalServerError)
	}

	response := Response{Message: "Song uploaded successfully"}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

/*
Handler for deleting song by id
*/
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	songId, err := strconv.Atoi(r.URL.Query().Get("song"))
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
		http.Error(w, fmt.Sprintf("Error deleting song: %s", err), http.StatusInternalServerError)
		return
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
`name` - filename
*/
func isMusic(name string) bool {
	splitted := strings.Split(name, ".")
	if splitted[len(splitted)-1] == "mp3" {
		return true
	}
	return false
}
