package main

import (
	"net/http"
	"log"
	"html/template"
	"regexp"
)

var (
	templates = template.Must(template.ParseFiles("views/index.html"))
	validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)$")
)

func renderTemplate(w http.ResponseWriter, tmpl string) {
	if err := templates.ExecuteTemplate(w, tmpl+".html", nil); err != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index")
}
func main() {
	http.HandleFunc("/", indexHandler)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("ListenAndSErve:", err)
	}
}
