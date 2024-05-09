package structs

type Song struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Path   string `json:"path"`
}

type Artist struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type Album struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Cover  string `json:"cover"`
}

type User struct {
	Id       int
	Name     string
	Login    string
	Password string
	IsAdmin  bool
}
