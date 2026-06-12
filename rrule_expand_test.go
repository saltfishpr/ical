package ical_test

import (
	"strings"
	"testing"
	"time"

	"github.com/saltfishpr/ical"
)

// ---- BYHOUR/BYMINUTE/BYSECOND on DAILY ----

func TestExpandDailyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;BYHOUR=9,17;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := utcTime(2026, 6, 12, 10, 0, 0)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		utcTime(2026, 6, 12, 17, 0, 0),
		utcTime(2026, 6, 13, 9, 0, 0),
		utcTime(2026, 6, 13, 17, 0, 0),
		utcTime(2026, 6, 14, 9, 0, 0),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandDailyWithByHourAndMinute(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=DAILY;BYHOUR=9;BYMINUTE=0,30;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := utcTime(2026, 6, 12, 9, 0, 0)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		utcTime(2026, 6, 12, 9, 0, 0),
		utcTime(2026, 6, 12, 9, 30, 0),
		utcTime(2026, 6, 13, 9, 0, 0),
		utcTime(2026, 6, 13, 9, 30, 0),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYHOUR on WEEKLY ----

func TestExpandWeeklyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=WEEKLY;BYDAY=MO;BYHOUR=10,14;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := utcTime(2026, 6, 1, 9, 0, 0) // Monday
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		utcTime(2026, 6, 1, 10, 0, 0),
		utcTime(2026, 6, 1, 14, 0, 0),
		utcTime(2026, 6, 8, 10, 0, 0),
		utcTime(2026, 6, 8, 14, 0, 0),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYHOUR on MONTHLY ----

func TestExpandMonthlyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYHOUR=9,17;COUNT=4")
	if err != nil {
		t.Fatal(err)
	}
	start := utcTime(2026, 6, 15, 10, 0, 0)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		utcTime(2026, 6, 15, 17, 0, 0),
		utcTime(2026, 7, 15, 9, 0, 0),
		utcTime(2026, 7, 15, 17, 0, 0),
		utcTime(2026, 8, 15, 9, 0, 0),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYHOUR on YEARLY ----

func TestExpandYearlyWithByHour(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYHOUR=9,17;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := utcTime(2026, 6, 15, 10, 0, 0)
	results := collect(rule, start)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3: %v", len(results), results)
	}
	expected := []time.Time{
		utcTime(2026, 6, 15, 17, 0, 0),
		utcTime(2027, 6, 15, 9, 0, 0),
		utcTime(2027, 6, 15, 17, 0, 0),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYYEARDAY ----

func TestExpandYearlyByYearDay(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYYEARDAY=1,100,200;COUNT=5")
	if err != nil {
		t.Fatal(err)
	}
	start := utcTime(2026, 1, 1, 10, 0, 0)
	results := collect(rule, start)
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5: %v", len(results), results)
	}
	expected := []time.Time{
		utcTime(2026, 1, 1, 10, 0, 0),
		utcTime(2026, 4, 10, 10, 0, 0),
		utcTime(2026, 7, 19, 10, 0, 0),
		utcTime(2027, 1, 1, 10, 0, 0),
		utcTime(2027, 4, 10, 10, 0, 0),
	}
	assertDatesEqual(t, results, expected)
}

func TestExpandYearlyByYearDayNegative(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYYEARDAY=-1;COUNT=2")
	if err != nil {
		t.Fatal(err)
	}
	start := utcTime(2025, 12, 31, 10, 0, 0)
	results := collect(rule, start)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2: %v", len(results), results)
	}
	expected := []time.Time{
		utcTime(2025, 12, 31, 10, 0, 0),
		utcTime(2026, 12, 31, 10, 0, 0),
	}
	assertDatesEqual(t, results, expected)
}

// ---- BYWEEKNO ----

func TestExpandYearlyByWeekNo(t *testing.T) {
	rule, err := ical.ParseRecurrenceRule("FREQ=YEARLY;BYWEEKNO=20;BYDAY=MO;COUNT=3")
	if err != nil {
		t.Fatal(err)
	}
	start := utcTime(2026, 5, 11, 10, 0, 0)
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
	start := utcTime(2026, 1, 5, 10, 0, 0) // Monday in week 1 of 2026
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
	start := utcTime(1996, 4, 1, 10, 0, 0)
	results := collect(rule, start)
	if len(results) != 10 {
		t.Fatalf("got %d results, want 10: %v", len(results), results)
	}
	want := utcTime(1996, 4, 10, 10, 0, 0)
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
	start := utcTime(2026, 1, 1, 10, 0, 0)
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
	start := utcTime(2026, 6, 12, 10, 0, 0)
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
	start := utcTime(2026, 1, 4, 9, 0, 0)
	results := collect(rule, start)
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %v", len(results), results)
	}
	expected := []time.Time{
		utcTime(2026, 1, 4, 9, 0, 0),
		utcTime(2026, 1, 4, 17, 0, 0),
		utcTime(2027, 1, 3, 9, 0, 0),
		utcTime(2027, 1, 3, 17, 0, 0),
	}
	assertDatesEqual(t, results, expected)
}
