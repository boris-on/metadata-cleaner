package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func cleanMetadata(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(10 << 25)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	tempFile, err := ioutil.TempFile("temp-docs", "upload-*.docx")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	tempFile.Write(fileBytes)
	fmt.Fprintf(w, "Successfully Uploaded File\n")

	// doc, err := document.Open("document.docx")
	// if err != nil {
	// 	log.Fatalf("error opening document: %s", err)
	// }
	// defer doc.Close()

	// cp := doc.CoreProperties
}

func home(w http.ResponseWriter, r *http.Request) {

}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/", home)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	fmt.Println("starting server at", port)
	http.ListenAndServe(":"+port, mux)
}
