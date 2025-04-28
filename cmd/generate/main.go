//go:build ignore
// +build ignore

package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
)

func main() {
	webDir := filepath.Join("web")
	fs := http.Dir(webDir)
	err := vfsgen.Generate(fs, vfsgen.Options{
		PackageName:  "web",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln("Failed to generate assets:", err)
	}
}
