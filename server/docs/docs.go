// Package classification stream.
//
//	 Schemes: http
//	 BasePath: /
//	 Version: 1.0.0
//
//	 Consumes:
//	 - application/json
//
//	 Produces:
//	 - application/json
//
//	 Security:
//	 - basic
//
//	SecurityDefinitions:
//	basic:
//	  type: basic
//
// swagger:meta
package docs

// swagger:parameters addToPlaylist
type AddToPlaylistRequest struct {
	//Id of playlist
	//
	// in:path
	PlaylistId int `json:"playlist_id"`
	//Id of song
	//
	// in:path
	SongId int `json:"song_id"`
}

// swagger:response statusOk
type StatusOk struct {
	//Code
	Message string `json:"message"`
	//Message
	Code int `json:"code"`
}

// swagger:response badRequest
type BadRequest struct {
	//Code
	Message string `json:"message"`
	//Message
	Code int `json:"code"`
}

// swagger:response internalServerError
type InternalServerError struct {
	//Code
	Message string `json:"message"`
	//Message
	Code int `json:"code"`
}
