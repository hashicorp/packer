package oss

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc64"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Get User Agent
// Go sdk相关信息，包括sdk版本，操作系统类型，GO版本
var userAgent = func() string {
	sys := getSysInfo()
	return fmt.Sprintf("aliyun-sdk-go/%s (%s/%s/%s;%s)", Version, sys.name,
		sys.release, sys.machine, runtime.Version())
}()

type sysInfo struct {
	name    string // 操作系统名称windows/Linux
	release string // 操作系统版本 2.6.32-220.23.2.ali1089.el5.x86_64等
	machine string // 机器类型amd64/x86_64
}

// Get　system info
// 获取操作系统信息、机器类型
func getSysInfo() sysInfo {
	name := runtime.GOOS
	release := "-"
	machine := runtime.GOARCH
	if out, err := exec.Command("uname", "-s").CombinedOutput(); err == nil {
		name = string(bytes.TrimSpace(out))
	}
	if out, err := exec.Command("uname", "-r").CombinedOutput(); err == nil {
		release = string(bytes.TrimSpace(out))
	}
	if out, err := exec.Command("uname", "-m").CombinedOutput(); err == nil {
		machine = string(bytes.TrimSpace(out))
	}
	return sysInfo{name: name, release: release, machine: machine}
}

// GetNowSec returns Unix time, the number of seconds elapsed since January 1, 1970 UTC.
// 获取当前时间，从UTC开始的秒数。
func GetNowSec() int64 {
	return time.Now().Unix()
}

// GetNowNanoSec returns t as a Unix time, the number of nanoseconds elapsed
// since January 1, 1970 UTC. The result is undefined if the Unix time
// in nanoseconds cannot be represented by an int64. Note that this
// means the result of calling UnixNano on the zero Time is undefined.
// 获取当前时间，从UTC开始的纳秒。
func GetNowNanoSec() int64 {
	return time.Now().UnixNano()
}

// GetNowGMT 获取当前时间，格式形如"Mon, 02 Jan 2006 15:04:05 GMT"，HTTP中使用的时间格式
func GetNowGMT() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

// FileChunk 文件片定义
type FileChunk struct {
	Number int   // 块序号
	Offset int64 // 块在文件中的偏移量
	Size   int64 // 块大小
}

// SplitFileByPartNum Split big file to part by the num of part
// 按指定的块数分割文件。返回值FileChunk为分割结果，error为nil时有效。
func SplitFileByPartNum(fileName string, chunkNum int) ([]FileChunk, error) {
	if chunkNum <= 0 || chunkNum > 10000 {
		return nil, errors.New("chunkNum invalid")
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if int64(chunkNum) > stat.Size() {
		return nil, errors.New("oss: chunkNum invalid")
	}

	var chunks []FileChunk
	var chunk = FileChunk{}
	var chunkN = (int64)(chunkNum)
	for i := int64(0); i < chunkN; i++ {
		chunk.Number = int(i + 1)
		chunk.Offset = i * (stat.Size() / chunkN)
		if i == chunkN-1 {
			chunk.Size = stat.Size()/chunkN + stat.Size()%chunkN
		} else {
			chunk.Size = stat.Size() / chunkN
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// SplitFileByPartSize Split big file to part by the size of part
// 按块大小分割文件。返回值FileChunk为分割结果，error为nil时有效。
func SplitFileByPartSize(fileName string, chunkSize int64) ([]FileChunk, error) {
	if chunkSize <= 0 {
		return nil, errors.New("chunkSize invalid")
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	var chunkN = stat.Size() / chunkSize
	if chunkN >= 10000 {
		return nil, errors.New("Too many parts, please increase part size.")
	}

	var chunks []FileChunk
	var chunk = FileChunk{}
	for i := int64(0); i < chunkN; i++ {
		chunk.Number = int(i + 1)
		chunk.Offset = i * chunkSize
		chunk.Size = chunkSize
		chunks = append(chunks, chunk)
	}

	if stat.Size()%chunkSize > 0 {
		chunk.Number = len(chunks) + 1
		chunk.Offset = int64(len(chunks)) * chunkSize
		chunk.Size = stat.Size() % chunkSize
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// GetPartEnd 计算结束位置
func GetPartEnd(begin int64, total int64, per int64) int64 {
	if begin+per > total {
		return total - 1
	}
	return begin + per - 1
}

// crcTable returns the Table constructed from the specified polynomial
var crcTable = func() *crc64.Table {
	return crc64.MakeTable(crc64.ECMA)
}
