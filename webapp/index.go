package webapp

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var content embed.FS

var root, _ = fs.Sub(content, "dist")

var Handler = http.FileServer(http.FS(root))
