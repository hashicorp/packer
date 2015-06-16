package utils

import (
	"time"
	"fmt"
)

func DecorateImageName(currentName string) string{
	now := time.Now()
	y,m,d := now.Date()
	return fmt.Sprintf("%s_%v-%v-%v_%v-%v", currentName, y,m,d, now.Hour(), now.Minute() )
}


