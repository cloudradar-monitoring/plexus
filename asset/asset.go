package asset

import (
	"embed"
	"text/template"
)

//go:embed *.html
var files embed.FS

var ShareTemplate = template.Must(template.ParseFS(files, "share.html"))
