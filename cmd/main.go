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

	"github.com/barasher/go-exiftool"
)

func deleteMetadata(filename string) error {

	e, _ := exiftool.NewExiftool()
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
			fmt.Println(err)
		}
	}
}

func serveFile(w http.ResponseWriter, multiFile *multipart.FileHeader, zipWriter *zip.Writer) {
	file, err := multiFile.Open()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	tempFile, err := os.Create("metawipe/" + multiFile.Filename)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	err = deleteMetadata(tempFile.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Println("opening file...")
	f, err := os.Open(tempFile.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer f.Close()

	fmt.Println("writing file to archive...")
	fileWriter, err := zipWriter.Create(tempFile.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	if _, err := io.Copy(fileWriter, f); err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 25)

	formdata := r.MultipartForm
	files := formdata.File["files[]"]

	fmt.Println("creating zip archive...")
	tempArchive, err := ioutil.TempFile("metawipe", "zip-*.zip")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer tempArchive.Close()
	zipWriter := zip.NewWriter(tempArchive)

	for _, file := range files {
		serveFile(w, file, zipWriter)
	}

	fmt.Println("closing zip archive...")
	zipWriter.Close()

	fileBytes, err := ioutil.ReadFile(tempArchive.Name())
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	fmt.Println(tempArchive.Name())
	io.Copy(w, bytes.NewReader(fileBytes))

}

func main() {
	mkDir()

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("public"))
	mux.Handle("/", fs)
	mux.HandleFunc("/upload", upload)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	fmt.Println("starting server at", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		fmt.Println(err)
	}
}
