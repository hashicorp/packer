package tencent

import (
	"testing"
)

func TestDescribeProject(t *testing.T) {
	var c Config
	requiredEnvVars := map[string]string{CRegion: "ap-singapore"}
	GetRequiredEnvVars(requiredEnvVars)
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]

	type args struct {
		c *Config
	}
	tests := []struct {
		name string
		args args
	}{
		{"DescribeProject test case 1", args{&c}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DescribeProject(tt.args.c)
		})
	}
}
