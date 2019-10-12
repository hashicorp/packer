package ufsdk

import (
	"bytes"
	"encoding/base64"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	fourMegabyte = 1 << 22 //4M
)

//FileDataSet  用于 FileListResponse 里面的 DataSet 字段。
type FileDataSet struct {
	BucketName    string `json:"BucketName,omitempty"`
	FileName      string `json:"FileName,omitempty"`
	Hash          string `json:"Hash,omitempty"`
	MimeType      string `json:"MimeType,omitempty"`
	FirstObject   string `json:"first_object,omitempty"`
	Size          int    `json:"Size,omitempty"`
	CreateTime    int    `json:"CreateTime,omitempty"`
	ModifyTime    int    `json:"ModifyTime,omitempty"`
	StorageClass  string `json:"StorageClass,omitempty"`
	RestoreStatus string `json:"RestoreStatus,omitempty"`
}

//FileListResponse 用 PrefixFileList 接口返回的 list 数据。
type FileListResponse struct {
	BucketName string        `json:"BucketName,omitempty"`
	BucketID   string        `json:"BucketId,omitempty"`
	NextMarker string        `json:"NextMarker,omitempty"`
	DataSet    []FileDataSet `json:"DataSet,omitempty"`
}

func (f FileListResponse) String() string {
	return structPrettyStr(f)
}

//UploadHit 文件秒传，它的原理是计算出文件的 etag 值与远端服务器进行对比，如果文件存在就快速返回。
func (u *UFileRequest) UploadHit(filePath, keyName string) (err error) {
	file, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	fsize := getFileSize(file)
	etag := calculateEtag(file)

	query := &url.Values{}
	query.Add("Hash", etag)
	query.Add("FileName", keyName)
	query.Add("FileSize", strconv.FormatInt(fsize, 10))
	reqURL := u.genFileURL("uploadhit") + "?" + query.Encode()
	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("POST", u.BucketName, keyName, req.Header)
	req.Header.Add("authorization", authorization)

	return u.request(req)
}

//PostFile 使用 HTTP Form 的方式上传一个文件。
//注意：使用本接口上传文件后，调用 UploadHit 接口会返回 404，因为经过 form 包装的文件，etag 值会不一样，所以会调用失败。
//mimeType 如果为空的话，会调用 net/http 里面的 DetectContentType 进行检测。
//keyName 表示传到 ufile 的文件名。
//小于 100M 的文件推荐使用本接口上传。
func (u *UFileRequest) PostFile(filePath, keyName, mimeType string) (err error) {
	file, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	h := make(http.Header)
	for k, v := range u.RequestHeader {
		for i := 0; i < len(v); i++ {
			h.Add(k, v[i])
		}
	}
	if mimeType == "" {
		mimeType = getMimeType(file)
	}
	h.Add("Content-Type", mimeType)

	authorization := u.Auth.Authorization("POST", u.BucketName, keyName, h)

	boundry := makeBoundry()
	body := makeFormBody(authorization, boundry, keyName, mimeType, u.verifyUploadMD5, file)
	//lastline 一定要写，否则后端解析不到。
	lastline := fmt.Sprintf("\r\n--%s--\r\n", boundry)
	body.Write([]byte(lastline))

	reqURL := u.genFileURL("")
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundry)
	contentLength := body.Len()
	req.Header.Add("Content-Length", strconv.Itoa(contentLength))
	for k, v := range u.RequestHeader {
		for i := 0; i < len(v); i++ {
			req.Header.Add(k, v[i])
		}
	}
	return u.request(req)
}

//PutFile 把文件直接放到 HTTP Body 里面上传，相对 PostFile 接口，这个要更简单，速度会更快（因为不用包装 form）。
//mimeType 如果为空的，会调用 net/http 里面的 DetectContentType 进行检测。
//keyName 表示传到 ufile 的文件名。
//小于 100M 的文件推荐使用本接口上传。
func (u *UFileRequest) PutFile(filePath, keyName, mimeType string) error {
	reqURL := u.genFileURL(keyName)
	file, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", reqURL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	if mimeType == "" {
		mimeType = getMimeType(file)
	}
	req.Header.Add("Content-Type", mimeType)
	for k, v := range u.RequestHeader {
		for i := 0; i < len(v); i++ {
			req.Header.Add(k, v[i])
		}
	}

	if u.verifyUploadMD5 {
		md5Str := fmt.Sprintf("%x", md5.Sum(b))
		req.Header.Add("Content-MD5", md5Str)
	}

	authorization := u.Auth.Authorization("PUT", u.BucketName, keyName, req.Header)
	req.Header.Add("authorization", authorization)
	fileSize := getFileSize(file)
	req.Header.Add("Content-Length", strconv.FormatInt(fileSize, 10))

	return u.request(req)
}

//PutFile 把文件直接放到 HTTP Body 里面上传，相对 PostFile 接口，这个要更简单，速度会更快（因为不用包装 form）。
//mimeType 如果为空的，会调用 net/http 里面的 DetectContentType 进行检测。
//keyName 表示传到 ufile 的文件名。
//小于 100M 的文件推荐使用本接口上传。
//支持带上传回调的参数, policy_json 为json 格式字符串
func (u *UFileRequest) PutFileWithPolicy(filePath, keyName, mimeType string, policy_json string) error {
	reqURL := u.genFileURL(keyName)
	file, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", reqURL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	if mimeType == "" {
		mimeType = getMimeType(file)
	}
	req.Header.Add("Content-Type", mimeType)

	if u.verifyUploadMD5 {
		md5Str := fmt.Sprintf("%x", md5.Sum(b))
		req.Header.Add("Content-MD5", md5Str)
	}

	policy := base64.URLEncoding.EncodeToString([]byte(policy_json))
	authorization := u.Auth.AuthorizationPolicy("PUT", u.BucketName, keyName, policy, req.Header)
	req.Header.Add("authorization", authorization)
	fileSize := getFileSize(file)
	req.Header.Add("Content-Length", strconv.FormatInt(fileSize, 10))

	return u.request(req)
}


//DeleteFile 删除一个文件，如果删除成功 statuscode 会返回 204，否则会返回 404 表示文件不存在。
//keyName 表示传到 ufile 的文件名。
func (u *UFileRequest) DeleteFile(keyName string) error {
	reqURL := u.genFileURL(keyName)
	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("DELETE", u.BucketName, keyName, req.Header)
	req.Header.Add("authorization", authorization)
	return u.request(req)
}

//HeadFile 获取一个文件的基本信息，返回的信息全在 header 里面。包含 mimeType, content-length（文件大小）, etag, Last-Modified:。
//keyName 表示传到 ufile 的文件名。
func (u *UFileRequest) HeadFile(keyName string) error {
	reqURL := u.genFileURL(keyName)
	req, err := http.NewRequest("HEAD", reqURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("HEAD", u.BucketName, keyName, req.Header)
	req.Header.Add("authorization", authorization)
	return u.request(req)
}

//PrefixFileList 获取文件列表。
//prefix 表示匹配文件前缀。
//marker 标志字符串
//limit 列表数量限制，传 0 会默认设置为 20.
func (u *UFileRequest) PrefixFileList(prefix, marker string, limit int) (list FileListResponse, err error) {
	query := &url.Values{}
	query.Add("prefix", prefix)
	query.Add("marker", marker)
	if limit == 0 {
		limit = 20
	}
	query.Add("limit", strconv.Itoa(limit))
	reqURL := u.genFileURL("") + "?list&" + query.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	authorization := u.Auth.Authorization("GET", u.BucketName, "", req.Header)
	req.Header.Add("authorization", authorization)

	err = u.request(req)
	if err != nil {
		return
	}
	err = json.Unmarshal(u.LastResponseBody, &list)
	return
}

//GetPublicURL 获取公有空间的文件下载 URL
//keyName 表示传到 ufile 的文件名。
func (u *UFileRequest) GetPublicURL(keyName string) string {
	return u.genFileURL(keyName)
}

//GetPrivateURL 获取私有空间的文件下载 URL。
//keyName 表示传到 ufile 的文件名。
//expiresDuation 表示下载链接的过期时间，从现在算起，24 * time.Hour 表示过期时间为一天。
func (u *UFileRequest) GetPrivateURL(keyName string, expiresDuation time.Duration) string {
	t := time.Now()
	t = t.Add(expiresDuation)
	expires := strconv.FormatInt(t.Unix(), 10)
	signature, publicKey := u.Auth.AuthorizationPrivateURL("GET", u.BucketName, keyName, expires, http.Header{})
	query := url.Values{}
	query.Add("UCloudPublicKey", publicKey)
	query.Add("Signature", signature)
	query.Add("Expires", expires)
	reqURL := u.genFileURL(keyName)
	return reqURL + "?" + query.Encode()
}

//Download 把文件下载到 HTTP Body 里面，这里只能用来下载小文件，建议使用 DownloadFile 来下载大文件。
func (u *UFileRequest) Download(reqURL string) error {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	return u.request(req)
}

//Download 文件下载接口，下载前会先获取文件大小，如果小于 4M 直接下载。大于 4M 每次会按 4M 的分片来下载。
func (u *UFileRequest) DownloadFile(writer io.Writer, keyName string) error {
	err := u.HeadFile(keyName)
	if err != nil {
		return err
	}
	size := u.LastResponseHeader.Get("Content-Length")
	fileSize, err := strconv.ParseInt(size, 10, 0)
	if err != nil || fileSize <= 0 {
		return fmt.Errorf("Parse content-lengt returned error")
	}

	reqURL := u.GetPrivateURL(keyName, 24*time.Hour)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	if fileSize < fourMegabyte {
		err = u.request(req)
		if err != nil {
			return err
		}
		writer.Write(u.LastResponseBody)
	} else {
		var i int64
		for i = 0; i < fileSize; i += fourMegabyte { // 一次下载 4 M
			start := i
			end := i + fourMegabyte - 1 //数组是从 0 开始的。 &_& .....
			if end > fileSize {
				end = fileSize
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
			err = u.request(req)
			if err != nil {
				return err
			}
			writer.Write(u.LastResponseBody)
		}
	}
	return nil
}

//CompareFileEtag 检查远程文件的 etag 和本地文件的 etag 是否一致
func (u *UFileRequest) CompareFileEtag(remoteKeyName, localFilePath string) bool {
	err := u.HeadFile(remoteKeyName)
	if err != nil {
		return false
	}
	remoteEtag := strings.Trim(u.LastResponseHeader.Get("Etag"), "\"")
	localEtag := GetFileEtag(localFilePath)
	return remoteEtag == localEtag
}

func (u *UFileRequest) genFileURL(keyName string) string {
	return u.baseURL.String() + keyName
}

//Restore 用于解冻冷存类型的文件
func (u *UFileRequest) Restore(keyName string) (err error) {
	reqURL := u.genFileURL(keyName) + "?restore"
	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("PUT", u.BucketName, keyName, req.Header)
	req.Header.Add("authorization", authorization)
	return u.request(req)
}

//ClassSwitch 存储类型转换接口
//keyName 文件名称
//storageClass 所要转换的新文件存储类型，目前支持的类型分别是标准:"STANDARD"、低频:"IA"、冷存:"ARCHIVE"
func (u *UFileRequest) ClassSwitch(keyName string, storageClass string) (err error) {
	query := &url.Values{}
	query.Add("storageClass", storageClass)
	reqURL := u.genFileURL(keyName) + "?" + query.Encode()
	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		return err
	}
	authorization := u.Auth.Authorization("PUT", u.BucketName, keyName, req.Header)
	req.Header.Add("authorization", authorization)
	return u.request(req)
}
