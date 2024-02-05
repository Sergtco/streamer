package pkg

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)


// To access the song url should look like: http://localhost:8080/get?song=song_name
//
// 'song_name' - the actual name of song to stream.
//
// Server will response with m3u8 file.
func ServeSong(w http.ResponseWriter, r *http.Request) {
	songName := r.URL.Query().Get("song")

	err := generateHLS(songName)
	if err != nil {
		fmt.Println("Error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, outputPath+songName+"/"+songName+".m3u8")
}

func generateHLS(songName string) error {
    err := os.MkdirAll(outputPath+songName, os.ModePerm)
    if err != nil {
        return fmt.Errorf("Failed to create hls directory")
    }
	inputFile := filepath.Join(cataloguePath, songName+".mp3")
	outputM3U8 := filepath.Join(outputPath+songName+"/", songName+".m3u8")

	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-c:a", "aac",
		"-b:a", "128k",
		"-hls_time", "10",
		"-hls_segment_filename", outputPath+songName+"/"+"segment_%03d.ts",
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

// To access the segment of song url should look like: http://localhost:8080/segments/segment_xxx.ts?song=song_name
//
// 'xxx' - the actual segment number.
// 'song_name' - the actual name of song to stream.
//
// Server will response with .ts file.
func ServeTS(w http.ResponseWriter, r *http.Request) {
	songName := r.URL.Query().Get("song")
    segmentFilename := strings.TrimPrefix(r.URL.Path, "/segments/")
	http.ServeFile(w, r, outputPath+songName+"/"+segmentFilename)
}
