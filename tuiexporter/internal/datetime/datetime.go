package datetime

import "time"

const (
	simple = "2006-01-02 15:04:05"
	full   = "2006-01-02 15:04:05.000000Z07:00"
)

// GetSimpleTime returns a string representation of the time in the format "2006-01-02 15:04:05".
func GetSimpleTime(t time.Time) string {
	return t.Format(simple)
}

// GetFullTime returns a string representation of the time in the format "2006-01-02 15:04:05.000000Z07:00".
func GetFullTime(t time.Time) string {
	return t.Format(full)
}
