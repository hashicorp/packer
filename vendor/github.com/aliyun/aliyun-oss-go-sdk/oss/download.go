package oss

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

//
// DownloadFile 分片下载文件
//
// objectKey  object key。
// filePath   本地文件。objectKey下载到文件。
// partSize   本次上传文件片的大小，字节数。比如100 * 1024为每片100KB。
// options    Object的属性限制项。详见GetObject。
//
// error 操作成功error为nil，非nil为错误信息。
//
func (bucket Bucket) DownloadFile(objectKey, filePath string, partSize int64, options ...Option) error {
	if partSize < 1 || partSize > MaxPartSize {
		return errors.New("oss: part size invalid range (1, 5GB]")
	}

	cpConf, err := getCpConfig(options, filePath)
	if err != nil {
		return err
	}

	routines := getRoutines(options)

	if cpConf.IsEnable {
		return bucket.downloadFileWithCp(objectKey, filePath, partSize, options, cpConf.FilePath, routines)
	}

	return bucket.downloadFile(objectKey, filePath, partSize, options, routines)
}

// ----- 并发无断点的下载  -----

// 工作协程参数
type downloadWorkerArg struct {
	bucket   *Bucket
	key      string
	filePath string
	options  []Option
	hook     downloadPartHook
}

// Hook用于测试
type downloadPartHook func(part downloadPart) error

var downloadPartHooker downloadPartHook = defaultDownloadPartHook

func defaultDownloadPartHook(part downloadPart) error {
	return nil
}

// 默认ProgressListener，屏蔽GetObject的Options中ProgressListener
type defaultDownloadProgressListener struct {
}

// ProgressChanged 静默处理
func (listener *defaultDownloadProgressListener) ProgressChanged(event *ProgressEvent) {
}

// 工作协程
func downloadWorker(id int, arg downloadWorkerArg, jobs <-chan downloadPart, results chan<- downloadPart, failed chan<- error, die <-chan bool) {
	for part := range jobs {
		if err := arg.hook(part); err != nil {
			failed <- err
			break
		}

		// resolve options
		r := Range(part.Start, part.End)
		p := Progress(&defaultDownloadProgressListener{})
		opts := make([]Option, len(arg.options)+2)
		// append orderly, can not be reversed!
		opts = append(opts, arg.options...)
		opts = append(opts, r, p)

		rd, err := arg.bucket.GetObject(arg.key, opts...)
		if err != nil {
			failed <- err
			break
		}
		defer rd.Close()

		select {
		case <-die:
			return
		default:
		}

		fd, err := os.OpenFile(arg.filePath, os.O_WRONLY, FilePermMode)
		if err != nil {
			failed <- err
			break
		}
		defer fd.Close()

		_, err = fd.Seek(part.Start, os.SEEK_SET)
		if err != nil {
			failed <- err
			break
		}

		_, err = io.Copy(fd, rd)
		if err != nil {
			failed <- err
			break
		}

		results <- part
	}
}

// 调度协程
func downloadScheduler(jobs chan downloadPart, parts []downloadPart) {
	for _, part := range parts {
		jobs <- part
	}
	close(jobs)
}

// 下载片
type downloadPart struct {
	Index int   // 片序号，从0开始编号
	Start int64 // 片起始位置
	End   int64 // 片结束位置
}

// 文件分片
func getDownloadParts(bucket *Bucket, objectKey string, partSize int64) ([]downloadPart, error) {
	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		return nil, err
	}

	parts := []downloadPart{}
	objectSize, err := strconv.ParseInt(meta.Get(HTTPHeaderContentLength), 10, 0)
	if err != nil {
		return nil, err
	}

	part := downloadPart{}
	i := 0
	for offset := int64(0); offset < objectSize; offset += partSize {
		part.Index = i
		part.Start = offset
		part.End = GetPartEnd(offset, objectSize, partSize)
		parts = append(parts, part)
		i++
	}
	return parts, nil
}

// 文件大小
func getObjectBytes(parts []downloadPart) int64 {
	var ob int64
	for _, part := range parts {
		ob += (part.End - part.Start + 1)
	}
	return ob
}

// 并发无断点续传的下载
func (bucket Bucket) downloadFile(objectKey, filePath string, partSize int64, options []Option, routines int) error {
	tempFilePath := filePath + TempFileSuffix
	listener := getProgressListener(options)

	// 如果文件不存在则创建，存在不清空，下载分片会重写文件内容
	fd, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE, FilePermMode)
	if err != nil {
		return err
	}
	fd.Close()

	// 分割文件
	parts, err := getDownloadParts(&bucket, objectKey, partSize)
	if err != nil {
		return err
	}

	jobs := make(chan downloadPart, len(parts))
	results := make(chan downloadPart, len(parts))
	failed := make(chan error)
	die := make(chan bool)

	var completedBytes int64
	totalBytes := getObjectBytes(parts)
	event := newProgressEvent(TransferStartedEvent, 0, totalBytes)
	publishProgress(listener, event)

	// 启动工作协程
	arg := downloadWorkerArg{&bucket, objectKey, tempFilePath, options, downloadPartHooker}
	for w := 1; w <= routines; w++ {
		go downloadWorker(w, arg, jobs, results, failed, die)
	}

	// 并发上传分片
	go downloadScheduler(jobs, parts)

	// 等待分片下载完成
	completed := 0
	ps := make([]downloadPart, len(parts))
	for completed < len(parts) {
		select {
		case part := <-results:
			completed++
			ps[part.Index] = part
			completedBytes += (part.End - part.Start + 1)
			event = newProgressEvent(TransferDataEvent, completedBytes, totalBytes)
			publishProgress(listener, event)
		case err := <-failed:
			close(die)
			event = newProgressEvent(TransferFailedEvent, completedBytes, totalBytes)
			publishProgress(listener, event)
			return err
		}

		if completed >= len(parts) {
			break
		}
	}

	event = newProgressEvent(TransferCompletedEvent, completedBytes, totalBytes)
	publishProgress(listener, event)

	return os.Rename(tempFilePath, filePath)
}

// ----- 并发有断点的下载  -----

const downloadCpMagic = "92611BED-89E2-46B6-89E5-72F273D4B0A3"

type downloadCheckpoint struct {
	Magic    string         // magic
	MD5      string         // cp内容的MD5
	FilePath string         // 本地文件
	Object   string         // key
	ObjStat  objectStat     // 文件状态
	Parts    []downloadPart // 全部分片
	PartStat []bool         // 分片下载是否完成
}

type objectStat struct {
	Size         int64  // 大小
	LastModified string // 最后修改时间
	Etag         string // etag
}

// CP数据是否有效，CP有效且Object没有更新时有效
func (cp downloadCheckpoint) isValid(bucket *Bucket, objectKey string) (bool, error) {
	// 比较CP的Magic及MD5
	cpb := cp
	cpb.MD5 = ""
	js, _ := json.Marshal(cpb)
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])

	if cp.Magic != downloadCpMagic || b64 != cp.MD5 {
		return false, nil
	}

	// 确认object没有更新
	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		return false, err
	}

	objectSize, err := strconv.ParseInt(meta.Get(HTTPHeaderContentLength), 10, 0)
	if err != nil {
		return false, err
	}

	// 比较Object的大小/最后修改时间/etag
	if cp.ObjStat.Size != objectSize ||
		cp.ObjStat.LastModified != meta.Get(HTTPHeaderLastModified) ||
		cp.ObjStat.Etag != meta.Get(HTTPHeaderEtag) {
		return false, nil
	}

	return true, nil
}

// 从文件中load
func (cp *downloadCheckpoint) load(filePath string) error {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, cp)
	return err
}

// dump到文件
func (cp *downloadCheckpoint) dump(filePath string) error {
	bcp := *cp

	// 计算MD5
	bcp.MD5 = ""
	js, err := json.Marshal(bcp)
	if err != nil {
		return err
	}
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	bcp.MD5 = b64

	// 序列化
	js, err = json.Marshal(bcp)
	if err != nil {
		return err
	}

	// dump
	return ioutil.WriteFile(filePath, js, FilePermMode)
}

// 未完成的分片
func (cp downloadCheckpoint) todoParts() []downloadPart {
	dps := []downloadPart{}
	for i, ps := range cp.PartStat {
		if !ps {
			dps = append(dps, cp.Parts[i])
		}
	}
	return dps
}

// 完成的字节数
func (cp downloadCheckpoint) getCompletedBytes() int64 {
	var completedBytes int64
	for i, part := range cp.Parts {
		if cp.PartStat[i] {
			completedBytes += (part.End - part.Start + 1)
		}
	}
	return completedBytes
}

// 初始化下载任务
func (cp *downloadCheckpoint) prepare(bucket *Bucket, objectKey, filePath string, partSize int64) error {
	// cp
	cp.Magic = downloadCpMagic
	cp.FilePath = filePath
	cp.Object = objectKey

	// object
	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		return err
	}

	objectSize, err := strconv.ParseInt(meta.Get(HTTPHeaderContentLength), 10, 0)
	if err != nil {
		return err
	}

	cp.ObjStat.Size = objectSize
	cp.ObjStat.LastModified = meta.Get(HTTPHeaderLastModified)
	cp.ObjStat.Etag = meta.Get(HTTPHeaderEtag)

	// parts
	cp.Parts, err = getDownloadParts(bucket, objectKey, partSize)
	if err != nil {
		return err
	}
	cp.PartStat = make([]bool, len(cp.Parts))
	for i := range cp.PartStat {
		cp.PartStat[i] = false
	}

	return nil
}

func (cp *downloadCheckpoint) complete(cpFilePath, downFilepath string) error {
	os.Remove(cpFilePath)
	return os.Rename(downFilepath, cp.FilePath)
}

// 并发带断点的下载
func (bucket Bucket) downloadFileWithCp(objectKey, filePath string, partSize int64, options []Option, cpFilePath string, routines int) error {
	tempFilePath := filePath + TempFileSuffix
	listener := getProgressListener(options)

	// LOAD CP数据
	dcp := downloadCheckpoint{}
	err := dcp.load(cpFilePath)
	if err != nil {
		os.Remove(cpFilePath)
	}

	// LOAD出错或数据无效重新初始化下载
	valid, err := dcp.isValid(&bucket, objectKey)
	if err != nil || !valid {
		if err = dcp.prepare(&bucket, objectKey, filePath, partSize); err != nil {
			return err
		}
		os.Remove(cpFilePath)
	}

	// 如果文件不存在则创建，存在不清空，下载分片会重写文件内容
	fd, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE, FilePermMode)
	if err != nil {
		return err
	}
	fd.Close()

	// 未完成的分片
	parts := dcp.todoParts()
	jobs := make(chan downloadPart, len(parts))
	results := make(chan downloadPart, len(parts))
	failed := make(chan error)
	die := make(chan bool)

	completedBytes := dcp.getCompletedBytes()
	event := newProgressEvent(TransferStartedEvent, completedBytes, dcp.ObjStat.Size)
	publishProgress(listener, event)

	// 启动工作协程
	arg := downloadWorkerArg{&bucket, objectKey, tempFilePath, options, downloadPartHooker}
	for w := 1; w <= routines; w++ {
		go downloadWorker(w, arg, jobs, results, failed, die)
	}

	// 并发下载分片
	go downloadScheduler(jobs, parts)

	// 等待分片下载完成
	completed := 0
	for completed < len(parts) {
		select {
		case part := <-results:
			completed++
			dcp.PartStat[part.Index] = true
			dcp.dump(cpFilePath)
			completedBytes += (part.End - part.Start + 1)
			event = newProgressEvent(TransferDataEvent, completedBytes, dcp.ObjStat.Size)
			publishProgress(listener, event)
		case err := <-failed:
			close(die)
			event = newProgressEvent(TransferFailedEvent, completedBytes, dcp.ObjStat.Size)
			publishProgress(listener, event)
			return err
		}

		if completed >= len(parts) {
			break
		}
	}

	event = newProgressEvent(TransferCompletedEvent, completedBytes, dcp.ObjStat.Size)
	publishProgress(listener, event)

	return dcp.complete(cpFilePath, tempFilePath)
}
