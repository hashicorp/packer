package vagrant

import (
	"testing"
)

func TestAWSProvider_impl(t *testing.T) {
	var _ Provider = new(AWSProvider)
}
