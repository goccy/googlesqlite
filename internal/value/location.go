package value

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	timeZoneOffsetPartialPattern = regexp.MustCompile(`([-+]\d{2})`)
	timeZoneOffsetPattern        = regexp.MustCompile(`([-+]\d{2}):(\d{2})`)
	locationCacheMap             = map[string]*time.Location{}
	locationCacheMu              sync.RWMutex
)

func getCachedLocation(timeZone string) *time.Location {
	locationCacheMu.RLock()
	defer locationCacheMu.RUnlock()
	if loc, exists := locationCacheMap[timeZone]; exists {
		return loc
	}
	if loc, exists := locationCacheMap[fmt.Sprintf("UTC%s", timeZone)]; exists {
		return loc
	}
	return nil
}

func setLocationCache(key string, loc *time.Location) {
	locationCacheMu.Lock()
	locationCacheMap[key] = loc
	locationCacheMu.Unlock()
}

func toLocation(timeZone string) (*time.Location, error) {
	if loc := getCachedLocation(timeZone); loc != nil {
		return loc, nil
	}
	if matched := timeZoneOffsetPattern.FindAllStringSubmatch(timeZone, -1); len(matched) != 0 && len(matched[0]) == 3 {
		offsetHour := matched[0][1]
		offsetMin := matched[0][2]
		// The regexp captures at most a sign plus two digits, so the
		// parsed values always fit in 32 bits; parsing with a 32-bit
		// width keeps the int conversions below provably in range.
		hour, err := strconv.ParseInt(offsetHour, 10, 32)
		if err != nil {
			return nil, err
		}
		min, err := strconv.ParseInt(offsetMin, 10, 32)
		if err != nil {
			return nil, err
		}
		locName := fmt.Sprintf("UTC%s", timeZone)
		loc := time.FixedZone(locName, int(hour)*60*60+int(min)*60)
		setLocationCache(locName, loc)
		return loc, nil
	}
	if matched := timeZoneOffsetPartialPattern.FindAllStringSubmatch(timeZone, -1); len(matched) != 0 && len(matched[0]) == 2 {
		offset := matched[0][1]
		hour, err := strconv.ParseInt(offset, 10, 32)
		if err != nil {
			return nil, err
		}
		locName := fmt.Sprintf("UTC%s", timeZone)
		loc := time.FixedZone(locName, int(hour)*60*60)
		setLocationCache(locName, loc)
		return loc, nil
	}

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, fmt.Errorf("failed to load location from %s: %w", timeZone, err)
	}
	setLocationCache(timeZone, loc)
	return loc, nil
}

func modifyTimeZone(t time.Time, loc *time.Location) (time.Time, error) {
	// remove timezone parameter from time
	format := t.Format("2006-01-02T15:04:05.999999999")
	return parseTimestamp(format, loc)
}
func timeFromUnixNano(unixNano int64) time.Time {
	return time.Unix(0, unixNano)
}
