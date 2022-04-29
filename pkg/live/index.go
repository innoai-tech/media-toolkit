package live

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var content embed.FS

var root, _ = fs.Sub(content, "dist")

var WebUI = http.FileServer(http.FS(root))
