# Usage
To start use server, you should have `catalogue/` and `hls/` directories placed at the root of server directory (directories will be created when the server starts).\
`catalogue/` - used to store your mp3 files.\
`hls/` - used to stream segmented songs.\
To start server run following commands in server root directory:
```sh
go build cmd/main.go
./main
```
