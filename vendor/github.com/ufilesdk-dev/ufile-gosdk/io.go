package ufsdk

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

// IOPut 流式 put 上传接口，你必须确保你的 reader 接口每次调用是递进式的调用，也就是像文件那样的读取方式。
// mimeType 在这里的检测不会很准确，你可以手动指定更精确的 mimetype。
// 这里的 reader 接口会把数据全部读到 HTTP Body 里面，如果你接口的数据特别大，请使用 IOMutipartAsyncUpload 接口。
func (u *UFileRequest) IOPut(reader io.Reader, keyName, mimeType string) (err error) {
	if keyName == "" {
		err = errors.New("keyName cannot be empty")
		return
	}

	switch reader.(type) {
	case *bytes.Buffer, *bytes.Reader, *strings.Reader:
		break
	default:
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	reqURL := u.genFileURL(keyName)
	req, err := http.NewRequest("PUT", reqURL, reader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", mimeType)

	authorization := u.Auth.Authorization("PUT", u.BucketName, keyName, req.Header)
	req.Header.Add("authorization", authorization)

	return u.request(req)
}

//
// IOMutipartAsyncUpload 流式分片上传接口，你必须确保你的 reader 接口每次调用是递进式的调用，也就是像文件那样的读取方式。
// mimeType 在这里的检测不会很准确，你可以手动指定更精确的 mimetype。
// 这里的会每次读取4M 的数据到 buffer 里面，适用于大量数据上传。
func (u *UFileRequest) IOMutipartAsyncUpload(reader io.Reader, keyName, mimeType string) (err error) {
	if keyName == "" {
		err = errors.New("keyName cannot be empty")
		return
	}
	state, err := u.InitiateMultipartUpload(keyName, mimeType)
	if err != nil {
		return
	}

	maxJobRunning := 10 //最多允许 10 个线程同时跑
	concurrentChan := make(chan error, maxJobRunning)
	for i := 0; i != maxJobRunning; i++ {
		concurrentChan <- nil
	}
	wg := &sync.WaitGroup{}
	for i := 0; ; i++ {
		uploadErr := <-concurrentChan //最初允许启动 10 个 goroutine，超出10个后，有分片返回才会开新的goroutine.
		if uploadErr != nil {
			u.AbortMultipartUpload(state)
			return uploadErr // 中间如果出现错误立即停止继续上传
		}

		chunk := make([]byte, state.BlkSize)
		bytesRead, readErr := reader.Read(chunk)
		if readErr == io.EOF || bytesRead == 0 {
			break
		}
		if readErr != nil {
			u.AbortMultipartUpload(state)
			return uploadErr // 检查读文件是否出现错误。
		}
		wg.Add(1)
		go func(pos int, buf *bytes.Buffer) {
			defer wg.Done()
			e := u.UploadPart(buf, state, pos)
			concurrentChan <- e //跑完一个 goroutine 后，发信号表示可以开启新的 goroutine。
		}(i, bytes.NewBuffer(chunk[:bytesRead]))
	}

	go func() {
		wg.Wait()
		close(concurrentChan) //close channel, when all upload goroutines has finished.
	}()

	for err = range concurrentChan { //waitting for all goroutine finished. It will blocked until the channel has been closed.
		if err != nil {
			u.AbortMultipartUpload(state)
			return err
		}
	}

	return u.FinishMultipartUpload(state)
}
