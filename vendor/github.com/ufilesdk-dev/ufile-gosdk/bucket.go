package ufsdk

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type response interface {
	Error() error
}

//BucketResponse 用于 Bucket 模块返回的数据。
type BucketResponse struct {
	RetCode    int    `json:"RetCode,omitempty"`
	Action     string `json:"Action,omitempty"`
	BucketName string `json:"BucketName,omitempty"`
	BucketID   string `json:"BucketId,omitempty"`
	Message    string `json:"Message,omitempty"`
}

func (b BucketResponse) Error() error {
	if b.RetCode != 0 {
		return errors.New(b.Message)
	}
	return nil
}

func (b BucketResponse) String() string {
	return structPrettyStr(b)
}

//DomainSet 用于 BucketDataSet 里面的 Domain 字段
type DomainSet struct {
	Src       []string `json:"Src,omitempty"`
	Cdn       []string `json:"Cdn,omitempty"`
	CustomSrc []string `json:"CustomSrc,omitempty"`
	CustomCdn []string `json:"CustomCdn,omitempty"`
}

//BucketDataSet 用于 BucketListResponse 里面的 DataSet 字段
type BucketDataSet struct {
	BucketName    string    `json:"BucketName,omitempty"`
	BucketID      string    `json:"BucketId,omitempty"`
	Domain        DomainSet `json:"Domain,omitempty"`
	Type          string    `json:"Type,omitempty"`
	CreateTime    int       `json:"CreateTime,omitempty"`
	ModifyTime    int       `json:"ModifyTime,omitempty"`
	CdnDomainID   []string  `json:"CdnDomainId,omitempty"`
	Biz           string    `json:"Biz,omitempty"`
	Region        string    `json:"Region,omitempty"`
	HasUserDomain int       `json:"HasUserDomain,omitempty"`
}

//BucketListResponse 用于 DescribeBucket 返回的数据。
type BucketListResponse struct {
	RetCode int             `json:"RetCode,omitempty"`
	Action  string          `json:"Action,omitempty"`
	Message string          `json:"Message,omitempty"`
	DataSet []BucketDataSet `json:"DataSet,omitempty"`
}

func (b BucketListResponse) Error() error {
	if b.RetCode != 0 {
		return errors.New(b.Message)
	}
	return nil
}

//String 把 BucketListResponse 里面的字段格式化。
func (b BucketListResponse) String() string {
	return structPrettyStr(b)
}

//CreateBucket 创建一个 bucket, bucketName 必须全部为小写字母，不能带符号和特殊字符。
//
//region 表示 ufile 所在的可用区，目前支持北京，香港，广州，上海二，雅加达，洛杉矶。一下是可用区值的映射：
//
//北京 cn-bj
//
//广州 cn-gd
//
//可用区以控制台列出来的为准，更多可用区具体的值在 https://docs.ucloud.cn/api/summary/regionlist 查看。
//
//bucketType 可以填 public（公有空间） 和 private（私有空间）
//projectID bucket 所在的项目 ID，可为空。
func (u *UFileRequest) CreateBucket(bucketName, region, bucketType, projectID string) (bucket BucketResponse, err error) {
	query := url.Values{}
	query.Add("Action", "CreateBucket")
	query.Add("BucketName", bucketName)
	query.Add("Type", bucketType)
	query.Add("Region", region)

	if projectID != "" {
		query.Add("ProjectId", projectID)
	}

	err = u.bucketRequest(query, &bucket)
	return
}

//DeleteBucket 删除一个 bucket，如果成功，status code 会返回 204 no-content
func (u *UFileRequest) DeleteBucket(bucketName, projectID string) (bucket BucketResponse, err error) {
	query := url.Values{}
	query.Add("Action", "DeleteBucket")
	query.Add("BucketName", bucketName)
	if projectID != "" {
		query.Add("ProjectId", projectID)
	}

	err = u.bucketRequest(query, &bucket)
	return
}

//UpdateBucket 更新一个 bucket，你可以改 bucket 的类型（私有或公有）和 项目 ID。
//bucketType 填公有（public）或私有（private）。
//projectID 没有可以填空（""）。
func (u *UFileRequest) UpdateBucket(bucketName, bucketType, projectID string) (bucket BucketResponse, err error) {
	query := url.Values{}
	query.Add("Action", "UpdateBucket")
	query.Add("BucketName", bucketName)
	query.Add("Type", bucketType)
	if projectID != "" {
		query.Add("ProjectId", projectID)
	}

	err = u.bucketRequest(query, &bucket)
	return
}

//DescribeBucket 获取 bucket 的详细信息，如果 bucketName 为空，返回当前账号下所有的 bucket。
//limit 是限制返回的 bucket 列表数量。
//offset 是列表的偏移量，默认为 0。
//projectID 可为空。
func (u *UFileRequest) DescribeBucket(bucketName string, offset, limit int, projectID string) (list BucketListResponse, err error) {
	query := url.Values{}
	query.Add("Action", "DescribeBucket")
	if bucketName != "" {
		query.Add("BucketName", bucketName)
	}
	//offset default is 0
	query.Add("Offset", strconv.Itoa(offset))

	if limit == 0 {
		limit = 20
	}
	query.Add("Limit", strconv.Itoa(limit))

	if projectID != "" {
		query.Add("ProjectId", projectID)
	}
	err = u.bucketRequest(query, &list)
	return
}

func (u *UFileRequest) bucketRequest(query url.Values, data response) error {
	reqURL := u.genBucketURL(query)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := u.Client.Do(req)
	if err != nil {
		return err
	}
	err = u.responseParse(resp)
	if err != nil {
		return err
	}
	err = json.Unmarshal(u.LastResponseBody, data)
	if err != nil {
		return err
	}
	return data.Error()
}

func (u *UFileRequest) genBucketURL(query url.Values) string {
	u.baseURL.RawQuery = u.Auth.AuthorizationBucketMgr(query)
	return u.baseURL.String()
}
