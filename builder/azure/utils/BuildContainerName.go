package utils

import (
	"time"
	"fmt"
)

func BuildContainerName() string{
	now := time.Now()
	y,m,d := now.Date()
	return fmt.Sprintf("packer-provision-%d-%d-%d-%d-%d-%d", now.Hour(), now.Minute(), now.Second(), d,m,y )
}

