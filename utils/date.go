package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var parseDateRegex *regexp.Regexp
var parseDateTimeRegex *regexp.Regexp

func init() {
	parseDateRegex = regexp.MustCompile(`^(\d+)-(\d+)-(\d+)$`)
	parseDateTimeRegex = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})Z$`)
}

func ParseDate(jalaliDate string) (time.Time, error) {
	dd := parseDateRegex.FindAllStringSubmatch(jalaliDate, -1)
	if len(dd) != 1 {
		return time.Time{}, errors.New("parse date failed")
	}

	jY, _ := strconv.Atoi(dd[0][1])
	jM, _ := strconv.Atoi(dd[0][2])
	jD, _ := strconv.Atoi(dd[0][3])

	date, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%02d-%02d", jY, jM, jD))
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
func ParseDateTime(datetime string) (time.Time, error) {
	dd := parseDateTimeRegex.FindAllStringSubmatch(datetime, -1)
	if len(dd) != 1 {
		return time.Time{}, errors.New("parse datetime failed")
	}

	year, _ := strconv.Atoi(dd[0][1])
	month, _ := strconv.Atoi(dd[0][2])
	day, _ := strconv.Atoi(dd[0][3])
	hour, _ := strconv.Atoi(dd[0][4])
	minute, _ := strconv.Atoi(dd[0][5])
	second, _ := strconv.Atoi(dd[0][6])

	parsedDatetime, err := time.Parse("2006-01-02T15:04:05Z", fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ", year, month, day, hour, minute, second))
	if err != nil {
		return time.Time{}, err
	}

	return parsedDatetime, nil
}
