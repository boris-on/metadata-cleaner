package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/barasher/go-exiftool"
	"github.com/sirupsen/logrus"
)

func deleteMetadata(filename string) error {

	e, err := exiftool.NewExiftool()
	if err != nil {
		logrus.Error(err)
	}
	defer e.Close()
	originals := e.ExtractMetadata(filename)

	originals[0].ClearAll()
	e.WriteMetadata(originals)
	return nil
}

func mkDir() {
	path := "metawipe"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			logrus.Error(err)
		}
	}
}

func serveFile(w http.ResponseWriter, multiFile *multipart.FileHeader, zipWriter *zip.Writer) {
	file, err := multiFile.Open()
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}
	defer file.Close()

	tempFile, err := os.Create("metawipe/" + multiFile.Filename)
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}

	err = deleteMetadata(tempFile.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}

	f, err := os.Open(tempFile.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}
	defer f.Close()

	fileWriter, err := zipWriter.Create(tempFile.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}
	if _, err := io.Copy(fileWriter, f); err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}

	err = os.Remove(tempFile.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}

}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 25)

	formdata := r.MultipartForm
	files := formdata.File["files[]"]

	tempArchive, err := ioutil.TempFile("metawipe", "zip-*.zip")
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}
	defer tempArchive.Close()
	zipWriter := zip.NewWriter(tempArchive)

	for _, file := range files {
		logrus.WithFields(logrus.Fields{
			"filename":       file.Filename,
			"remote address": r.RemoteAddr,
		}).Info("Serving file")
		serveFile(w, file, zipWriter)
	}

	zipWriter.Close()

	fileBytes, err := ioutil.ReadFile(tempArchive.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}
	io.Copy(w, bytes.NewReader(fileBytes))

	err = os.Remove(tempArchive.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		logrus.Error(err)
		return
	}
}

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	mkDir()

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("public"))
	mux.Handle("/", fs)
	mux.HandleFunc("/upload", upload)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	logrus.Info("Starting server at port: ", port)
	go http.ListenAndServe(":"+port, mux)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Printf("Server shutting down")

}
