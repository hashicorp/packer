package triton

import (
	"testing"
)

func TestAccessConfig_Prepare(t *testing.T) {
	ac := testAccessConfig()
	errs := ac.Prepare(nil)
	if errs != nil {
		t.Fatal("should not error")
	}

	ac = testAccessConfig()
	ac.Account = ""
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should error")
	}

	ac = testAccessConfig()
	ac.KeyID = ""
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should error")
	}
}

func testAccessConfig() AccessConfig {
	return AccessConfig{
		Endpoint:    "test-endpoint",
		Account:     "test-account",
		KeyID:       "c5:9d:37:d2:28:d3:ef:39:1b:0a:0e:37:d5:b4:7c:59",
		KeyMaterial: testKeyMaterial,
	}
}

const testKeyMaterial = `-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEAuujYxDCBu9W6es650mLcy9RaLGqjHT2KXPs4fHVr1sfBxPDk
ChpekrVEfE69wpf7/oduQwLmBTIBBNtr/aH5e8gt2uCe1kD6swjnAG+nWZB63/BW
XF9zFFE/Vs/dOyHIkqoLhVHurYYFBGqDXH4w1N02vfyQaH/VQDumF9ZiH6b9u28/
WtrHSLPeGidgrt3csh0Q4Bydm2xSCx4Kfeouv0rM25mFoiq/QaTXkfWS0sIzrhhU
rXL4N6B8tBRojHghpjh8LjG4ufJ7Q0QMWfeBfTqQ9llaEtiMIBznyq+oF7vwv0pc
Cw2eXcURfg/9e5M8S3gSthkqGN9NjQUSeNsgCQIDAQABAoIBAQCcy6zcmFyc+GTh
lP5psanMDC5BSIvhgbjK26y9K7v1h8nTrsl+eDSSGiKDrYKe9eTd1zr2WD4iaZpV
OsVTFkg2QO3Gydw1nHkzK+qtgP0As6WAqxunjiL6DlZ2OxY5/tNFxgS4KM1zIBSh
acEdHHdWeuTraC60m1iH9AIXyS6zoW+YvKr3Cu+gjQgDxg90Uzx7gB7/tAT9uTCG
NHXRCJFrjLlKwWap5QpbbrEMZXjwwb4FEC6KOWaTHDGtB6V2NHBYfpAucuLXx19H
jKUnogZHxTFbYwf7oZSVCR6tUm/Dytq0DmZv+wkCtUSqP0hljqO71yOOMiWA7fVq
4cyD8TGJAoGBAPVVebrIEm8gMTt7s37YFwZXXcC2GT/jVi5bOhWhL5zOA0VxJMH7
hUmrRNffNbFciSUM7CeSqh7Zq9v2l+qCjy9iZfUN6NY1s7d5q7NVkVX/GBuZ8ULp
d81L4ifnr9KsEIzWz8X3Y/efO/20YqoEqLJm6qUyZYHWJbv9Z8Cteef7AoGBAMMJ
HkzRop/VAi5RFzuJyNEzeOyveGpngtsnAY/UcEeWoxkUfHy/GAZAH8aAz9FqLaBv
xGL++g3l8nHtov+gkrnJpK/f0QEWY+PbRWxSRHLW0rBdQJRB8nisNrWJwj4ysNhj
ejYgBfSSmwkLBnvjNce6RwtZ5d+VRFGRl63CfMTLAoGBAK7Vaxqg2gI3ft5VGWWb
uUzblgRvwS62ZARFHu+rHrMwXURvjTJwfFwzoav1dd4fg9zTiLfq3TF/DeqDoV+O
C1xJUz9/2h5NxvVJ0ALNR/VxBU0mN7jniGjVWyX1BmesF19G9mquEp+06puyoV1o
VJBOp4lykMQmSF3gCMBW4DlhAoGBAINdauk28iBRqqxjthBGF9rAnpxc+/A/VCYk
OasU3aN6VNSZtdeYJqhfHIfpTxCwQZckcNR1BRvDW+9crkMbdnho1uIXEIF5AUMB
99qj9rKa+0ILLWoumRCqfhb8eLbIEdFN/4zhOOGotX/7yxw6x4iFcUC2Blz3/xIp
zE4fB0bNAoGBAK/ms0TixeqdNCY8WVHjT6H1W34Y1gvMlwSDodlUk0v8rUHAK4bO
TKvdCYbxCQ+kqAFlicY1ZwxoeJzW6k3K+kJ+qWBn0yH4W61M8uKOIvciu1w1CXxG
XZHg281yLxOfJj9YnPG73+sZFucyhtNPiq/1pR4tpm6YLMk8KSTy7XU5
-----END RSA PRIVATE KEY-----`
