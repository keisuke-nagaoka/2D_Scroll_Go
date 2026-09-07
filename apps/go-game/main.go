package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

// APP_BASE_PATH is only used for generating links in templates.
// nginx strips the path prefix before forwarding requests, so routing here stays prefix-less.
var basePath = os.Getenv("APP_BASE_PATH")

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

type pageData struct {
	Title    string
	BasePath string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := pageData{Title: "Go Game", BasePath: basePath}
	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", indexHandler)

	log.Println("listening on :8080 (base path for links:", basePath, ")")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
