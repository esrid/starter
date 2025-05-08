package web

import "embed"

//go:embed templates/**/*.html
//go:embed static/css/*.css
//go:embed static/js/*.js
var FileFS embed.FS
