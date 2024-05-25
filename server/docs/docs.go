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

// swagger:parameters uploadSong
type UploadSongRequest struct {
	// Song data
	Data []byte `json:"data"`
}

// swagger:response statusOk
type StatusOk struct {
	//Code
	Message string `json:"message"`
	//Message
	Code int `json:"code"`
}

// swagger:response seeOther
type SeeOther struct {
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

// swagger:response playResponse
type PlayResponse struct {
	//Content
	//in:body
	Content []byte `json:"content"`
}

// swagger:response fetchResponse
type FetchResponse struct {
	//id
	//in:query
	Id int `json:"id"`
	//type
	Type string `json:"type"`
}

// swagger:response segmentResponse
type SegmentResponse struct {
	//segment
	//in:body
	Data []byte `json:"data"`
}

// swagger:parameters addUser
type FormValues struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Name     string `json:"name"`
	IsAdmin  bool   `json:"is_admin"`
}
