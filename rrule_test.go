package ical_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/saltfishpr/ical"
)

// collect expands the rule from start and returns all results as a slice.
func collect(rule ical.RecurrenceRule, start time.Time) []time.Time {
	var results []time.Time
	for t := range rule.Expand(start) {
		results = append(results, t)
	}
	return results
}

// assertDatesEqual is a test helper that compares two time slices by their date/time equality.
func assertDatesEqual(t *testing.T, got, want []time.Time) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d results, want %d: got=%v, want=%v", len(got), len(want), got, want)
	}
	for i := range got {
		if !got[i].Equal(want[i]) {
			t.Fatalf("result[%d]: got %v, want %v", i, got[i], want[i])
		}
	}
}

// ---- Daily ----

func TestExpandDailyCount(t *testing.T) {
	// FREQ=DAILY;COUNT=100 (from vendor events/event_with_recurrence.ics)
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;COUNT=10")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(1996, time.Month(4), 1, 1, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 10 {
		t.Fatalf("got %d results, want 10", len(results))
	}
	want := time.Date(1996, time.Month(4), 10, 1, 0, 0, 0, time.UTC)
	if !results[9].Equal(want) {
		t.Fatalf("last result = %s, want %s", results[9], want)
	}
}

func TestExpandDailyUntil(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;UNTIL=19960410T010000Z")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(1996, time.Month(4), 1, 1, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 10 {
		t.Fatalf("got %d results, want 10", len(results))
	}
	want := time.Date(1996, time.Month(4), 10, 1, 0, 0, 0, time.UTC)
	if !results[9].Equal(want) {
		t.Fatalf("last result = %s, want %s", results[9], want)
	}
}

func TestExpandDailyInterval(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;INTERVAL=2;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 1, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5", len(results))
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 3, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 5, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 7, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 9, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandDailyWithByMonth(t *testing.T) {
	// Only recur in January and March
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;BYMONTH=1,3;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 30, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5", len(results))
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 30, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(3), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(3), 2, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(3), 3, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandDailyWithByMonthDay(t *testing.T) {
	// Only recur on the 15th and -1 (last day) of each month
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;BYMONTHDAY=15,-1;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	// 2026-01 has 31 days, so -1 => 31
	expected := []time.Time{
		time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(2), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(2), 28, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- Weekly ----

func TestExpandWeeklyImplicitByDay(t *testing.T) {
	// FREQ=WEEKLY without BYDAY uses DTSTART's weekday (from vendor events/issue_70_rrule_causes_attribute_error.ics)
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;INTERVAL=1;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	// 2007-02-20 is a Tuesday
	start := time.Date(2007, time.Month(2), 20, 17, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4", len(results))
	}
	for _, r := range results {
		if r.Weekday() != time.Tuesday {
			t.Fatalf("expected Tuesday, got %s at %s", r.Weekday(), r)
		}
	}
}

func TestExpandWeeklyByDay(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;BYDAY=MO,WE,FR;COUNT=6")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 1, 9, 0, 0, 0, time.UTC) // Monday
	results := collect(rule, start)
	if len(results) != 6 {
		t.Fatalf("got %d results, want 6", len(results))
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 1, 9, 0, 0, 0, time.UTC),  // Mon
		time.Date(2026, time.Month(6), 3, 9, 0, 0, 0, time.UTC),  // Wed
		time.Date(2026, time.Month(6), 5, 9, 0, 0, 0, time.UTC),  // Fri
		time.Date(2026, time.Month(6), 8, 9, 0, 0, 0, time.UTC),  // Mon
		time.Date(2026, time.Month(6), 10, 9, 0, 0, 0, time.UTC), // Wed
		time.Date(2026, time.Month(6), 12, 9, 0, 0, 0, time.UTC), // Fri
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandWeeklyByDaySkipsBeforeStart(t *testing.T) {
	// Start on Wednesday; MO in the same week is before start and should be skipped
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;BYDAY=MO,WE,FR;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 3, 9, 0, 0, 0, time.UTC) // Wednesday
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5: %v", len(results), results)
	}
	// First result should be Wednesday (not Monday which is before start)
	if !results[0].Equal(start) {
		t.Fatalf("first result = %s, want %s", results[0], start)
	}
	if results[0].Weekday() != time.Wednesday {
		t.Fatalf("expected Wednesday, got %s", results[0].Weekday())
	}
}

func TestExpandWeeklyInterval(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;INTERVAL=2;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 1, 10, 0, 0, 0, time.UTC) // Monday
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 29, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandWeeklyWeekStart(t *testing.T) {
	// WKST=SU means weeks start on Sunday; BYDAY=SA falls on the right offset
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;BYDAY=MO;WKST=SU;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	// 2026-06-07 is a Sunday; with WKST=SU, Monday is the next day
	start := time.Date(2026, time.Month(6), 7, 10, 0, 0, 0, time.UTC) // Sunday
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4", len(results))
	}
	// When WKST=SU, week starts on Sunday. BYDAY=MO looks for Monday within the week.
	// Since start is Sunday, the first Monday in that week is 2026-06-08.
	if !results[0].Equal(time.Date(2026, time.Month(6), 8, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("first result = %s, want 2026-06-08 (Monday in Sunday-start week)", results[0])
	}
}

func TestExpandWeeklyWithByMonth(t *testing.T) {
	// Only expand in January and February
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;BYDAY=MO;BYMONTH=1,2;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 5, 10, 0, 0, 0, time.UTC) // Monday in January
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	for _, r := range results {
		if !slices.Contains([]time.Month{time.January, time.February}, r.Month()) {
			t.Fatalf("result %s is outside Jan/Feb", r)
		}
	}
}

func TestExpandWeeklyUntil(t *testing.T) {
	// FREQ=WEEKLY;UNTIL=20070619T225959 (from vendor events/issue_70_rrule_causes_attribute_error.ics)
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;INTERVAL=1;UNTIL=20070619T225959Z")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2007, time.Month(2), 20, 17, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	// 2007-02-20 to 2007-06-19 is 17 weeks of Tuesdays = 18 occurrences
	if len(results) != 18 {
		t.Fatalf("got %d results, want 18", len(results))
	}
	if results[len(results)-1].After(time.Date(2007, time.Month(6), 19, 22, 59, 59, 0, time.UTC)) {
		t.Fatal("last result is after UNTIL")
	}
}

// ---- Monthly ----

func TestExpandMonthlySimple(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4", len(results))
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(2), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(3), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(4), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyInterval(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;INTERVAL=2;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(3), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(5), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyByMonthDay(t *testing.T) {
	// 15th and last day of each month
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYMONTHDAY=15,-1;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(2), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(2), 28, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyByMonthDaySkipsInvalid(t *testing.T) {
	// BYMONTHDAY=31 skips months with fewer than 31 days
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYMONTHDAY=31;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 31, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	// Jan (31), Mar (31), May (31) — Feb and Apr are skipped
	expected := []time.Time{
		time.Date(2026, time.Month(1), 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(3), 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(5), 31, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyByOrdinalDayPositive(t *testing.T) {
	// 2nd Thursday of each month (from vendor test_recurrence.py: "MONTHLY", "2TH")
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYDAY=2TH;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2003, time.Month(4), 10, 0, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	// April 2003: 2nd Thu = 10, May 2003: 2nd Thu = 8, June 2003: 2nd Thu = 11
	expected := []time.Time{
		time.Date(2003, time.Month(4), 10, 0, 0, 0, 0, time.UTC),
		time.Date(2003, time.Month(5), 8, 0, 0, 0, 0, time.UTC),
		time.Date(2003, time.Month(6), 12, 0, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyByOrdinalDayNegative(t *testing.T) {
	// 3rd-to-last Friday of each month (from vendor test_recurrence.py: "MONTHLY", "-3FR")
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYDAY=-3FR;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2017, time.Month(5), 12, 0, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	// May 2017: -3FR = 12 (31 days: 31=Wed, 30=Tue, 29=Mon, 28=Sun, 27=Sat, 26=Fri(-1), 19=Fri(-2), 12=Fri(-3))
	// June 2017: -3FR = 16 (30 days)
	// July 2017: -3FR = 14 (31 days)
	expected := []time.Time{
		time.Date(2017, time.Month(5), 12, 0, 0, 0, 0, time.UTC),
		time.Date(2017, time.Month(6), 16, 0, 0, 0, 0, time.UTC),
		time.Date(2017, time.Month(7), 14, 0, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyByOrdinalDayMultiple(t *testing.T) {
	// 1st Monday and last Sunday of each month
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYDAY=1MO,-1SU;COUNT=6")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 1, 10, 0, 0, 0, time.UTC) // This is a Monday
	results := collect(rule, start)
	if len(results) != 6 {
		t.Fatalf("got %d results, want 6: %v", len(results), results)
	}
	// June 2026: 1MO=1, -1SU=28
	// July 2026: 1MO=6, -1SU=26
	// Aug 2026: 1MO=3, -1SU=30
	expected := []time.Time{
		time.Date(2026, time.Month(6), 1, 10, 0, 0, 0, time.UTC),  // 1st Monday June
		time.Date(2026, time.Month(6), 28, 10, 0, 0, 0, time.UTC), // last Sunday June
		time.Date(2026, time.Month(7), 6, 10, 0, 0, 0, time.UTC),  // 1st Monday July
		time.Date(2026, time.Month(7), 26, 10, 0, 0, 0, time.UTC), // last Sunday July
		time.Date(2026, time.Month(8), 3, 10, 0, 0, 0, time.UTC),  // 1st Monday August
		time.Date(2026, time.Month(8), 30, 10, 0, 0, 0, time.UTC), // last Sunday August
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyByOrdinalDaySkipsBeforeStart(t *testing.T) {
	// 1st Monday, but start is after the 1st Monday of that month
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYDAY=1MO;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	// June 2026: 1st Monday is June 1. We start on June 5, so the June 1st Monday is skipped.
	start := time.Date(2026, time.Month(6), 5, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(7), 6, 10, 0, 0, 0, time.UTC), // 1st Monday July
		time.Date(2026, time.Month(8), 3, 10, 0, 0, 0, time.UTC), // 1st Monday August
		time.Date(2026, time.Month(9), 7, 10, 0, 0, 0, time.UTC), // 1st Monday September
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyByOrdinalDayLeapYear(t *testing.T) {
	// -1SU (last Sunday) in February of a leap year
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYDAY=-1SU;COUNT=2")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2024, time.Month(2), 25, 10, 0, 0, 0, time.UTC) // Last Sunday of Feb 2024
	results := collect(rule, start)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2: %v", len(results), results)
	}
	// Feb 2024: 29 days, last Sunday = 25
	// Mar 2024: 31 days, last Sunday = 31
	expected := []time.Time{
		time.Date(2024, time.Month(2), 25, 10, 0, 0, 0, time.UTC),
		time.Date(2024, time.Month(3), 31, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyWithByMonth(t *testing.T) {
	// Monthly but only in specific months
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYMONTH=3,6,9;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(3), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(3), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(9), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- Yearly ----

func TestExpandYearlySimple(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2020, time.Month(6), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	expected := []time.Time{
		time.Date(2020, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2021, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2022, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyInterval(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;INTERVAL=2;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2020, time.Month(6), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	expected := []time.Time{
		time.Date(2020, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2022, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2024, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyByMonth(t *testing.T) {
	// Yearly in January and July
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYMONTH=1,7;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(7), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(1), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(7), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyByMonthSkipsBeforeStart(t *testing.T) {
	// BYMONTH=3,6; start in May should skip March
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYMONTH=3,6;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(5), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(3), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyByMonthSkipsInvalidDay(t *testing.T) {
	// BYMONTH=2 on a leap year for day 29
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYMONTH=2;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2024, time.Month(2), 29, 10, 0, 0, 0, time.UTC) // Leap year
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2024, time.Month(2), 29, 10, 0, 0, 0, time.UTC),
		time.Date(2028, time.Month(2), 29, 10, 0, 0, 0, time.UTC),
		time.Date(2032, time.Month(2), 29, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyByMonthUntil(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYMONTH=1,7;UNTIL=20280101T000000Z")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	// Should stop before Jan 2028 (Jan and Jul 2026, Jan and Jul 2027)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
}

// ---- Edge cases ----

func TestExpandCountOne(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;COUNT=1")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if !results[0].Equal(start) {
		t.Fatalf("result = %s, want %s", results[0], start)
	}
}

func TestExpandStartEqualsUntil(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;UNTIL=20260612T100000Z")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	// start == UNTIL, should yield the first occurrence
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
}

func TestExpandStartAfterUntil(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;UNTIL=20260601T000000Z")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}

func TestExpandSecondly(t *testing.T) {
	rule := ical.RecurrenceRule{Frequency: ical.Secondly, Count: 10}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 10 {
		t.Fatalf("got %d results, want 10", len(results))
	}
	// Each result should be 1 second apart
	for i := 1; i < len(results); i++ {
		diff := results[i].Sub(results[i-1])
		if diff != time.Second {
			t.Fatalf("expected 1s between occurrences, got %v at index %d", diff, i)
		}
	}
}

func TestExpandDailyLeapYear(t *testing.T) {
	// Daily across a leap year boundary
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2024, time.Month(2), 28, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5", len(results))
	}
	expected := []time.Time{
		time.Date(2024, time.Month(2), 28, 10, 0, 0, 0, time.UTC),
		time.Date(2024, time.Month(2), 29, 10, 0, 0, 0, time.UTC), // leap day
		time.Date(2024, time.Month(3), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2024, time.Month(3), 2, 10, 0, 0, 0, time.UTC),
		time.Date(2024, time.Month(3), 3, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandWeeklyEndOfYear(t *testing.T) {
	// Weekly across year boundary
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2025, time.Month(12), 29, 10, 0, 0, 0, time.UTC) // Monday
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2025, time.Month(12), 29, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 5, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 12, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 19, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlySimpleEndOfMonth(t *testing.T) {
	// DTSTART=31, monthly should preserve day-of-month where possible
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 31, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(2), 28, 10, 0, 0, 0, time.UTC), // truncated to last day of Feb
		time.Date(2026, time.Month(3), 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(4), 30, 10, 0, 0, 0, time.UTC), // truncated to last day of Apr
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandMonthlyByDayAndByMonthDay(t *testing.T) {
	// When both BYDAY and BYMONTHDAY are present, BYMONTHDAY takes precedence
	// (BYDAY without ordinal is ignored in monthly expansion)
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYMONTHDAY=1,15;BYDAY=MO;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 1, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(7), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(7), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyByMonthWithInterval(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;INTERVAL=2;BYMONTH=6;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2020, time.Month(6), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2020, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2022, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2024, time.Month(6), 15, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- Hourly ----

func TestExpandHourly(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=HOURLY;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5", len(results))
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 11, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 12, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 13, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 14, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandHourlyInterval(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=HOURLY;INTERVAL=2;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 12, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 14, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandHourlyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=HOURLY;BYHOUR=9,17;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 9, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 12, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 17, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 13, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 13, 17, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- Minutely ----

func TestExpandMinutely(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=MINUTELY;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5", len(results))
	}
	// Every minute
	for i := range results {
		want := start.Add(time.Duration(i) * time.Minute)
		if !results[i].Equal(want) {
			t.Fatalf("results[%d] = %s, want %s", i, results[i], want)
		}
	}
}

func TestExpandMinutelyWithByMinute(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=MINUTELY;BYMINUTE=0,30;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 10, 30, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 11, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 11, 30, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- Secondly with BY ----

func TestExpandSecondlyWithBySecond(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=SECONDLY;BYSECOND=0,30;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 10, 0, 30, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 10, 1, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 10, 1, 30, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 10, 2, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- Yearly BYDAY ----

func TestExpandYearlyByDayFirstSunday(t *testing.T) {
	// 1st Sunday of each year
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYDAY=1SU;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2016, time.Month(1), 3, 0, 0, 0, 0, time.UTC) // 1st Sunday of 2016
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2016, time.Month(1), 3, 0, 0, 0, 0, time.UTC),
		time.Date(2017, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
		time.Date(2018, time.Month(1), 7, 0, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyByDay53rdMonday(t *testing.T) {
	// 53rd Monday in leap year
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYDAY=53MO;COUNT=1")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(1984, time.Month(12), 31, 0, 0, 0, 0, time.UTC) // 53rd Monday of 1984
	results := collect(rule, start)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Month() != 12 || results[0].Day() != 31 || results[0].Year() != 1984 {
		t.Fatalf("expected 1984-12-31, got %s", results[0])
	}
}

func TestExpandYearlyByDayLastTuesday(t *testing.T) {
	// Last Tuesday of each year
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYDAY=-1TU;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(1999, time.Month(12), 28, 0, 0, 0, 0, time.UTC) // Last Tuesday of 1999
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	for _, r := range results {
		if r.Weekday() != time.Tuesday {
			t.Fatalf("expected Tuesday, got %s at %s", r.Weekday(), r)
		}
	}
}

func TestExpandYearlyByDay17thToLastWednesday(t *testing.T) {
	// 17th-to-last Wednesday
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYDAY=-17WE;COUNT=1")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2000, time.Month(9), 6, 0, 0, 0, 0, time.UTC) // 17th to last Wed of 2000
	results := collect(rule, start)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Weekday() != time.Wednesday {
		t.Fatalf("expected Wednesday, got %s", results[0].Weekday())
	}
}

func TestExpandYearlyByDay9thMonday(t *testing.T) {
	// 9th Monday in year (from issue #518)
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYDAY=9MO;COUNT=2")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2023, time.Month(2), 27, 0, 0, 0, 0, time.UTC) // 9th Monday of 2023
	results := collect(rule, start)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2: %v", len(results), results)
	}
	for _, r := range results {
		if r.Weekday() != time.Monday {
			t.Fatalf("expected Monday, got %s at %s", r.Weekday(), r)
		}
	}
}

// ---- BYSETPOS ----

func TestExpandMonthlyBySetPos(t *testing.T) {
	// BYSETPOS=-1 selects the last occurrence in each month
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYMONTHDAY=1,15,30;BYSETPOS=-1;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 1, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	// January: 30, February: 15 (no 30 in Feb), March: 30
	expected := []time.Time{
		time.Date(2026, time.Month(1), 30, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(2), 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(3), 30, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYHOUR/BYMINUTE/BYSECOND on DAILY ----

func TestExpandDailyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;BYHOUR=9,17;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 12, 17, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 13, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 13, 17, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 14, 9, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandDailyWithByHourAndMinute(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;BYHOUR=9;BYMINUTE=0,30;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 9, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 12, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 12, 9, 30, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 13, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 13, 9, 30, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYHOUR on WEEKLY ----

func TestExpandWeeklyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;BYDAY=MO;BYHOUR=10,14;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 1, 9, 0, 0, 0, time.UTC) // Monday
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 1, 14, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 8, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(6), 8, 14, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYHOUR on MONTHLY ----

func TestExpandMonthlyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYHOUR=9,17;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 15, 17, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(7), 15, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(7), 15, 17, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(8), 15, 9, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYHOUR on YEARLY ----

func TestExpandYearlyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYHOUR=9,17;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 15, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(6), 15, 17, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(6), 15, 9, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(6), 15, 17, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYYEARDAY ----

func TestExpandYearlyByYearDay(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYYEARDAY=1,100,200;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 1, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(4), 10, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(7), 19, 10, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(1), 1, 10, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(4), 10, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyByYearDayNegative(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYYEARDAY=-1;COUNT=2")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2025, time.Month(12), 31, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2025, time.Month(12), 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(12), 31, 10, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYWEEKNO ----

func TestExpandYearlyByWeekNo(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYWEEKNO=20;BYDAY=MO;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(5), 11, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	for _, r := range results {
		if r.Weekday() != time.Monday {
			t.Fatalf("expected Monday, got %s at %s", r.Weekday(), r)
		}
	}
}

func TestExpandYearlyByWeekNoImplicitDay(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYWEEKNO=1;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 5, 10, 0, 0, 0, time.UTC) // Monday in week 1 of 2026
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	for _, r := range results {
		if r.Weekday() != time.Monday {
			t.Fatalf("expected Monday, got %s at %s", r.Weekday(), r)
		}
	}
}

// ---- UNTIL with DATE type ----

func TestExpandDailyUntilDate(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;UNTIL=19960410")
	if err != nil {
		t.Fatal(err)
	}
	if !rule.UntilIsDate {
		t.Fatal("expected UntilIsDate to be true for DATE-typed UNTIL")
	}
	start := time.Date(1996, time.Month(4), 1, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 10 {
		t.Fatalf("got %d results, want 10: %v", len(results), results)
	}
	want := time.Date(1996, time.Month(4), 10, 10, 0, 0, 0, time.UTC)
	if !results[9].Equal(want) {
		t.Fatalf("last result = %s, want %s", results[9], want)
	}
}

func TestParseRRuleUntilDate(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;UNTIL=20260115;BYMONTH=1")
	if err != nil {
		t.Fatal(err)
	}
	if !rule.UntilIsDate {
		t.Fatal("expected UntilIsDate=true for DATE UNTIL")
	}
	s := rule.String()
	if !strings.Contains(s, "UNTIL=20260115") || strings.Contains(s, "UNTIL=20260115T") {
		t.Fatalf("expected DATE-format UNTIL, got %q", s)
	}
}

// ---- BYSETPOS with different frequencies ----

func TestExpandYearlyBySetPos(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYMONTH=1,6;BYMONTHDAY=1,15;BYSETPOS=1;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 1, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	for i, r := range results {
		if r.Day() != 1 || r.Month() != time.January {
			t.Fatalf("result[%d] expected Jan 1, got %s", i, r)
		}
	}
}

func TestExpandDailyBySetPos(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;BYHOUR=9,12,17;BYSETPOS=-1;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(6), 12, 10, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	for _, r := range results {
		if r.Hour() != 17 {
			t.Fatalf("expected hour 17 (last per day), got %s", r)
		}
	}
}

// ---- Ordinal BYDAY on YEARLY with BYHOUR ----

func TestExpandYearlyByDayWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYDAY=1SU;BYHOUR=9,17;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.Month(1), 4, 9, 0, 0, 0, time.UTC)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		time.Date(2026, time.Month(1), 4, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.Month(1), 4, 17, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(1), 3, 9, 0, 0, 0, time.UTC),
		time.Date(2027, time.Month(1), 3, 17, 0, 0, 0, time.UTC),
	}
	assertDatesEqual(t, results, expected)
}
