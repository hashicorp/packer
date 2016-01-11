package common

// ValidateRegion returns true if the supplied region is a valid AWS
// region and false if it's not.
func ValidateRegion(region string) bool {
	var regions = [12]string{
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"cn-north-1",
		"eu-central-1",
		"eu-west-1",
		"sa-east-1",
		"us-east-1",
		"us-gov-west-1",
		"us-west-1",
		"us-west-2",
	}

	for _, valid := range regions {
		if region == valid {
			return true
		}
	}
	return false
}
