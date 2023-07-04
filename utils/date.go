package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var parseDateRegex *regexp.Regexp

func init() {
	parseDateRegex = regexp.MustCompile(`^(\d+)-(\d+)-(\d+)$`)
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
