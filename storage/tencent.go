package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"os"

	"net/http"

	"github.com/joho/godotenv"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
)

type TCConfig struct {
	COSURL    string
	SecretID  string
	SecretKey string
	TCPrefix  string
}

func WithCOSURL(s string) TCConfigOption {
	return func(cfg *TCConfig) {
		cfg.COSURL = s
	}
}
func WithSecretID(s string) TCConfigOption {
	return func(cfg *TCConfig) {
		cfg.SecretID = s
	}
}
func WithSecretKey(s string) TCConfigOption {
	return func(cfg *TCConfig) {
		cfg.SecretKey = s
	}
}
func WithTCPrefix(s string) TCConfigOption {
	return func(cfg *TCConfig) {
		cfg.TCPrefix = s
	}
}

type TCConfigOption func(*TCConfig)

func Options(options ...TCConfigOption) TCConfigOption {
	return func(h *TCConfig) {
		for _, option := range options {
			option(h)
		}
	}
}

type TCStorage struct {
	bucketTC *cos.Client
	Prefix   string
}

func NewTCStorage(options ...TCConfigOption) (*TCStorage, error) {

	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}
	var DefaultPreset = Options(WithCOSURL(os.Getenv("COS_URL")), WithSecretID(os.Getenv("COS_SECRETID")), WithSecretKey(os.Getenv("COS_SECRETKEY")), WithTCPrefix(os.Getenv("TC_PREFIX")))

	cfg := TCConfig{}
	DefaultPreset(&cfg)

	for _, option := range options {
		option(&cfg)
	}

	u, err := url.Parse(cfg.COSURL)
	if err != nil {
		return nil, err
	}
	b := &cos.BaseURL{BucketURL: u}

	// c := cos.NewClient(b, &http.Client{
	// 	Transport: &cos.AuthorizationTransport{
	// 		SecretID:  cfg.SecretID,  // 替换为用户的 SecretId，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
	// 		SecretKey: cfg.SecretKey, // 替换为用户的 SecretKey，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
	// 	},
	// })

	//With Detail
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			// 通过环境变量获取密钥
			// 环境变量 COS_SECRETID 表示用户的 SecretId，登录访问管理控制台查看密钥，https://console.cloud.tencent.com/cam/capi
			SecretID: cfg.SecretID,
			// 环境变量 COS_SECRETKEY 表示用户的 SecretKey，登录访问管理控制台查看密钥，https://console.cloud.tencent.com/cam/capi
			SecretKey: cfg.SecretKey,
			// Debug 模式，把对应 请求头部、请求内容、响应头部、响应内容 输出到标准输出
			Transport: &debug.DebugRequestTransport{
				RequestHeader: true,
				// Notice when put a large file and set need the request body, might happend out of memory error.
				RequestBody:    false,
				ResponseHeader: true,
				ResponseBody:   false,
			},
		},
	})

	return &TCStorage{
		bucketTC: c,
		Prefix:   cfg.TCPrefix,
	}, nil
}

func log_status(err error) {
	if err == nil {
		return
	}
	if cos.IsNotFoundError(err) {
		// WARN
		fmt.Println("WARN: Resource is not existed")
	} else if e, ok := cos.IsCOSError(err); ok {
		fmt.Printf("ERROR: Code: %v\n", e.Code)
		fmt.Printf("ERROR: Message: %v\n", e.Message)
		fmt.Printf("ERROR: Resource: %v\n", e.Resource)
		fmt.Printf("ERROR: RequestId: %v\n", e.RequestID)
		// ERROR
	} else {
		fmt.Printf("ERROR: %v\n", err)
		// ERROR
	}
}

func UploadTC(fn string, t []byte, cont string) {
	// 存储桶名称，由bucketname-appid 组成，appid必须填入，可以在COS控制台查看存储桶名称。 https://console.cloud.tencent.com/cos5/bucket
	// 替换为用户的 region，存储桶region可以在COS控制台“存储桶概览”查看 https://console.cloud.tencent.com/ ，关于地域的详情见 https://cloud.tencent.com/document/product/436/6224 。

	tc, err := NewTCStorage()
	if err != nil {
		panic(err)
	}

	// Case1 Upload Object
	// f := bytes.NewReader(t.Bytes())

	// _, err := tc.bucketTC.Object.Put(context.Background(), name, fn, nil)
	// log_status(err)

	// Case2 Upload Object With Options
	f := bytes.NewReader(t)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: cont,
		},
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			//XCosACL: "public-read",
			XCosACL: "private",
		},
	}
	_, err2 := tc.bucketTC.Object.Put(context.Background(), tc.Prefix+fn, f, opt)
	log_status(err2)

	// // Case3 Upload Object From Local File
	// _, err = tc.bucketTC.Object.PutFromFile(context.Background(), name, "./test", nil)
	// log_status(err)

	// // Case4 Check Upload Progression
	// opt.ObjectPutHeaderOptions.Listener = &cos.DefaultProgressListener{}
	// _, err = tc.bucketTC.Object.PutFromFile(context.Background(), name, "./test", opt)
	// log_status(err)
}

func UploadSingle(preFix string, f *multipart.FileHeader) {

	var buf bytes.Buffer

	file, err := f.Open()
	if err != nil {
		panic(err)
	}
	defer file.Close()

	io.Copy(&buf, file)

	bcon := buf.Bytes()

	UploadTC(preFix+f.Filename, bcon, http.DetectContentType(bcon))

}

func UploadMulti(preFix string, fl []*multipart.FileHeader) {

	for _, fh := range fl {

		var buf bytes.Buffer

		fmt.Println(fh.Filename)
		f, err := fh.Open()
		if err != nil {
			panic(err)
		}

		io.Copy(&buf, f)
		bcon := buf.Bytes()

		UploadTC(preFix+fh.Filename, bcon, http.DetectContentType(bcon))

		f.Close()
		buf.Reset()
	}
}

func ListTC(filename string) {

	tc, err := NewTCStorage()
	if err != nil {
		panic(err)
	}

	opt := &cos.BucketGetOptions{
		Prefix:  filename,
		MaxKeys: 3,
	}
	v, _, err := tc.bucketTC.Bucket.Get(context.Background(), opt)
	if err != nil {
		panic(err)
	}

	for _, c := range v.Contents {
		fmt.Printf("%s, %d\n", c.Key, c.Size)
	}
}

func DownloadTC(filename string) {

	tc, err := NewTCStorage()
	if err != nil {
		panic(err)
	}

	// 1.Get Object From Response Body
	resp, err := tc.bucketTC.Object.Get(context.Background(), tc.Prefix+filename, nil)
	if err != nil {
		panic(err)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	fmt.Printf("%s\n", string(bs))
	// 2. Save Object To Local
	_, err = tc.bucketTC.Object.GetToFile(context.Background(), tc.Prefix+filename, "temp/downloaded.jpg", nil)
	if err != nil {
		panic(err)
	}
}

func DeleteTC(filename string) {
	// 存储桶名称，由bucketname-appid 组成，appid必须填入，可以在COS控制台查看存储桶名称。 https://console.cloud.tencent.com/cos5/bucket
	// 替换为用户的 region，存储桶region可以在COS控制台“存储桶概览”查看 https://console.cloud.tencent.com/ ，关于地域的详情见 https://cloud.tencent.com/document/product/436/6224 。

	tc, err := NewTCStorage()
	if err != nil {
		panic(err)
	}

	_, err2 := tc.bucketTC.Object.Delete(context.Background(), tc.Prefix+filename)
	if err2 != nil {
		panic(err2)
	}
}
