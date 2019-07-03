package main

import (
	"html/template"
	"log"
	"net/http"

    "github.com/gorilla/mux"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("tmpl/create.html")
	t.Execute(w, nil)
}

func main() {
    r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)

	// Note that the path given to the http.Dir function is relative to the project
	// directory root.
	fileServer := http.FileServer(http.Dir("static/"))
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
