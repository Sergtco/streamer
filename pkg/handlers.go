package pkg

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)


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

	http.ServeFile(w, r, outputPath+songId+"/"+songId+".m3u8")
}

func generateHLS(songId string) error {
    err := os.MkdirAll(outputPath+songId, os.ModePerm)
    if err != nil {
        return fmt.Errorf("Failed to create hls directory")
    }
	inputFile := filepath.Join(cataloguePath, songId+".mp3")
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
		outputM3U8,
	)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("FFmpeg command failed: %w", err)
	}

	return nil
}

// To access the segment of song url should look like: http://localhost:8080/segments/segment_xxx.ts?song=song_id
//
// 'xxx' - the actual segment number.
// 'song_id' - the actual id of song to stream.
//
// Server will response with .ts file.
func ServeTS(w http.ResponseWriter, r *http.Request) {
	songId := r.URL.Query().Get("song")
    segmentFilename := strings.TrimPrefix(r.URL.Path, "/segments/")
	http.ServeFile(w, r, outputPath+songId+"/"+segmentFilename)
}
