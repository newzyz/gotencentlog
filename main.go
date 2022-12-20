package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	s "github.com/newzyz/storage"
	cls "github.com/tencentcloud/tencentcloud-cls-sdk-go"
)

func LogService(content map[string]string) {

	producerConfig := cls.GetDefaultAsyncProducerClientConfig()
	producerConfig.Endpoint = os.Getenv("CLS_ENDPOINT")
	producerConfig.AccessKeyID = os.Getenv("COS_SECRETID")
	producerConfig.AccessKeySecret = os.Getenv("COS_SECRETKEY")
	topicId := os.Getenv("TOPIC_ID")

	producerInstance, err := cls.NewAsyncProducerClient(producerConfig)
	if err != nil {
		// t.Error(err)
		log.Fatalln(err)
	}

	//Sender Async，Start
	producerInstance.Start()

	var m sync.WaitGroup
	callBack := &Callback{}
	m.Add(1)
	go func() {
		defer m.Done()
		log := cls.NewCLSLog(time.Now().Unix(), content)
		err = producerInstance.SendLog(topicId, log, callBack)
		if err != nil {
			fmt.Println(err)
		}
	}()

	m.Wait()
	producerInstance.Close(10000)
}

func Upload(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(32 << 20)
	var buf bytes.Buffer

	_, fhs, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}

	s.UploadSingle("single/", fhs)

	buf.Reset()

	LogService(map[string]string{"content": fhs.Filename + " has been uploaded"})

}

func UploadMulti(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(32 << 20) // 32MB is the default used by FormFile
	fhs := r.MultipartForm.File["file"]
	s.UploadMulti("multi/", fhs)

	LogService(map[string]string{"content": "Upload Multi has been success"})

}

func Delete(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	filename, ok := vars["filename"]
	if !ok {
		fmt.Println("filename is missing in parameters")
	}

	s.DeleteTC("single/" + filename)

	LogService(map[string]string{"content": "Delete Completed"})
}

func Download(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	filename, ok := vars["filename"]
	if !ok {
		fmt.Println("filename is missing in parameters")
	}

	s.DownloadTC("single/" + filename)

	LogService(map[string]string{"content": "Download Success"})
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
	err2 := http.ListenAndServe(":8089", router)
	if err2 != nil {
		log.Fatalln("There's an error with the server", err2)
	}

}

type Callback struct {
}

func (callback *Callback) Success(result *cls.Result) {
	attemptList := result.GetReservedAttempts()
	for _, attempt := range attemptList {
		fmt.Printf("%+v \n", attempt)
	}
}

func (callback *Callback) Fail(result *cls.Result) {
	fmt.Println(result.IsSuccessful())
	fmt.Println(result.GetErrorCode())
	fmt.Println(result.GetErrorMessage())
	fmt.Println(result.GetReservedAttempts())
	fmt.Println(result.GetRequestId())
	fmt.Println(result.GetTimeStampMs())
}
