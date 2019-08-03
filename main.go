package main

import (
	"html/template"
	"log"
	"net/http"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, "")
}

func main() {
	http.HandleFunc("/", IndexHandler)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
