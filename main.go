package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func GetS3UploadManager() *s3manager.Uploader {
	var AWS_ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY")
	var AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
	creds := credentials.NewStaticCredentials(AWS_ACCESS_KEY, AWS_SECRET_ACCESS_KEY, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String("us-east-2")},
	)
	if err != nil {
		log.Fatal(err)
	}
	return s3manager.NewUploader(sess)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, "")
}

func UploadToS3(file io.Reader, fileName string) error {
	bucket := "odmishienbucket"
	uploader := GetS3UploadManager()
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		log.Fatal("upload s3: ", err)
		return err
	}
	return nil
}

func MakeTempFile(file io.Reader, fileName string) error {
	f, err := os.OpenFile("imgs/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
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

	err = MakeTempFile(file, handler.Filename)
	if err != nil {
		log.Fatal(err)
	}
	tf, err := os.Open("imgs/" + handler.Filename)
	err = UploadToS3(tf, handler.Filename)

	if err != nil {
		log.Fatal(err)
	}

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
