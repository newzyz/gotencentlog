package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	s "github.com/newzyz/storage"
)

func Upload(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(32 << 20)
	var buf bytes.Buffer

	_, fhs, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}

	s.UploadSingle("single/", fhs)

	buf.Reset()
}

func UploadMulti(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(32 << 20) // 32MB is the default used by FormFile
	fhs := r.MultipartForm.File["file"]
	s.UploadMulti("multi/", fhs)

}

func Delete(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	filename, ok := vars["filename"]
	if !ok {
		fmt.Println("filename is missing in parameters")
	}

	s.DeleteTC("single/" + filename)

}

func Download(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	filename, ok := vars["filename"]
	if !ok {
		fmt.Println("filename is missing in parameters")
	}

	s.DownloadTC("single/" + filename)

}

func List(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	filename, ok := vars["filename"]
	if !ok {
		fmt.Println("filename is missing in parameters")
	}

	s.ListTC(filename)

}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/upload", Upload).Methods("POST")
	router.HandleFunc("/upload-multi", UploadMulti).Methods("POST")
	router.HandleFunc("/download/{filename}", Download).Methods("GET")
	// router.HandleFunc("/list/{filename}", List).Methods("GET")
	router.HandleFunc("/delete/{filename}", Delete).Methods("DELETE")

	fmt.Println("Server ruuning at port 8089")
	err := http.ListenAndServe(":8089", router)
	if err != nil {
		log.Fatalln("There's an error with the server", err)
	}

}
