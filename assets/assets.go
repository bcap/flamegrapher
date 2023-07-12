package assets

import "embed"

//go:embed *.css *.js *.html *.json
var FS embed.FS
