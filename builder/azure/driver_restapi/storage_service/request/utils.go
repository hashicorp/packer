package request

import "time"

func currentTimeRfc1123Formatted() string {
	t := time.Now().UTC()
	return t.Format(dateLayout)
}
