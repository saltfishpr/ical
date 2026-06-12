package ical

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	minRFC5545Integer = -2147483648
	maxRFC5545Integer = 2147483647
)

// FormatBoolean formats an RFC 5545 BOOLEAN value, returning "TRUE" or "FALSE".
func FormatBoolean(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
}

// ParseBoolean parses an RFC 5545 BOOLEAN value, accepting "TRUE" or "FALSE"
// case-insensitively.
func ParseBoolean(value string) (bool, error) {
	switch strings.ToUpper(value) {
	case "TRUE":
		return true, nil
	case "FALSE":
		return false, nil
	default:
		return false, fmt.Errorf("ical: expected TRUE or FALSE, got %q", value)
	}
}

// FormatInteger formats an RFC 5545 INTEGER value.
func FormatInteger(value int) string {
	return strconv.Itoa(value)
}

// ParseInteger parses an RFC 5545 INTEGER value.
func ParseInteger(value string) (int, error) {
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	if parsed < minRFC5545Integer || parsed > maxRFC5545Integer {
		return 0, fmt.Errorf("ical: INTEGER out of range")
	}
	return int(parsed), nil
}

// FormatFloat formats an RFC 5545 FLOAT value.
func FormatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// ParseFloat parses an RFC 5545 FLOAT value.
func ParseFloat(value string) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, fmt.Errorf("ical: FLOAT must be finite")
	}
	return parsed, nil
}

// Geo is an RFC 5545 GEO value in latitude, longitude order.
type Geo struct {
	Latitude  float64
	Longitude float64
}

// String serializes the GEO value.
func (g Geo) String() string {
	return FormatFloat(g.Latitude) + ";" + FormatFloat(g.Longitude)
}

// ParseGeo parses an RFC 5545 GEO value.
func ParseGeo(value string) (Geo, error) {
	latitudeRaw, longitudeRaw, ok := strings.Cut(value, ";")
	if !ok {
		return Geo{}, fmt.Errorf("ical: GEO must be latitude;longitude")
	}
	latitude, err := ParseFloat(latitudeRaw)
	if err != nil {
		return Geo{}, err
	}
	longitude, err := ParseFloat(longitudeRaw)
	if err != nil {
		return Geo{}, err
	}
	return Geo{Latitude: latitude, Longitude: longitude}, nil
}

// FormatCalAddress formats a CAL-ADDRESS value, normalizing bare email
// addresses to mailto: URIs. Values already beginning with "mailto:" are
// returned unchanged.
func FormatCalAddress(value string) string {
	if len(value) >= len("mailto:") && strings.EqualFold(value[:len("mailto:")], "mailto:") {
		return value
	}
	return "mailto:" + value
}

// Time is an RFC 5545 TIME value. UTC marks the Z suffix form; TZID is carried
// by the property parameter rather than this value.
type Time struct {
	Hour   int
	Minute int
	Second int
	UTC    bool
}

// String serializes the TIME value.
func (t Time) String() string {
	suffix := ""
	if t.UTC {
		suffix = "Z"
	}
	return fmt.Sprintf("%02d%02d%02d%s", t.Hour, t.Minute, t.Second, suffix)
}

// ParseTime parses an RFC 5545 TIME value.
func ParseTime(value string) (Time, error) {
	utc := strings.HasSuffix(value, "Z")
	if utc {
		value = strings.TrimSuffix(value, "Z")
	}
	if len(value) != 6 {
		return Time{}, fmt.Errorf("ical: TIME must be HHMMSS")
	}
	hour, err := strconv.Atoi(value[0:2])
	if err != nil {
		return Time{}, err
	}
	minute, err := strconv.Atoi(value[2:4])
	if err != nil {
		return Time{}, err
	}
	second, err := strconv.Atoi(value[4:6])
	if err != nil {
		return Time{}, err
	}
	if hour > 23 || minute > 59 || second > 60 {
		return Time{}, fmt.Errorf("ical: invalid TIME %q", value)
	}
	return Time{Hour: hour, Minute: minute, Second: second, UTC: utc}, nil
}

// FormatBinary formats bytes as an RFC 5545 BINARY base64 value.
func FormatBinary(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

// ParseBinary parses an RFC 5545 BINARY base64 value.
func ParseBinary(value string) ([]byte, error) {
	return base64.StdEncoding.Strict().DecodeString(value)
}

// FormatDateTime formats a DATE-TIME value. UTC times use the RFC 5545 Z
// suffix; floating/local times are serialized without a zone suffix.
func FormatDateTime(value time.Time) string {
	_, offset := value.Zone()
	if offset == 0 {
		return value.UTC().Format("20060102T150405Z")
	}
	return value.Format("20060102T150405")
}

// ParseDateTime parses UTC or floating RFC 5545 DATE-TIME values.
func ParseDateTime(value string) (time.Time, error) {
	if strings.HasSuffix(value, "Z") {
		return time.Parse("20060102T150405Z", value)
	}
	return time.Parse("20060102T150405", value)
}

// FormatDuration formats an RFC 5545 DURATION value.
func FormatDuration(value time.Duration) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	total := int64(value / time.Second)
	weeks := total / (7 * 24 * 60 * 60)
	if weeks > 0 && total%(7*24*60*60) == 0 {
		return fmt.Sprintf("%sP%dW", sign, weeks)
	}
	days := total / (24 * 60 * 60)
	total %= 24 * 60 * 60
	hours := total / (60 * 60)
	total %= 60 * 60
	minutes := total / 60
	seconds := total % 60
	var b strings.Builder
	b.WriteString(sign)
	b.WriteByte('P')
	if days > 0 {
		fmt.Fprintf(&b, "%dD", days)
	}
	if hours > 0 || minutes > 0 || seconds > 0 || days == 0 {
		b.WriteByte('T')
		if hours > 0 {
			fmt.Fprintf(&b, "%dH", hours)
		}
		if minutes > 0 {
			fmt.Fprintf(&b, "%dM", minutes)
		}
		if seconds > 0 || (days == 0 && hours == 0 && minutes == 0) {
			fmt.Fprintf(&b, "%dS", seconds)
		}
	}
	return b.String()
}

// ParseDuration parses an RFC 5545 DURATION value.
// Valid forms include "P1W", "P3D", "PT5H30M", "PT0S", and "-PT1H".
func ParseDuration(value string) (time.Duration, error) {
	sign := int64(1)
	if strings.HasPrefix(value, "-") {
		sign = -1
		value = value[1:]
	}
	if !strings.HasPrefix(value, "P") {
		return 0, fmt.Errorf("ical: duration must start with P")
	}
	value = value[1:]
	if value == "" {
		return 0, fmt.Errorf("ical: duration has no components after P")
	}
	inTime := false
	var num strings.Builder
	var seconds int64
	parsed := false
	for _, r := range value {
		if r >= '0' && r <= '9' {
			num.WriteRune(r)
			continue
		}
		if r == 'T' {
			inTime = true
			continue
		}
		if num.Len() == 0 {
			return 0, fmt.Errorf("ical: duration unit %q has no number", r)
		}
		n, _ := strconv.ParseInt(num.String(), 10, 64)
		num.Reset()
		parsed = true
		switch r {
		case 'W':
			seconds += n * 7 * 24 * 60 * 60
		case 'D':
			seconds += n * 24 * 60 * 60
		case 'H':
			if !inTime {
				return 0, fmt.Errorf("ical: H appears outside time duration")
			}
			seconds += n * 60 * 60
		case 'M':
			if !inTime {
				return 0, fmt.Errorf("ical: M appears outside time duration")
			}
			seconds += n * 60
		case 'S':
			if !inTime {
				return 0, fmt.Errorf("ical: S appears outside time duration")
			}
			seconds += n
		default:
			return 0, fmt.Errorf("ical: unsupported duration unit %q", r)
		}
	}
	if num.Len() != 0 {
		return 0, fmt.Errorf("ical: trailing duration number")
	}
	if !parsed {
		return 0, fmt.Errorf("ical: duration has no components")
	}
	return time.Duration(sign*seconds) * time.Second, nil
}

// FormatUTCOffset formats an RFC 5545 UTC-OFFSET value.
func FormatUTCOffset(value time.Duration) string {
	sign := "+"
	if value < 0 {
		sign = "-"
		value = -value
	}
	total := int64(value / time.Second)
	hours := total / 3600
	total %= 3600
	minutes := total / 60
	seconds := total % 60
	if seconds > 0 {
		return fmt.Sprintf("%s%02d%02d%02d", sign, hours, minutes, seconds)
	}
	return fmt.Sprintf("%s%02d%02d", sign, hours, minutes)
}

// ParseUTCOffset parses an RFC 5545 UTC-OFFSET value.
func ParseUTCOffset(value string) (time.Duration, error) {
	if len(value) != 5 && len(value) != 7 {
		return 0, fmt.Errorf("ical: UTC offset must be +HHMM or +HHMMSS")
	}
	sign := value[0]
	if sign != '+' && sign != '-' {
		return 0, fmt.Errorf("ical: UTC offset must start with + or -")
	}
	hours, err := strconv.Atoi(value[1:3])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(value[3:5])
	if err != nil {
		return 0, err
	}
	seconds := 0
	if len(value) == 7 {
		seconds, err = strconv.Atoi(value[5:7])
		if err != nil {
			return 0, err
		}
	}
	if hours >= 24 || minutes >= 60 || seconds >= 60 {
		return 0, fmt.Errorf("ical: invalid UTC offset %q", value)
	}
	if sign == '-' && hours == 0 && minutes == 0 && seconds == 0 {
		return 0, fmt.Errorf("ical: -0000 UTC offset is not allowed")
	}
	d := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	if sign == '-' {
		return -d, nil
	}
	return d, nil
}

// Period is an RFC 5545 PERIOD value. Use End for explicit periods or
// Duration for start/duration periods.
type Period struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
}

// String serializes the PERIOD value.
func (p Period) String() string {
	if p.Duration != 0 {
		return FormatDateTime(p.Start) + "/" + FormatDuration(p.Duration)
	}
	return FormatDateTime(p.Start) + "/" + FormatDateTime(p.End)
}

// ParsePeriod parses one RFC 5545 PERIOD value.
func ParsePeriod(value string) (Period, error) {
	startRaw, endRaw, ok := strings.Cut(value, "/")
	if !ok {
		return Period{}, fmt.Errorf("ical: invalid PERIOD %q", value)
	}
	start, err := ParseDateTime(startRaw)
	if err != nil {
		return Period{}, err
	}
	if strings.HasPrefix(endRaw, "P") || strings.HasPrefix(endRaw, "+P") || strings.HasPrefix(endRaw, "-P") {
		duration, err := ParseDuration(endRaw)
		if err != nil {
			return Period{}, err
		}
		if duration <= 0 {
			return Period{}, fmt.Errorf("ical: PERIOD duration must be positive")
		}
		return Period{Start: start, Duration: duration}, nil
	}
	end, err := ParseDateTime(endRaw)
	if err != nil {
		return Period{}, err
	}
	if !start.Before(end) {
		return Period{}, fmt.Errorf("ical: PERIOD start must be before end")
	}
	return Period{Start: start, End: end}, nil
}

// ParsePeriodList parses comma-separated PERIOD values.
func ParsePeriodList(value string) ([]Period, error) {
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	periods := make([]Period, 0, len(parts))
	for _, part := range parts {
		period, err := ParsePeriod(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		periods = append(periods, period)
	}
	return periods, nil
}
