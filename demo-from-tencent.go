package main

// import (
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"

// 	cls "github.com/tencentcloud/tencentcloud-cls-sdk-go"
// )

// func main() {
// 	producerConfig := cls.GetDefaultAsyncProducerClientConfig()
// 	producerConfig.Endpoint = ""
// 	producerConfig.AccessKeyID = ""
// 	producerConfig.AccessKeySecret = "
// 	topicId := ""
// 	producerInstance, err := cls.NewAsyncProducerClient(producerConfig)
// 	if err != nil {
// 		// t.Error(err)
// 		log.Fatalln(err)
// 	}

// 	// Sender Async，Start
// 	producerInstance.Start()

// 	var m sync.WaitGroup
// 	callBack := &Callback{}
// 	for i := 0; i < 10; i++ {
// 		m.Add(1)
// 		go func() {
// 			defer m.Done()
// 			for i := 0; i < 10; i++ {
// 				log := cls.NewCLSLog(time.Now().Unix(), map[string]string{"content": "hello world| I'm new from Bangkok", "content2": fmt.Sprintf("%v", i)})
// 				err = producerInstance.SendLog(topicId, log, callBack)
// 				if err != nil {
// 					fmt.Println(err)
// 				}
// 			}
// 		}()
// 	}
// 	m.Wait()
// 	producerInstance.Close(60000)
// }

// type Callback struct {
// }

// func (callback *Callback) Success(result *cls.Result) {
// 	attemptList := result.GetReservedAttempts()
// 	for _, attempt := range attemptList {
// 		fmt.Printf("%+v \n", attempt)
// 	}
// }

// func (callback *Callback) Fail(result *cls.Result) {
// 	fmt.Println(result.IsSuccessful())
// 	fmt.Println(result.GetErrorCode())
// 	fmt.Println(result.GetErrorMessage())
// 	fmt.Println(result.GetReservedAttempts())
// 	fmt.Println(result.GetRequestId())
// 	fmt.Println(result.GetTimeStampMs())
// }
