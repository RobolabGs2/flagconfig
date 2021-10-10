package flagconfig

import (
	"strconv"
	"time"
)

func defaultBool(str string) (bool, error) {
	if str == "" {
		return false, nil
	}
	return strconv.ParseBool(str)
}

func defaultDuration(str string) (time.Duration, error) {
	if str == "" {
		return 0, nil
	}
	return time.ParseDuration(str)
}

func defaultFloat64(str string) (float64, error) {
	if str == "" {
		return 0, nil
	}
	return strconv.ParseFloat(str, 64)
}

func defaultInt(str string) (int, error) {
	if str == "" {
		return 0, nil
	}
	parseInt, err := strconv.ParseInt(str, 0, 0)
	return int(parseInt), err
}

func defaultInt64(str string) (int64, error) {
	if str == "" {
		return 0, nil
	}
	parseInt, err := strconv.ParseInt(str, 0, 64)
	return parseInt, err
}

func defaultUint(str string) (uint, error) {
	if str == "" {
		return 0, nil
	}
	parseInt, err := strconv.ParseUint(str, 0, 0)
	return uint(parseInt), err
}

func defaultUint64(str string) (uint64, error) {
	if str == "" {
		return 0, nil
	}
	parseInt, err := strconv.ParseUint(str, 0, 64)
	return parseInt, err
}
