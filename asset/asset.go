package asset

import (
	"embed"
	"text/template"
)

//go:embed share.html
var files embed.FS

//go:embed index.html
var IndexHTML []byte

var (
	ShareTemplate = template.Must(template.ParseFS(files, "share.html"))
)
