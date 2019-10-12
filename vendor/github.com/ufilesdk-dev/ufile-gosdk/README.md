# UCloud 对象存储 SDK <a href="https://godoc.org/github.com/ufilesdk-dev/ufile-gosdk"><img src="https://godoc.org/github.com/ufilesdk-dev/ufile-gosdk?status.svg" alt="GoDoc"></a>
> Modules are interface and implementation.    
> The best modules are where interface is much simpler than implementation.  
> **By: John Ousterhout**

## UFile 对象存储基本概念
在对象存储系统中，存储空间（Bucket）是文件（File）的组织管理单位，文件（File）是存储空间的逻辑存储单元。对于每个账号，该账号里存放的每个文件都有唯一的一对存储空间（Bucket）与键（Key）作为标识。我们可以把 Bucket 理解成一类文件的集合，Key 理解成文件名。由于每个 Bucket 需要配置和权限不同，每个账户里面会有多个 Bucket。在 UFile 里面，Bucket 主要分为公有和私有两种，公有 Bucket 里面的文件可以对任何人开放，私有 Bucket 需要配置对应访问签名才能访问。

### 签名
本 SDK 接口是基于 HTTP 的，为了连接的安全性，UFile 使用 HMAC SHA1 对每个连接进行签名校验。使用本 SDK 可以忽略签名相关的算法过程，只要把公私钥写入到配置文件里面（注意不要传到版本控制里面），读取并传给 UFileRequest 里面的 New 方法即可。  
签名相关的算法与详细实现请见 [Auth 模块](auth.go)

## 安装
`go get github.com/ufilesdk-dev/ufile-gosdk`

### 执行测试
`cd example; go run demo_file.go`

## 功能列表
### 文件操作相关功能
[Put 上传](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.PutFile)
[Post 上传](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.PostFile)  
分片上传 [同步分片上传](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.MPut)，[异步分片上传](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.AsyncMPut)  
手动分片上传，[步骤一](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.InitiateMultipartUpload)，[步骤二](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.UploadPart)，[步骤三](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.FinishMultipartUpload)。[取消分片上传](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.AbortMultipartUpload)  
[文件秒传](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.UploadHit)  
[获取文件列表](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.PrefixFileList)  
[获取私有空间下载地址](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.GetPrivateURL)，[获取公有空间下载地址](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.GetPublicURL)。  
[删除文件](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.DeleteFile)  
[查看文件信息](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.HeadFile)  
[下载文件](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.DownloadFile)  
[比对本地与远程文件的 Etag](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.CompareFileEtag)  
[Put 带回调上传](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.PutFileWithPolicy)
[同步分片上传-带回调](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.MPutWithPolicy)，
[异步分片上传-带回调](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.AsyncMPutWithPolicy)  

### Bucket 操作相关功能
[创建 bucket](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.CreateBucket)  
[删除 bucket](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.DeleteBucket)  
[获取 bucket 列表](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.DescribeBucket)  
[修改 bucket](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#UFileRequest.UpdateBucket)  

### 签名构造
[构造文件管理签名](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#Auth.Authorization)  
[构造私有空间下载签名](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#Auth.AuthorizationPrivateURL)  
[构造 bucket 管理签名](https://godoc.org/pkg/github.com/ufilesdk-dev/ufile-gosdk/#Auth.AuthorizationBucketMgr)  

## 示例代码
SDK 主要分为两个模块，一个是 bucket 管理，一个是 file 管理。使用对象存储你需要频繁的调用 file 管理相关的接口，bucket 管理用到的地方不会太频繁。以下是用 SDK 上传一个文件的例子：
```go
import ufsdk "github.com/ufilesdk-dev/ufile-gosdk"
config, err := ufsdk.LoadConfig(configFile)
if err != nil {
    panic(err.Error())
}
req := ufsdk.NewFileRequest(config, nil)
err = req.PutFile(filePath, keyName, "")
if err != nil {
    fmt.Println("文件上传失败!!，错误信息为：", err.Error())
    //把 HTTP 详细的 HTTP response dump 出来
    fmt.Printf("%s\n",req.DumpResponse(true))
}
```
更详细的代码请参考 [example/demo_file.go](/example/demo_file.go) 和 [example/demo_bucket.go](example/demo_bucket.go)

## 文档说明
本 SDK 使用 [godoc](https://blog.golang.org/godoc-documenting-go-code) 约定的方法对每个 export 出来的接口进行注释。
你可以直接访问生成好的[在线文档](https://godoc.org/github.com/ufilesdk-dev/ufile-gosdk)。  

## 如何排错？
使用 UFileRequest 里面的方法对返回的 error 进行检查。如果不为 nil，调用 Error() 查看错误信息。调用 DumpResponse(true) 并获取返回值查看详细的 HTTP 返回值。

