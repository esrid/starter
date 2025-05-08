package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFS(HTMLFS, "pages/index.html"))
	if err := t.Execute(w, nil); err != nil {
		log.Printf("index.html not served")
		return
	}
}

func About(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFS(HTMLFS, "pages/about.html"))
	if err := t.Execute(w, nil); err != nil {
		log.Printf("about.html not served")
		return
	}
}
