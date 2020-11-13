package ufsdk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

//UFileRequest SDK 主要的 request 模块。本 SDK 遵从以下原则：
//
//1.接口尽可能简洁，隐藏所有复杂实现。
//
//2.本 SDK 主要的作用是封装 HTTP 请求，不做过多其他的封装（如 HTTP Body 序列化，详细的错误检查）。
//
//3.只简单封装 HTTP 请求所需要的参数，给接口使用者提供所有原生的 HTTP response header,body,status code 返回，以便排错。
//
//4.远端请求返回值统一返回一个 error，如果为 nil 表示无错。LastResponseStatus，LastResponseHeader，LastResponseBody 可以查看具体的 HTTP 返回信息（）。如果你想少敲几行代码可以直接调用 DumpResponse(true) 查看详细返回。
//
type UFileRequest struct {
	Auth               Auth
	BucketName         string
	Host               string
	Client             *http.Client
	Context            context.Context
	baseURL            *url.URL
	RequestHeader      http.Header

	LastResponseStatus int
	LastResponseHeader http.Header
	LastResponseBody   []byte
	verifyUploadMD5    bool
	lastResponse       *http.Response
}

//NewFileRequest 创建一个用于管理文件的 request，管理文件的 url 与 管理 bucket 接口不一样，
//请将 bucket 和文件管理所需要的分开，NewUBucketRequest 是用来管理 bucket 的。
//Request 创建后的 instance 不是线程安全的，如果你需要做并发的操作，请创建多个 UFileRequest。
//config 参数里面包含了公私钥，以及其他必填的参数。详情见 config 相关文档。
//client 这里你可以传空，会使用默认的 http.Client。如果你需要设置超时以及一些其他相关的网络配置选项请传入一个自定义的 client。
func NewFileRequest(config *Config, client *http.Client) (*UFileRequest, error) {
	config.BucketName = strings.TrimSpace(config.BucketName)
	config.FileHost = strings.TrimSpace(config.FileHost)
	if config.BucketName == "" || config.FileHost == "" {
		return nil, errors.New("管理文件上传必须要提供 bucket 名字和所在地域的 Host 域名")
	}
	req := newRequest(config.PublicKey, config.PrivateKey,
		config.BucketName, config.FileHost, client)
	req.verifyUploadMD5 = config.VerifyUploadMD5
	if req.baseURL.Scheme == "" { //用户传了非自定义域名
		req.baseURL.Host = req.BucketName + "." + req.Host
		req.baseURL.Scheme = "http"
	}
	return req, nil
}

//NewFileRequestWithHeader 创建一个可携带用户自定义http请求头的 request
//Request 创建后的 instance 不是线程安全的，如果你需要做并发的操作，请创建多个 UFileRequest。
//config 参数里面包含了公私钥，以及其他必填的参数。详情见 config 相关文档。
//header 自定义http请求头
//client 这里你可以传空，会使用默认的 http.Client。如果你需要设置超时以及一些其他相关的网络配置选项请传入一个自定义的 client。
func NewFileRequestWithHeader(config *Config, header http.Header, client *http.Client) (*UFileRequest, error) {
	req, err := NewFileRequest(config, client)
	if err != nil{
		return nil, err
	}
	req.RequestHeader = header
	return req, nil
}

//NewBucketRequest 创建一个用于管理 bucket 的 request。
//注意：不要拿它去调用文件管理的 request，我文件管理和 bucket 管理接口放到一个 request 里面的目的就是让接口更统一，代码更清晰，简洁。
//config 参数里面包含了公私钥，以及其他必填的参数。详情见 config 相关文档。
func NewBucketRequest(config *Config, client *http.Client) (*UFileRequest, error) {
	config.BucketHost = strings.TrimSpace(config.BucketHost)
	if config.BucketHost == "" {
		return nil, errors.New("管理 Bucket 必须要提供对应的 API host")
	}
	req := newRequest(config.PublicKey, config.PrivateKey, "", config.BucketHost, client)
	req.verifyUploadMD5 = config.VerifyUploadMD5
	if req.baseURL.Scheme == "" {
		req.baseURL.Scheme = "http"
	}
	return req, nil
}

//DumpResponse dump 当前请求的返回结果，里面有一个 print 函数，会把 body,header,status code 直接输出到 stdout。
//如果你需要 Dump 到其他的地方，直接拿返回值即可。
func (u *UFileRequest) DumpResponse(isDumpBody bool) []byte {
	var b bytes.Buffer
	if u.lastResponse == nil {
		return nil
	}
	b.WriteString(fmt.Sprintf("%s %d\n", u.lastResponse.Proto, u.LastResponseStatus))
	for k, vs := range u.LastResponseHeader {
		str := k + ": "
		for i, v := range vs {
			if i != 0 {
				str += "; " + v
			} else {
				str += v
			}
		}
		b.WriteString(str)
	}
	if isDumpBody {
		b.Write(u.LastResponseBody)
	}
	return b.Bytes()
}

func newRequest(publicKey, privateKey, bucket, host string, client *http.Client) *UFileRequest {
	req := new(UFileRequest)
	req.Auth = NewAuth(publicKey, privateKey)
	req.BucketName = bucket
	req.Host = strings.TrimSpace(host)
	req.baseURL = new(url.URL)
	req.baseURL.Host = req.Host
	req.baseURL.Path = "/" //for default usage.

	if client == nil {
		client = new(http.Client)
	}
	req.Client = client
	req.Context = context.TODO()
	return req
}

func (u *UFileRequest) responseParse(resp *http.Response) error {
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	u.LastResponseStatus = resp.StatusCode
	u.LastResponseHeader = resp.Header
	u.LastResponseBody = resBody
	u.lastResponse = resp
	return nil
}

func (u *UFileRequest) request(req *http.Request) error {
	resp, err := u.requestWithResp(req)
	if err != nil {
		return err
	}

	err = u.responseParse(resp)
	if err != nil {
		return err
	}

	if !VerifyHTTPCode(resp.StatusCode) {
		return fmt.Errorf("Remote response code is %d - %s not 2xx call DumpResponse(true) show details",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func (u *UFileRequest) requestWithResp(req *http.Request) (resp *http.Response, err error) {
	req.Header.Set("User-Agent", "UFile-GoSDK-Client/2.0")

	resp, err = u.Client.Do(req.WithContext(u.Context))
	// If we got an error, and the context has been canceled,
	// the context's error is probably more useful.
	if err != nil {
		select {
		case <-u.Context.Done():
			err = u.Context.Err()
		default:
		}
		return
	}
	return
}
