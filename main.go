package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, "")
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	type Image struct {
		FileName string
		FilePath string
	}

	if r.Method != "POST" {
		http.Error(w, "POST is only allowed.", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "File size is too large", http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()
	f, err := os.OpenFile("imgs/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer f.Close()
	io.Copy(f, file)
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	img := Image{
		FileName: handler.Filename,
		FilePath: "/imgs/" + handler.Filename}

	tmpl.Execute(w, img)
}

func main() {
	http.Handle("/imgs/", http.StripPrefix("/imgs/", http.FileServer(http.Dir("imgs"))))
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/upload", UploadHandler)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
