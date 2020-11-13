package ufsdk

import (
	"bytes"
	"io"
	"encoding/base64"	
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

//带回调策略的mput 接口，一些基础函数依赖于 file_mput 定义的函数

//MPut 分片上传一个文件，filePath 是本地文件所在的路径，内部会自动对文件进行分片上传，上传的方式是同步一片一片的上传。
//mimeType 如果为空的话，会调用 net/http 里面的 DetectContentType 进行检测。
//keyName 表示传到 ufile 的文件名。
//大于 100M 的文件推荐使用本接口上传。
func (u *UFileRequest) MPutWithPolicy(filePath, keyName, mimeType string, policy_json string) error {
	file, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if mimeType == "" {
		mimeType = getMimeType(file)
	}

	state, err := u.InitiateMultipartUpload(keyName, mimeType)
	if err != nil {
		return err
	}

	chunk := make([]byte, state.BlkSize)
	var pos int
	for {
		bytesRead, fileErr := file.Read(chunk)
		if fileErr == io.EOF || bytesRead == 0 { //后面直接读到了结尾
			break
		}
		buf := bytes.NewBuffer(chunk[:bytesRead])
		err := u.UploadPart(buf, state, pos)
		if err != nil {
			u.AbortMultipartUpload(state)
			return err
		}
		pos++
	}

	return u.FinishMultipartUploadWithPolicy(state, policy_json)
}

//AsyncMPut 异步分片上传一个文件，filePath 是本地文件所在的路径，内部会自动对文件进行分片上传，上传的方式是使用异步的方式同时传多个分片的块。
//mimeType 如果为空的话，会调用 net/http 里面的 DetectContentType 进行检测。
//keyName 表示传到 ufile 的文件名。
//大于 100M 的文件推荐使用本接口上传。
//同时并发上传的分片数量为10
func (u *UFileRequest) AsyncMPutWithPolicy(filePath, keyName, mimeType string, policy_json string) error {
	return u.AsyncUploadWithPolicy(filePath, keyName, mimeType, 10, policy_json)
}

//AsyncUpload AsyncMPut 的升级版, jobs 表示同时并发的数量。
func (u *UFileRequest) AsyncUploadWithPolicy(filePath, keyName, mimeType string, jobs int, policy_json string) error {
	if jobs <= 0 {
		jobs = 1
	}

	if jobs >= 30 {
		jobs = 10
	}

	file, err := openFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if mimeType == "" {
		mimeType = getMimeType(file)
	}

	state, err := u.InitiateMultipartUpload(keyName, mimeType)
	if err != nil {
		return err
	}
	fsize := getFileSize(file)
	chunkCount := divideCeil(fsize, int64(state.BlkSize)) //向上取整
	concurrentChan := make(chan error, jobs)
	for i := 0; i != jobs; i++ {
		concurrentChan <- nil
	}

	wg := &sync.WaitGroup{}
	for i := 0; i != chunkCount; i++ {
		uploadErr := <-concurrentChan //最初允许启动 10 个 goroutine，超出10个后，有分片返回才会开新的goroutine.
		if uploadErr != nil {
			err = uploadErr
			break // 中间如果出现错误立即停止继续上传
		}
		wg.Add(1)
		go func(pos int) {
			defer wg.Done()
			offset := int64(state.BlkSize * pos)
			chunk := make([]byte, state.BlkSize)
			bytesRead, _ := file.ReadAt(chunk, offset)
			e := u.UploadPart(bytes.NewBuffer(chunk[:bytesRead]), state, pos)
			concurrentChan <- e //跑完一个 goroutine 后，发信号表示可以开启新的 goroutine。
		}(i)
	}
	wg.Wait()       //等待所有任务返回
	if err == nil { //再次检查剩余上传完的分片是否有错误
	loopCheck:
		for {
			select {
			case e := <-concurrentChan:
				err = e
				if err != nil {
					break loopCheck
				}
			default:
				break loopCheck
			}
		}
	}
	close(concurrentChan)
	if err != nil {
		u.AbortMultipartUpload(state)
		return err
	}

	return u.FinishMultipartUploadWithPolicy(state, policy_json)
}


//FinishMultipartUpload 完成分片上传。分片上传必须要调用的接口。
//state 参数是 InitiateMultipartUpload 返回的
func (u *UFileRequest) FinishMultipartUploadWithPolicy(state *MultipartState, policy_json string) error {
	query := &url.Values{}
	query.Add("uploadId", state.uploadID)
	reqURL := u.genFileURL(state.keyName) + "?" + query.Encode()
	var etagsStr string
	etagLen := len(state.etags)
	for i := 0; i != etagLen; i++ {
		etagsStr += state.etags[i]
		if i != etagLen-1 {
			etagsStr += ","
		}
	}

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(etagsStr))
	if err != nil {
		return err
	}
	
	req.Header.Add("Content-Type", state.mimeType)

	policy := base64.URLEncoding.EncodeToString([]byte(policy_json))
	authorization := u.Auth.AuthorizationPolicy("POST", u.BucketName, state.keyName, policy, req.Header)

	req.Header.Add("Authorization", authorization)
	req.Header.Add("Content-Length", strconv.Itoa(len(etagsStr)))

	return u.request(req)
}


