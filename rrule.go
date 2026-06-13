package ical

import (
	"fmt"
	"iter"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Frequency is an RFC 5545 FREQ value.
type Frequency string

const (
	Secondly Frequency = "SECONDLY"
	Minutely Frequency = "MINUTELY"
	Hourly   Frequency = "HOURLY"
	Daily    Frequency = "DAILY"
	Weekly   Frequency = "WEEKLY"
	Monthly  Frequency = "MONTHLY"
	Yearly   Frequency = "YEARLY"
)

// Weekday is an RFC 5545 BYDAY weekday token.
type Weekday string

const (
	Monday    Weekday = "MO"
	Tuesday   Weekday = "TU"
	Wednesday Weekday = "WE"
	Thursday  Weekday = "TH"
	Friday    Weekday = "FR"
	Saturday  Weekday = "SA"
	Sunday    Weekday = "SU"
)

// WeekdayNum is one RFC 5545 BYDAY value. Ordinal is zero for plain weekdays
// such as MO, or ranges from -53..-1 and 1..53 for values such as -1SU or 1MO.
type WeekdayNum struct {
	Ordinal int
	Weekday Weekday
}

// String serializes the BYDAY value, e.g. "MO" or "-1SU".
func (d WeekdayNum) String() string {
	if d.Ordinal == 0 {
		return string(d.Weekday)
	}
	return strconv.Itoa(d.Ordinal) + string(d.Weekday)
}

// ParseWeekdayNum parses one RFC 5545 BYDAY weekday value.
func ParseWeekdayNum(value string) (WeekdayNum, error) {
	value = strings.ToUpper(value)
	if len(value) < 2 {
		return WeekdayNum{}, fmt.Errorf("ical: invalid BYDAY value %q", value)
	}
	weekday := Weekday(value[len(value)-2:])
	if !validWeekday(weekday) {
		return WeekdayNum{}, fmt.Errorf("ical: invalid BYDAY weekday %q", weekday)
	}
	prefix := value[:len(value)-2]
	if prefix == "" {
		return WeekdayNum{Weekday: weekday}, nil
	}
	ordinal, err := strconv.Atoi(prefix)
	if err != nil {
		return WeekdayNum{}, err
	}
	if ordinal == 0 || ordinal < -53 || ordinal > 53 {
		return WeekdayNum{}, fmt.Errorf("ical: BYDAY ordinal %d out of range -53..53 excluding 0", ordinal)
	}
	return WeekdayNum{Ordinal: ordinal, Weekday: weekday}, nil
}

// RecurrenceRule represents the common RRULE fields defined by RFC 5545 Section 3.3.10.
//
// Fields:
//   - Frequency: the recurrence frequency (required).
//   - Count: maximum number of occurrences (mutually exclusive with Until).
//   - Interval: interval multiplier; defaults to 1 when zero.
//   - Until: inclusive end date (mutually exclusive with Count).
//   - ByDay/ByDayNum: plain or ordinal weekdays for BYDAY.
//   - ByMonthDay, ByYearDay, ByWeekNo, ByMonth, BySetPos: numeric BY* parts.
//   - BySecond, ByMinute, ByHour: time-level BY* parts for sub-daily frequencies.
//   - WeekStart: WKST value; defaults to Monday when empty.
//   - Parts: raw key→value map that preserves RFC parts not yet modeled as typed fields.
//
// Use [ParseRecurrenceRule] to create a RecurrenceRule, or construct one directly
// and pass it to [Component.AddRecurrenceRule].
type RecurrenceRule struct {
	Frequency   Frequency
	Count       int
	Interval    int
	Until       *time.Time
	UntilIsDate bool
	BySecond    []int
	ByMinute    []int
	ByHour      []int
	ByDay       []Weekday
	ByDayNum    []WeekdayNum
	ByMonthDay  []int
	ByYearDay   []int
	ByWeekNo    []int
	ByMonth     []int
	BySetPos    []int
	WeekStart   Weekday
	Parts       map[string][]string
}

// String serializes the recurrence rule to an RFC 5545 RRULE value string.
func (r RecurrenceRule) String() string {
	var parts []string
	if r.Frequency != "" {
		parts = append(parts, "FREQ="+string(r.Frequency))
	}
	if r.Until != nil {
		if r.UntilIsDate {
			parts = append(parts, "UNTIL="+r.Until.Format("20060102"))
		} else {
			parts = append(parts, "UNTIL="+FormatDateTime(*r.Until))
		}
	}
	if r.Count > 0 {
		parts = append(parts, fmt.Sprintf("COUNT=%d", r.Count))
	}
	if r.Interval > 0 {
		parts = append(parts, fmt.Sprintf("INTERVAL=%d", r.Interval))
	}
	if len(r.BySecond) > 0 {
		parts = append(parts, "BYSECOND="+joinInts(r.BySecond))
	}
	if len(r.ByMinute) > 0 {
		parts = append(parts, "BYMINUTE="+joinInts(r.ByMinute))
	}
	if len(r.ByHour) > 0 {
		parts = append(parts, "BYHOUR="+joinInts(r.ByHour))
	}
	if len(r.ByDayNum) > 0 {
		values := make([]string, len(r.ByDayNum))
		for i, d := range r.ByDayNum {
			values[i] = d.String()
		}
		parts = append(parts, "BYDAY="+strings.Join(values, ","))
	} else if len(r.ByDay) > 0 {
		values := make([]string, len(r.ByDay))
		for i, d := range r.ByDay {
			values[i] = string(d)
		}
		parts = append(parts, "BYDAY="+strings.Join(values, ","))
	}
	if len(r.ByMonthDay) > 0 {
		parts = append(parts, "BYMONTHDAY="+joinInts(r.ByMonthDay))
	}
	if len(r.ByYearDay) > 0 {
		parts = append(parts, "BYYEARDAY="+joinInts(r.ByYearDay))
	}
	if len(r.ByWeekNo) > 0 {
		parts = append(parts, "BYWEEKNO="+joinInts(r.ByWeekNo))
	}
	if len(r.ByMonth) > 0 {
		parts = append(parts, "BYMONTH="+joinInts(r.ByMonth))
	}
	if len(r.BySetPos) > 0 {
		parts = append(parts, "BYSETPOS="+joinInts(r.BySetPos))
	}
	if r.WeekStart != "" {
		parts = append(parts, "WKST="+string(r.WeekStart))
	}
	return strings.Join(parts, ";")
}

// ParseRecurrenceRule parses an RFC 5545 RRULE value string into a RecurrenceRule.
// It validates FREQ is present and non-repeating, rejects COUNT+UNTIL coexistence,
// and checks numeric range constraints for each BY* part.
func ParseRecurrenceRule(value string) (RecurrenceRule, error) {
	rule := RecurrenceRule{Parts: map[string][]string{}}
	counts := map[string]int{}
	for _, part := range strings.Split(value, ";") {
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			return rule, fmt.Errorf("ical: invalid RRULE part %q", part)
		}
		key = strings.ToUpper(key)
		counts[key]++
		values := strings.Split(val, ",")
		var err error
		rule.Parts[key] = values
		switch key {
		case "FREQ":
			rule.Frequency = Frequency(strings.ToUpper(val))
		case "COUNT":
			count, err := strconv.Atoi(val)
			if err != nil {
				return rule, err
			}
			if count < 1 {
				return rule, fmt.Errorf("ical: RRULE COUNT must be positive")
			}
			rule.Count = count
		case "INTERVAL":
			interval, err := strconv.Atoi(val)
			if err != nil {
				return rule, err
			}
			if interval < 1 {
				return rule, fmt.Errorf("ical: RRULE INTERVAL must be positive")
			}
			rule.Interval = interval
		case "UNTIL":
			// UNTIL can be DATE or DATE-TIME per RFC 5545 Section 3.3.10.
			if strings.Contains(val, "T") {
				until, err := ParseDateTime(val)
				if err != nil {
					return rule, err
				}
				rule.Until = &until
			} else {
				until, err := time.Parse("20060102", val)
				if err != nil {
					return rule, err
				}
				rule.Until = &until
				rule.UntilIsDate = true
			}
		case "BYDAY":
			for _, value := range values {
				day, err := ParseWeekdayNum(value)
				if err != nil {
					return rule, err
				}
				rule.ByDayNum = append(rule.ByDayNum, day)
				rule.ByDay = append(rule.ByDay, day.Weekday)
			}
		case "BYSECOND":
			rule.BySecond, err = parseIntList(values)
			if err != nil {
				return rule, err
			}
			if err := validateIntRange("BYSECOND", rule.BySecond, 0, 60); err != nil {
				return rule, err
			}
		case "BYMINUTE":
			rule.ByMinute, err = parseIntList(values)
			if err != nil {
				return rule, err
			}
			if err := validateIntRange("BYMINUTE", rule.ByMinute, 0, 59); err != nil {
				return rule, err
			}
		case "BYHOUR":
			rule.ByHour, err = parseIntList(values)
			if err != nil {
				return rule, err
			}
			if err := validateIntRange("BYHOUR", rule.ByHour, 0, 23); err != nil {
				return rule, err
			}
		case "BYMONTHDAY":
			rule.ByMonthDay, err = parseIntList(values)
			if err != nil {
				return rule, err
			}
			if err := validateIntRangeNoZero("BYMONTHDAY", rule.ByMonthDay, -31, 31); err != nil {
				return rule, err
			}
		case "BYYEARDAY":
			rule.ByYearDay, err = parseIntList(values)
			if err != nil {
				return rule, err
			}
			if err := validateIntRangeNoZero("BYYEARDAY", rule.ByYearDay, -366, 366); err != nil {
				return rule, err
			}
		case "BYWEEKNO":
			rule.ByWeekNo, err = parseIntList(values)
			if err != nil {
				return rule, err
			}
			if err := validateIntRangeNoZero("BYWEEKNO", rule.ByWeekNo, -53, 53); err != nil {
				return rule, err
			}
		case "BYMONTH":
			rule.ByMonth, err = parseIntList(values)
			if err != nil {
				return rule, err
			}
			if err := validateIntRange("BYMONTH", rule.ByMonth, 1, 12); err != nil {
				return rule, err
			}
		case "BYSETPOS":
			rule.BySetPos, err = parseIntList(values)
			if err != nil {
				return rule, err
			}
			if err := validateIntRangeNoZero("BYSETPOS", rule.BySetPos, -366, 366); err != nil {
				return rule, err
			}
		case "WKST":
			rule.WeekStart = Weekday(strings.ToUpper(val))
			if !validWeekday(rule.WeekStart) {
				return rule, fmt.Errorf("ical: invalid RRULE WKST %q", rule.WeekStart)
			}
		}
	}
	if counts["FREQ"] != 1 {
		return rule, fmt.Errorf("ical: RRULE requires exactly one FREQ")
	}
	if !validFrequency(rule.Frequency) {
		return rule, fmt.Errorf("ical: invalid RRULE FREQ %q", rule.Frequency)
	}
	for key, count := range counts {
		if count > 1 {
			return rule, fmt.Errorf("ical: RRULE part %s appears more than once", key)
		}
	}
	if counts["COUNT"] > 0 && counts["UNTIL"] > 0 {
		return rule, fmt.Errorf("ical: RRULE cannot contain both COUNT and UNTIL")
	}
	return rule, nil
}

func validFrequency(f Frequency) bool {
	switch f {
	case Secondly, Minutely, Hourly, Daily, Weekly, Monthly, Yearly:
		return true
	default:
		return false
	}
}

func validWeekday(day Weekday) bool {
	switch day {
	case Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday:
		return true
	default:
		return false
	}
}

// Expand returns an iterator over recurrence start instants generated from
// start.
func (r RecurrenceRule) Expand(start time.Time) iter.Seq[time.Time] {
	limit := r.Count
	if limit <= 0 {
		limit = math.MaxInt
	}
	interval := r.Interval
	if interval <= 0 {
		interval = 1
	}
	if len(r.BySetPos) > 0 {
		// BYSETPOS is a post-filter: generate more candidates than needed,
		// then select per-interval. Use a generous multiplier so the base
		// expand doesn't stop early; UNTIL still bounds the iteration.
		base := r.baseExpand(start, interval, limit*len(r.BySetPos)*10+100)
		return applyBySetPos(base, r.BySetPos, r.Frequency, limit)
	}
	return r.baseExpand(start, interval, limit)
}

func (r RecurrenceRule) baseExpand(start time.Time, interval, limit int) iter.Seq[time.Time] {
	switch r.Frequency {
	case Secondly:
		return expandSecondly(start, r, interval, limit)
	case Minutely:
		return expandMinutely(start, r, interval, limit)
	case Hourly:
		return expandHourly(start, r, interval, limit)
	case Daily:
		return expandDaily(start, r, interval, limit)
	case Weekly:
		return expandWeekly(start, r, interval, limit)
	case Monthly:
		return expandMonthly(start, r, interval, limit)
	case Yearly:
		return expandYearly(start, r, interval, limit)
	default:
		return func(yield func(time.Time) bool) {}
	}
}

// expandTimesOn returns all time-of-day candidates for a given date, applying
// BYHOUR, BYMINUTE, and BYSECOND filters. When no filter is specified for a
// component, the value from the date is used.
func expandTimesOn(date time.Time, r RecurrenceRule) []time.Time {
	hours := r.ByHour
	if len(hours) == 0 {
		hours = []int{date.Hour()}
	}
	mins := r.ByMinute
	if len(mins) == 0 {
		mins = []int{date.Minute()}
	}
	secs := r.BySecond
	if len(secs) == 0 {
		secs = []int{date.Second()}
	}
	results := make([]time.Time, 0, len(hours)*len(mins)*len(secs))
	for _, h := range hours {
		for _, m := range mins {
			for _, s := range secs {
				results = append(results, time.Date(
					date.Year(), date.Month(), date.Day(),
					h, m, s, date.Nanosecond(), date.Location(),
				))
			}
		}
	}
	return results
}

func expandDaily(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		cur := start
		count := 0
		for {
			if !r.shouldContinue(cur, count, limit) {
				return
			}
			if matchesByMonth(cur, r.ByMonth) && matchesByMonthDay(cur, r.ByMonthDay) {
				for _, t := range expandTimesOn(cur, r) {
					if t.Before(start) {
						continue
					}
					if !r.shouldContinue(t, count, limit) {
						return
					}
					count++
					if !yield(t) {
						return
					}
				}
			}
			cur = cur.AddDate(0, 0, interval)
		}
	}
}

func expandHourly(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		cur := start
		count := 0
		for {
			if !r.shouldContinue(cur, count, limit) {
				return
			}
			if matchesByHour(cur, r.ByHour) && matchesByMinute(cur, r.ByMinute) && matchesBySecond(cur, r.BySecond) {
				count++
				if !yield(cur) {
					return
				}
			}
			cur = cur.Add(time.Duration(interval) * time.Hour)
		}
	}
}

func expandMinutely(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		cur := start
		count := 0
		for {
			if !r.shouldContinue(cur, count, limit) {
				return
			}
			if matchesByHour(cur, r.ByHour) && matchesByMinute(cur, r.ByMinute) && matchesBySecond(cur, r.BySecond) {
				count++
				if !yield(cur) {
					return
				}
			}
			cur = cur.Add(time.Duration(interval) * time.Minute)
		}
	}
}

func expandSecondly(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		cur := start
		count := 0
		for {
			if !r.shouldContinue(cur, count, limit) {
				return
			}
			if matchesByHour(cur, r.ByHour) && matchesByMinute(cur, r.ByMinute) && matchesBySecond(cur, r.BySecond) {
				count++
				if !yield(cur) {
					return
				}
			}
			cur = cur.Add(time.Duration(interval) * time.Second)
		}
	}
}

// applyBySetPos filters occurrences by their position within each interval's set.
// BYSETPOS=1 selects the first, BYSETPOS=-1 selects the last, etc.
// It groups occurrences by the interval boundary function and applies position
// selection per group. The interval boundary is determined by the FREQ value:
// YEARLY groups by year, MONTHLY by month, etc.
func applyBySetPos(base iter.Seq[time.Time], bySetPos []int, freq Frequency, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		var current []time.Time
		var currentKey string
		count := 0

		for t := range base {
			key := intervalKey(t, freq)
			if key != currentKey {
				count = yieldFromGroup(current, bySetPos, count, limit, yield)
				if count >= limit {
					return
				}
				current = nil
				currentKey = key
			}
			current = append(current, t)
		}
		yieldFromGroup(current, bySetPos, count, limit, yield)
	}
}

// intervalKey returns a grouping key for BYSETPOS interval boundaries based on
// the recurrence frequency. This ensures BYSETPOS selects the correct Nth
// occurrence per FREQ-defined interval.
func intervalKey(t time.Time, freq Frequency) string {
	switch freq {
	case Yearly:
		return t.Format("2006")
	case Monthly:
		return t.Format("2006-01")
	case Weekly:
		year, week := t.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case Daily:
		return t.Format("2006-01-02")
	case Hourly:
		return t.Format("2006-01-02T15")
	case Minutely:
		return t.Format("2006-01-02T15:04")
	case Secondly:
		return t.Format("2006-01-02T15:04:05")
	default:
		return t.Format("2006-01-02T15:04:05")
	}
}

func yieldFromGroup(group []time.Time, bySetPos []int, count, limit int, yield func(time.Time) bool) int {
	if len(group) == 0 {
		return count
	}
	// Sort positions for consistent selection order
	sortSet := make([]int, len(bySetPos))
	copy(sortSet, bySetPos)
	slices.Sort(sortSet)
	for _, pos := range sortSet {
		if pos > 0 {
			idx := pos - 1
			if idx >= 0 && idx < len(group) {
				count++
				if !yield(group[idx]) || count >= limit {
					return count
				}
			}
		} else {
			idx := len(group) + pos
			if idx >= 0 && idx < len(group) {
				count++
				if !yield(group[idx]) || count >= limit {
					return count
				}
			}
		}
	}
	return count
}

func matchesByHour(t time.Time, hours []int) bool {
	if len(hours) == 0 {
		return true
	}
	return slices.Contains(hours, t.Hour())
}

func matchesByMinute(t time.Time, minutes []int) bool {
	if len(minutes) == 0 {
		return true
	}
	return slices.Contains(minutes, t.Minute())
}

func matchesBySecond(t time.Time, seconds []int) bool {
	if len(seconds) == 0 {
		return true
	}
	return slices.Contains(seconds, t.Second())
}

func expandWeekly(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		weekStart := r.WeekStart
		if weekStart == "" {
			weekStart = Monday
		}
		byDay := r.ByDay
		if len(byDay) == 0 {
			byDay = []Weekday{weekdayFromTime(start.Weekday())}
		}
		week := startOfWeek(start, weekStart)
		count := 0
		for {
			if !r.shouldContinue(week, count, limit) {
				return
			}
			for _, day := range byDay {
				cur := week.AddDate(0, 0, weekdayOffset(weekStart, day))
				if cur.Before(start) {
					continue
				}
				if !matchesByMonth(cur, r.ByMonth) {
					continue
				}
				for _, t := range expandTimesOn(cur, r) {
					if t.Before(start) {
						continue
					}
					if !r.shouldContinue(t, count, limit) {
						return
					}
					count++
					if !yield(t) {
						return
					}
				}
			}
			week = week.AddDate(0, 0, 7*interval)
		}
	}
}

func expandMonthly(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	if len(r.ByMonthDay) > 0 {
		return expandMonthlyByMonthDay(start, r, interval, limit)
	}
	byDay := ordinalByDayValues(r.ByDayNum)
	if len(byDay) == 0 {
		return expandMonthlySimple(start, r, interval, limit)
	}
	return expandMonthlyByOrdinalDay(start, r, interval, limit, byDay)
}

func expandMonthlySimple(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		cur := start
		count := 0
		for {
			if !r.shouldContinue(cur, count, limit) {
				return
			}
			if matchesByMonth(cur, r.ByMonth) {
				for _, t := range expandTimesOn(cur, r) {
					if t.Before(start) {
						continue
					}
					if !r.shouldContinue(t, count, limit) {
						return
					}
					count++
					if !yield(t) {
						return
					}
				}
			}
			cur = addMonthsClamped(cur, interval, start.Day())
		}
	}
}

// addMonthsClamped adds n months to t, clamping day to lastDay when targetDay
// exceeds the month's length, per RFC 5545 month-end behavior.
func addMonthsClamped(t time.Time, n, targetDay int) time.Time {
	y, m := t.Year(), int(t.Month())
	m += n
	y += (m - 1) / 12
	m = ((m - 1) % 12) + 1

	lastDay := time.Date(y, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if targetDay > lastDay {
		targetDay = lastDay
	}
	return time.Date(y, time.Month(m), targetDay, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func expandMonthlyByOrdinalDay(start time.Time, r RecurrenceRule, interval, limit int, byDay []WeekdayNum) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		month := time.Date(start.Year(), start.Month(), 1, start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), start.Location())
		count := 0
		for {
			if !r.shouldContinue(month, count, limit) {
				return
			}
			if !matchesByMonth(month, r.ByMonth) {
				month = month.AddDate(0, interval, 0)
				continue
			}
			for _, day := range byDay {
				cur, ok := monthlyOrdinalWeekday(month, day)
				if !ok || cur.Before(start) {
					continue
				}
				for _, t := range expandTimesOn(cur, r) {
					if t.Before(start) {
						continue
					}
					if !r.shouldContinue(t, count, limit) {
						return
					}
					count++
					if !yield(t) {
						return
					}
				}
			}
			month = month.AddDate(0, interval, 0)
		}
	}
}

func expandMonthlyByMonthDay(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		month := time.Date(start.Year(), start.Month(), 1, start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), start.Location())
		count := 0
		for {
			if !r.shouldContinue(month, count, limit) {
				return
			}
			if !matchesByMonth(month, r.ByMonth) {
				month = month.AddDate(0, interval, 0)
				continue
			}
			for _, day := range r.ByMonthDay {
				cur, ok := monthlyMonthDay(month, day)
				if !ok || cur.Before(start) {
					continue
				}
				for _, t := range expandTimesOn(cur, r) {
					if t.Before(start) {
						continue
					}
					if !r.shouldContinue(t, count, limit) {
						return
					}
					count++
					if !yield(t) {
						return
					}
				}
			}
			month = month.AddDate(0, interval, 0)
		}
	}
}

func expandYearly(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	if len(r.ByWeekNo) > 0 {
		return expandYearlyByWeekNo(start, r, interval, limit)
	}
	if len(r.ByYearDay) > 0 {
		return expandYearlyByYearDay(start, r, interval, limit)
	}
	byDay := ordinalByDayValues(r.ByDayNum)
	if len(byDay) > 0 {
		return expandYearlyByOrdinalDay(start, r, interval, limit, byDay)
	}
	if len(r.ByMonth) == 0 && len(r.ByMonthDay) == 0 {
		return expandYearlySimple(start, r, interval, limit)
	}
	return expandYearlyByMonth(start, r, interval, limit)
}

func expandYearlySimple(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		cur := start
		count := 0
		for {
			if !r.shouldContinue(cur, count, limit) {
				return
			}
			for _, t := range expandTimesOn(cur, r) {
				if t.Before(start) {
					continue
				}
				if !r.shouldContinue(t, count, limit) {
					return
				}
				count++
				if !yield(t) {
					return
				}
			}
			cur = cur.AddDate(interval, 0, 0)
		}
	}
}

func expandYearlyByMonth(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		count := 0
		for year := start.Year(); ; year += interval {
			for _, month := range r.ByMonth {
				base, ok := yearlyMonth(start, year, month)
				if !ok || base.Before(start) {
					continue
				}
				for _, t := range expandTimesOn(base, r) {
					if t.Before(start) {
						continue
					}
					if !r.shouldContinue(t, count, limit) {
						return
					}
					count++
					if !yield(t) {
						return
					}
				}
			}
		}
	}
}

func expandYearlyByOrdinalDay(start time.Time, r RecurrenceRule, interval, limit int, byDay []WeekdayNum) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		count := 0
		year := start.Year()
		for {
			yearStart := time.Date(year, 1, 1, start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), start.Location())
			if !r.shouldContinue(yearStart, count, limit) {
				return
			}
			for _, day := range byDay {
				cur, ok := yearlyOrdinalWeekday(yearStart, day)
				if !ok || cur.Before(start) {
					continue
				}
				for _, t := range expandTimesOn(cur, r) {
					if t.Before(start) {
						continue
					}
					if !r.shouldContinue(t, count, limit) {
						return
					}
					count++
					if !yield(t) {
						return
					}
				}
			}
			year += interval
		}
	}
}

func expandYearlyByWeekNo(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		weekStart := r.WeekStart
		if weekStart == "" {
			weekStart = Monday
		}
		byDay := r.ByDay
		if len(byDay) == 0 {
			byDay = []Weekday{weekdayFromTime(start.Weekday())}
		}
		count := 0
		for year := start.Year(); ; year += interval {
			for _, weekNo := range r.ByWeekNo {
				week := weekOfYear(year, weekNo, weekStart)
				for _, day := range byDay {
					cur := week.AddDate(0, 0, weekdayOffset(weekStart, day))
					if cur.Before(start) {
						continue
					}
					if !matchesByMonth(cur, r.ByMonth) {
						continue
					}
					for _, t := range expandTimesOn(cur, r) {
						if t.Before(start) {
							continue
						}
						if !r.shouldContinue(t, count, limit) {
							return
						}
						count++
						if !yield(t) {
							return
						}
					}
				}
			}
		}
	}
}

func expandYearlyByYearDay(start time.Time, r RecurrenceRule, interval, limit int) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		count := 0
		for year := start.Year(); ; year += interval {
			for _, yearDay := range r.ByYearDay {
				var day time.Time
				if yearDay > 0 {
					day = time.Date(year, 1, 1, start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), start.Location())
					day = day.AddDate(0, 0, yearDay-1)
				} else {
					yearEnd := time.Date(year, 12, 31, start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), start.Location())
					day = yearEnd.AddDate(0, 0, yearDay+1)
				}
				if day.Year() != year {
					continue
				}
				if day.Before(start) {
					continue
				}
				if !matchesByMonth(day, r.ByMonth) {
					continue
				}
				for _, t := range expandTimesOn(day, r) {
					if t.Before(start) {
						continue
					}
					if !r.shouldContinue(t, count, limit) {
						return
					}
					count++
					if !yield(t) {
						return
					}
				}
			}
		}
	}
}

// weekOfYear returns the start of week weekNo in the given year, using weekStart
// as the first day of the week. Week numbering follows RFC 5545 (ISO week):
// week 1 is the first week containing at least 4 days in the new year.
func weekOfYear(year, weekNo int, weekStart Weekday) time.Time {
	// Week 1 is the week containing Jan 4 (at least 4 days in the new year).
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	week1Start := startOfWeek(jan4, weekStart)

	if weekNo > 0 {
		return week1Start.AddDate(0, 0, 7*(weekNo-1))
	}
	// Negative: count from end of year
	dec28 := time.Date(year, 12, 28, 0, 0, 0, 0, time.UTC)
	lastWeekStart := startOfWeek(dec28, weekStart)
	return lastWeekStart.AddDate(0, 0, 7*(weekNo+1))
}

// yearlyOrdinalWeekday returns the Nth weekday of the given year.
// Positive ordinals count from the start of the year; negative from the end.
func yearlyOrdinalWeekday(yearStart time.Time, day WeekdayNum) (time.Time, bool) {
	if day.Ordinal > 0 {
		offset := weekdayOffset(weekdayFromTime(yearStart.Weekday()), day.Weekday)
		cur := yearStart.AddDate(0, 0, offset+7*(day.Ordinal-1))
		if cur.Year() != yearStart.Year() {
			return time.Time{}, false
		}
		return cur, true
	}
	yearEnd := time.Date(yearStart.Year(), 12, 31, yearStart.Hour(), yearStart.Minute(), yearStart.Second(), yearStart.Nanosecond(), yearStart.Location())
	offset := weekdayOffset(day.Weekday, weekdayFromTime(yearEnd.Weekday()))
	cur := yearEnd.AddDate(0, 0, -offset+7*(day.Ordinal+1))
	if cur.Year() != yearStart.Year() {
		return time.Time{}, false
	}
	return cur, true
}

func matchesByMonth(t time.Time, months []int) bool {
	if len(months) == 0 {
		return true
	}
	return slices.Contains(months, int(t.Month()))
}

func matchesByMonthDay(t time.Time, days []int) bool {
	if len(days) == 0 {
		return true
	}
	last := time.Date(t.Year(), t.Month()+1, 0, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()).Day()
	for _, day := range days {
		if day < 0 {
			day = last + day + 1
		}
		if t.Day() == day {
			return true
		}
	}
	return false
}

// yearlyMonth returns the time for the given year and month with the same day and time as start, or false if that day is not valid for the month.
func yearlyMonth(start time.Time, year, month int) (time.Time, bool) {
	cur := time.Date(year, time.Month(month), start.Day(), start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), start.Location())
	if cur.Month() != time.Month(month) {
		return time.Time{}, false
	}
	return cur, true
}

func monthlyMonthDay(month time.Time, day int) (time.Time, bool) {
	last := month.AddDate(0, 1, -1)
	if day < 0 {
		day = last.Day() + day + 1
	}
	if day < 1 || day > last.Day() {
		return time.Time{}, false
	}
	return time.Date(month.Year(), month.Month(), day, month.Hour(), month.Minute(), month.Second(), month.Nanosecond(), month.Location()), true
}

func ordinalByDayValues(values []WeekdayNum) []WeekdayNum {
	var out []WeekdayNum
	for _, value := range values {
		if value.Ordinal != 0 {
			out = append(out, value)
		}
	}
	return out
}

func monthlyOrdinalWeekday(month time.Time, day WeekdayNum) (time.Time, bool) {
	if day.Ordinal > 0 {
		first := month
		offset := weekdayOffset(weekdayFromTime(first.Weekday()), day.Weekday)
		cur := first.AddDate(0, 0, offset+7*(day.Ordinal-1))
		if cur.Month() != month.Month() {
			return time.Time{}, false
		}
		return cur, true
	}
	last := month.AddDate(0, 1, -1)
	offset := weekdayOffset(day.Weekday, weekdayFromTime(last.Weekday()))
	cur := last.AddDate(0, 0, -offset+7*(day.Ordinal+1))
	if cur.Month() != month.Month() {
		return time.Time{}, false
	}
	return cur, true
}

func (r RecurrenceRule) shouldContinue(cur time.Time, count, limit int) bool {
	if count >= limit {
		return false
	}
	if r.Until == nil {
		return true
	}
	if r.UntilIsDate {
		// DATE comparison: cur's calendar date must not exceed UNTIL's date.
		curYear, curMonth, curDay := cur.Date()
		untilYear, untilMonth, untilDay := r.Until.Date()
		if curYear > untilYear {
			return false
		}
		if curYear == untilYear {
			if curMonth > untilMonth {
				return false
			}
			if curMonth == untilMonth && curDay > untilDay {
				return false
			}
		}
		return true
	}
	return !cur.After(*r.Until)
}

func startOfWeek(t time.Time, start Weekday) time.Time {
	offset := weekdayOffset(start, weekdayFromTime(t.Weekday()))
	return t.AddDate(0, 0, -offset)
}

func weekdayOffset(start, day Weekday) int {
	startIndex := weekdayIndex(start)
	dayIndex := weekdayIndex(day)
	if dayIndex < startIndex {
		dayIndex += 7
	}
	return dayIndex - startIndex
}

func weekdayIndex(day Weekday) int {
	switch day {
	case Monday:
		return 0
	case Tuesday:
		return 1
	case Wednesday:
		return 2
	case Thursday:
		return 3
	case Friday:
		return 4
	case Saturday:
		return 5
	case Sunday:
		return 6
	default:
		return 0
	}
}

func weekdayFromTime(day time.Weekday) Weekday {
	switch day {
	case time.Monday:
		return Monday
	case time.Tuesday:
		return Tuesday
	case time.Wednesday:
		return Wednesday
	case time.Thursday:
		return Thursday
	case time.Friday:
		return Friday
	case time.Saturday:
		return Saturday
	case time.Sunday:
		return Sunday
	default:
		return Monday
	}
}

func joinInts(values []int) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.Itoa(value)
	}
	return strings.Join(parts, ",")
}

func parseIntList(values []string) ([]int, error) {
	result := make([]int, len(values))
	for i, value := range values {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		result[i] = parsed
	}
	return result, nil
}

func validateIntRange(name string, values []int, min, max int) error {
	for _, value := range values {
		if value < min || value > max {
			return fmt.Errorf("ical: RRULE %s value %d out of range %d..%d", name, value, min, max)
		}
	}
	return nil
}

func validateIntRangeNoZero(name string, values []int, min, max int) error {
	for _, value := range values {
		if value == 0 || value < min || value > max {
			return fmt.Errorf("ical: RRULE %s value %d out of range %d..%d excluding 0", name, value, min, max)
		}
	}
	return nil
}
