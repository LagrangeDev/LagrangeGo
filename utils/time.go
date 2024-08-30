package utils

import (
	"fmt"
	"time"
)

func GenerateUinTimestamp(uin uint32) string {
	currentTime := time.Now()
	formattedTime := currentTime.Format("0102150405")
	milliseconds := currentTime.Nanosecond() / 1000000
	return fmt.Sprintf("%d_%s%02d_%d", uin, formattedTime, currentTime.Year()%100, milliseconds)
}
