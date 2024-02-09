package structs

type Song struct {
    Id     int
    Name   string
    Artist string
    Album  string
    Path string
}

type Artist struct {
    Id   int
    Name string
}
type Album struct {
    Id     int
    Name   string
    Artist string
    Cover  string
}
