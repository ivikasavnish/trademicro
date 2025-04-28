//go:build dev
// +build dev

package web

import (
	"net/http"
	"os"
	"path/filepath"
)

// Assets contains the project's web assets.
var Assets http.FileSystem

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// In development mode, use the actual directory
	webDir := filepath.Join(wd, "web")
	Assets = http.Dir(webDir)
}
