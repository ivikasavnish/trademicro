//go:build generate
// +build generate

package web

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	fs := http.Dir("./")
	err := vfsgen.Generate(fs, vfsgen.Options{
		PackageName:  "web",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln("Failed to generate assets:", err)
	}
}
