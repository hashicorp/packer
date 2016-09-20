package common

func listEC2Regions() []string {
	return []string{
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-south-1",
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
}

// ValidateRegion returns true if the supplied region is a valid AWS
// region and false if it's not.
func ValidateRegion(region string) bool {
	for _, valid := range listEC2Regions() {
		if region == valid {
			return true
		}
	}
	return false
}
