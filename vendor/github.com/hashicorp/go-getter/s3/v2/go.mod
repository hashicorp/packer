module github.com/hashicorp/go-getter/s3/v2

go 1.14

replace github.com/hashicorp/go-getter/v2 => ../

require (
	github.com/aws/aws-sdk-go v1.30.8
	github.com/hashicorp/go-getter/v2 v2.0.0-20200511090339-3107ec4af37a
)
