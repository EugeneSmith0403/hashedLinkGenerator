package helpers

import "time"

func ToHours(hours int) time.Duration {
	return time.Hour * time.Duration(hours)
}

func ToMinutes(minutes int) time.Duration {
	return time.Minute * time.Duration(minutes)
}
