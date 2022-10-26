package ip_manager

import (
	"time"
)

func CreateTransactionTimestamp() time.Time {
	return now()
}

func parseTransactionTimestamp(timeStampAnnotation string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, timeStampAnnotation)
}
