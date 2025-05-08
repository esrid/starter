package handlers

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"template/web"
)

var (
	HTMLFS, _  = fs.Sub(web.FileFS, "templates")
	FrameFS, _ = fs.Sub(web.FileFS, "templates/frame")
)

func renderFrame(w http.ResponseWriter, frame string, data any) error {
	t := template.Must(template.ParseFS(FrameFS, frame))
	return t.Execute(w, data)
}

func unprocessable(w http.ResponseWriter) { w.WriteHeader(http.StatusUnprocessableEntity) }

func badRequest(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}

func internal(w http.ResponseWriter, err error) {
	log.Println(err.Error())
	http.Error(w, "internal", http.StatusInternalServerError)
}

func conflict(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusConflict)
}

// return true if string is == ""
func shouldNotBeNull(w http.ResponseWriter, s ...string) bool {
	for _, v := range s {
		if v == "" {
			unprocessable(w)
			return true
		}
	}
	return false
}
