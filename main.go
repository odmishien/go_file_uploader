package main

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
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

func UploadToS3(file io.Reader, fileName string) (*s3manager.UploadOutput, error) {
	bucket := "odmishienbucket"
	uploader := GetS3UploadManager()
	output, err := uploader.Upload(&s3manager.UploadInput{
		ACL:    aws.String("public-read"),
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		log.Fatal("upload s3: ", err)
		return nil, err
	}
	return output, nil
}

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

	fb, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	buf := bytes.NewBuffer(fb)
	output, err := UploadToS3(buf, handler.Filename)

	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	img := Image{
		FileName: handler.Filename,
		FilePath: output.Location}

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
