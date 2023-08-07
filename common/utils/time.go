package utils

import (
	"fmt"
	"strings"
)

const (
	DefaultTimeLayout = "2006-01-02 15:04:05"
)

func TimeFormatToKafka(file string) string {
	timeStr := strings.Replace(strings.Split(file, "+")[0], "T", " ", -1)
	return fmt.Sprintf("%s.000", timeStr)
}
