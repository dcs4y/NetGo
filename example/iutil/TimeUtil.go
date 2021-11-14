package iutil

import "time"

func GetFormatTime() string {
	return time.Now().Format(TimeFormat)
}
