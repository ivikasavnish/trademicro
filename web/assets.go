package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed index.html static
var embeddedFiles embed.FS

// Assets returns the embedded file system as an http.FileSystem
func GetFileSystem() http.FileSystem {
	fsys, err := fs.Sub(embeddedFiles, ".")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}
