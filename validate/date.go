package validate

import (
	"strconv"
	"strings"
	"time"

	"github.com/hhrutter/pdfcpu/types"
)

func prevalidate(s string) (string, bool) {

	// utf16 conversion if applicable.
	if types.IsStringUTF16BE(s) {
		utf16s, err := types.DecodeUTF16String(s)
		if err != nil {
			return "", false
		}
		s = utf16s
	}

	// "D:YYYY" is mandatory
	if len(s) < 6 {
		return "", false
	}

	return s, strings.HasPrefix(s, "D:")
}

func validateTimezoneMinutes(s string, o byte) bool {

	tzmin := s[20:22]
	logDebugValidate.Printf("validateTimezoneMinutes: tz minutes offset string = <%s>\n", tzmin)
	tzm, err := strconv.Atoi(tzmin)
	if err != nil {
		return false
	}

	if tzm > 59 {
		return false
	}

	if o == 'Z' && tzm != 0 {
		return false
	}

	logDebugValidate.Printf("validateTimezoneMinutes: returning %v\n", true)

	// "D:YYYYMMDDHHmmSSZHH'mm"
	if len(s) == 22 {
		return false
	}

	// Accept a trailing '
	return s[22] == '\''
}

func validateTimezone(s string) bool {

	o := s[16]
	logDebugValidate.Printf("timezone operator:%s\n", string(o))

	if o != '+' && o != '-' && o != 'Z' {
		return false
	}

	// local time equal to UT.
	// "D:YYYYMMDDHHmmSSZ"
	if o == 'Z' && len(s) == 17 {
		return true
	}

	if len(s) < 20 {
		return false
	}

	tzhours := s[17:19]
	logDebugValidate.Printf("validateTimezone: tz hour offset string = <%s>\n", tzhours)
	tzh, err := strconv.Atoi(tzhours)
	if err != nil {
		return false
	}

	if tzh > 23 {
		return false
	}

	if o == 'Z' && tzh != 0 {
		return false
	}

	if s[19] != '\'' {
		return false
	}

	// "D:YYYYMMDDHHmmSSZHH'"
	if len(s) == 20 {
		return true
	}

	if len(s) != 22 && len(s) != 23 {
		return false
	}

	return validateTimezoneMinutes(s, o)
}

func validateYear(s string) (y int, finished, ok bool) {

	year := s[2:6]

	logDebugValidate.Printf("validateYear: year string = <%s>\n", year)

	y, err := strconv.Atoi(year)
	if err != nil {
		return 0, false, false
	}

	// "D:YYYY"
	if len(s) == 6 {
		return 0, true, true
	}

	if len(s) == 7 {
		return 0, false, false
	}

	return y, false, true
}

func validateMonth(s string) (m int, finished, ok bool) {

	month := s[6:8]

	logDebugValidate.Printf("validateMonth: month string = <%s>\n", month)

	var err error
	m, err = strconv.Atoi(month)
	if err != nil {
		return 0, false, false
	}

	if m < 1 || m > 12 {
		return 0, false, false
	}

	// "D:YYYYMM"
	if len(s) == 8 {
		return m, true, true
	}

	if len(s) == 9 {
		return 0, false, false
	}

	return m, false, true
}

func validateDay(s string, y, m int) (finished, ok bool) {

	day := s[8:10]

	logDebugValidate.Printf("validateDay: day string = <%s>\n", day)

	d, err := strconv.Atoi(day)
	if err != nil {
		return false, false
	}

	if d < 1 || d > 31 {
		return false, false
	}

	// check valid Date(year,month,day)
	t := time.Date(y, time.Month(m+1), 0, 0, 0, 0, 0, time.UTC)
	logDebugValidate.Printf("last day of month is %d\n", t.Day())

	if d > t.Day() {
		return false, false
	}

	// "D:YYYYMMDD"
	if len(s) == 10 {
		return true, true
	}

	if len(s) == 11 {
		return false, false
	}

	return false, true
}

func validateHour(s string) (finished, ok bool) {

	hour := s[10:12]

	logDebugValidate.Printf("validateHour: hour string = <%s>\n", hour)

	h, err := strconv.Atoi(hour)
	if err != nil {
		return false, false
	}

	if h > 23 {
		return false, false
	}

	// "D:YYYYMMDDHH"
	if len(s) == 12 {
		return true, true
	}

	if len(s) == 13 {
		return false, false
	}

	return false, true
}

func validateMinute(s string) (finished, ok bool) {

	minute := s[12:14]

	logDebugValidate.Printf("validateMinute: minute string = <%s>\n", minute)

	min, err := strconv.Atoi(minute)
	if err != nil {
		return false, false
	}

	if min > 59 {
		return false, false
	}

	// "D:YYYYMMDDHHmm"
	if len(s) == 14 {
		return true, true
	}

	if len(s) == 15 {
		return false, false
	}

	return false, true
}

func validateSecond(s string) (finished, ok bool) {

	second := s[14:16]

	logDebugValidate.Printf("validateSecond: second string = <%s>\n", second)

	sec, err := strconv.Atoi(second)
	if err != nil {
		return false, false
	}

	if sec > 59 {
		return false, false
	}

	// "D:YYYYMMDDHHmmSS"
	if len(s) == 16 {
		return true, true
	}

	return false, true
}

// Date validates an ISO/IEC 8824 compliant date string.
func validateDate(s string) bool {

	// 7.9.4 Dates
	// (D:YYYYMMDDHHmmSSOHH'mm')

	logDebugValidate.Printf("validateDate(%s)\n", s)

	var ok bool
	s, ok = prevalidate(s)
	if !ok {
		return false
	}

	y, finished, ok := validateYear(s)
	if !ok {
		return false
	}
	if finished {
		return true
	}

	m, finished, ok := validateMonth(s)
	if !ok {
		return false
	}
	if finished {
		return true
	}

	finished, ok = validateDay(s, y, m)
	if !ok {
		return false
	}
	if finished {
		return true
	}

	finished, ok = validateHour(s)
	if !ok {
		return false
	}
	if finished {
		return true
	}

	finished, ok = validateMinute(s)
	if !ok {
		return false
	}
	if finished {
		return true
	}

	finished, ok = validateSecond(s)
	if !ok {
		return false
	}
	if finished {
		return true
	}

	// Process timezone
	return validateTimezone(s)
}
