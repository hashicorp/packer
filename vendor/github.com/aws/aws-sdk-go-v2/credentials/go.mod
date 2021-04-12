module github.com/aws/aws-sdk-go-v2/credentials

go 1.15

require (
	github.com/aws/aws-sdk-go-v2 v1.2.1
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.0.3
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.0.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.1.2
	github.com/aws/aws-sdk-go-v2/service/sts v1.1.2
	github.com/aws/smithy-go v1.2.0
	github.com/google/go-cmp v0.5.4
)

replace (
	github.com/aws/aws-sdk-go-v2 => ../
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds => ../feature/ec2/imds/
	github.com/aws/aws-sdk-go-v2/service/sts => ../service/sts/
)

replace github.com/aws/aws-sdk-go-v2/service/internal/presigned-url => ../service/internal/presigned-url/

replace github.com/aws/aws-sdk-go-v2/service/sso => ../service/sso/
