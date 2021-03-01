package egoscale

// BucketUsage represents the usage (in bytes) for a bucket
type BucketUsage struct {
	Created string `json:"created"`
	Name    string `json:"name"`
	Region  string `json:"region"`
	Usage   int64  `json:"usage"`
}

// ListBucketsUsage represents a listBucketsUsage API request
type ListBucketsUsage struct {
	_ bool `name:"listBucketsUsage" description:"List"`
}

// ListBucketsUsageResponse represents a listBucketsUsage API response
type ListBucketsUsageResponse struct {
	Count        int           `json:"count"`
	BucketsUsage []BucketUsage `json:"bucketsusage"`
}

// Response returns the struct to unmarshal
func (ListBucketsUsage) Response() interface{} {
	return new(ListBucketsUsageResponse)
}
