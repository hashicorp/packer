package common

// IsValidRegion returns true if the supplied region is a valid AWS
// region and false if it's not.
func ValidateRegion(region string) bool {
	var regions = [11]string{"us-east-1", "us-west-2", "us-west-1", "eu-west-1",
		"eu-central-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
		"sa-east-1", "cn-north-1", "us-gov-west-1"}

	for _, valid := range regions {
		if region == valid {
			return true
		}
	}
	return false
}
