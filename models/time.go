package models

import (
	"strconv"
	"time"
)

/* XXX: This should move elsewhere */
func parseTimestamp(str string) (time.Time, error) {
	unix, er := strconv.ParseInt(str, 10, 64)
	if er != nil {
		return time.Time{}, er
	}

	return time.Unix(unix, 0).UTC(), nil
}

func unparseTimestamp(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

func unparseTimestampOptional(t time.Time) *string {
	if t.IsZero() {
		return nil
	}

	var result string
	result = unparseTimestamp(t)
	return &result
}

func parseBool(i int64) bool {
	return i != 0
}

func unparseBool(b bool) int64 {
	if b {
		return 1
	}

	return 0
}
