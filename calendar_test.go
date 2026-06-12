package ical_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saltfishpr/ical"
)

func TestRoundTripCalendarExample(t *testing.T) {
	data := readTestdata(t, "calendars/example.ics")
	cal, err := ical.ParseCalendar(data)
	if err != nil {
		t.Fatal(err)
	}
	if cal.Name != ical.CompVCalendar {
		t.Fatalf("expected VCALENDAR, got %s", cal.Name)
	}
	events := cal.Events()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	// Round-trip: serialize and re-parse
	reparsed, err := ical.ParseCalendar([]byte(cal.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(reparsed.Events()) != len(events) {
		t.Fatalf("round-trip event count mismatch: %d vs %d", len(reparsed.Events()), len(events))
	}
}

func TestParseEventWithRecurrence(t *testing.T) {
	data := readTestdata(t, "events/event_with_recurrence.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	if event.Name != ical.CompVEvent {
		t.Fatalf("expected VEVENT, got %s", event.Name)
	}
	rule, ok := event.RecurrenceRule(ical.PropRRule)
	if !ok {
		t.Fatal("RRULE not found")
	}
	if rule.Frequency != ical.Daily {
		t.Fatalf("expected DAILY, got %s", rule.Frequency)
	}
	if rule.Count != 100 {
		t.Fatalf("expected COUNT=100, got %d", rule.Count)
	}
	// Check EXDATE
	dates, ok := event.DateTimes(ical.PropExDate)
	if !ok {
		t.Fatal("EXDATE not found")
	}
	if len(dates) != 3 {
		t.Fatalf("expected 3 EXDATE values, got %d", len(dates))
	}
	// Round-trip
	reparsed, err := ical.ParseComponent([]byte(event.String()))
	if err != nil {
		t.Fatal(err)
	}
	dates2, ok := reparsed.DateTimes(ical.PropExDate)
	if !ok || len(dates2) != 3 {
		t.Fatal("EXDATE round-trip failed")
	}
}

func TestParseEventWithExdatesOnMultipleLines(t *testing.T) {
	data := readTestdata(t, "events/event_with_recurrence_exdates_on_different_lines.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	exdates := event.PropertiesByName(ical.PropExDate)
	if len(exdates) != 5 {
		t.Fatalf("expected 5 EXDATE properties, got %d", len(exdates))
	}
	// Verify TZID parameter is preserved on each EXDATE
	for _, exd := range exdates {
		if tzid := exd.Params.Get(ical.ParamTZID); tzid != "Europe/Vienna" {
			t.Fatalf("expected TZID=Europe/Vienna, got %q", tzid)
		}
	}
}

func TestParseEventWithPeriod(t *testing.T) {
	data := readTestdata(t, "events/issue_156_RDATE_with_PERIOD.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	periods, ok := event.Periods(ical.PropRDate)
	if !ok {
		t.Fatal("RDATE PERIOD not found")
	}
	if len(periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(periods))
	}
	// Round-trip
	reparsed, err := ical.ParseComponent([]byte(event.String()))
	if err != nil {
		t.Fatal(err)
	}
	periods2, ok := reparsed.Periods(ical.PropRDate)
	if !ok || len(periods2) != 1 {
		t.Fatal("PERIOD round-trip failed")
	}
}

func TestParseEventWithPeriodList(t *testing.T) {
	data := readTestdata(t, "events/issue_156_RDATE_with_PERIOD_list.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	periods, ok := event.Periods(ical.PropRDate)
	if !ok {
		t.Fatal("RDATE PERIOD not found")
	}
	if len(periods) != 2 {
		t.Fatalf("expected 2 periods, got %d", len(periods))
	}
	// Second period should have duration PT5H30M
	if periods[1].Duration == 0 {
		t.Fatal("second period should have duration")
	}
	// Round-trip serialization
	serialized := event.String()
	reparsed, err := ical.ParseComponent([]byte(serialized))
	if err != nil {
		t.Fatalf("re-parse failed: %v\nserialized: %s", err, serialized)
	}
	periods2, ok := reparsed.Periods(ical.PropRDate)
	if !ok || len(periods2) != 2 {
		t.Fatalf("PERIOD list round-trip failed: got %d periods", len(periods2))
	}
}

func TestParseRRuleWeeklyUntil(t *testing.T) {
	data := readTestdata(t, "events/issue_70_rrule_causes_attribute_error.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	rule, ok := event.RecurrenceRule(ical.PropRRule)
	if !ok {
		t.Fatal("RRULE not found")
	}
	if rule.Frequency != ical.Weekly {
		t.Fatalf("expected WEEKLY, got %s", rule.Frequency)
	}
	if rule.Until == nil {
		t.Fatal("expected UNTIL")
	}
	// Round-trip RRULE
	serialized := rule.String()
	if !strings.Contains(serialized, "FREQ=WEEKLY") {
		t.Fatalf("RRULE serialization missing FREQ: %s", serialized)
	}
}

func TestParseEventWithEscapedCharacters(t *testing.T) {
	data := readTestdata(t, "events/event_with_escaped_characters.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	props := event.PropertiesByName(ical.PropOrganizer)
	if len(props) == 0 {
		t.Fatal("ORGANIZER not found")
	}
	cn := props[0].Params.Get("CN")
	if cn == "" {
		t.Fatal("CN parameter not found")
	}
	// Round-trip
	reparsed, err := ical.ParseComponent([]byte(event.String()))
	if err != nil {
		t.Fatal(err)
	}
	props2 := reparsed.PropertiesByName(ical.PropOrganizer)
	if len(props2) == 0 {
		t.Fatal("ORGANIZER lost in round-trip")
	}
	cn2 := props2[0].Params.Get("CN")
	if cn2 != cn {
		t.Fatalf("CN round-trip mismatch: %q vs %q", cn, cn2)
	}
}

func TestParseEventWithUnicodeFields(t *testing.T) {
	data := readTestdata(t, "events/event_with_unicode_fields.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	summary := event.Text(ical.PropSummary)
	if summary == "" {
		t.Fatal("SUMMARY not found")
	}
	// Round-trip preserves unicode
	reparsed, err := ical.ParseComponent([]byte(event.String()))
	if err != nil {
		t.Fatal(err)
	}
	summary2 := reparsed.Text(ical.PropSummary)
	if summary2 != summary {
		t.Fatalf("SUMMARY round-trip mismatch: %q vs %q", summary, summary2)
	}
}

func TestParseEventWithUnicodeOrganizer(t *testing.T) {
	data := readTestdata(t, "events/event_with_unicode_organizer.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	props := event.PropertiesByName(ical.PropOrganizer)
	if len(props) == 0 {
		t.Fatal("ORGANIZER not found")
	}
	cn := props[0].Params.Get("CN")
	if cn == "" {
		t.Fatal("CN parameter not found")
	}
	// Round-trip preserves unicode CN
	reparsed, err := ical.ParseComponent([]byte(event.String()))
	if err != nil {
		t.Fatal(err)
	}
	props2 := reparsed.PropertiesByName(ical.PropOrganizer)
	if len(props2) == 0 {
		t.Fatal("ORGANIZER lost in round-trip")
	}
	cn2 := props2[0].Params.Get("CN")
	if cn2 != cn {
		t.Fatalf("CN round-trip mismatch: %q vs %q", cn, cn2)
	}
}

func TestParseCalendarWithUnicode(t *testing.T) {
	data := readTestdata(t, "calendars/calendar_with_unicode.ics")
	cal, err := ical.ParseCalendar(data)
	if err != nil {
		t.Fatal(err)
	}
	prodid := cal.Text(ical.PropProdID)
	if prodid == "" {
		t.Fatal("PRODID not found")
	}
	// Round-trip
	reparsed, err := ical.ParseCalendar([]byte(cal.String()))
	if err != nil {
		t.Fatal(err)
	}
	prodid2 := reparsed.Text(ical.PropProdID)
	if prodid2 != prodid {
		t.Fatalf("PRODID round-trip mismatch: %q vs %q", prodid, prodid2)
	}
}

func TestParseRFC5545RDATEExample(t *testing.T) {
	data := readTestdata(t, "calendars/rfc_5545_RDATE_example.ics")
	cal, err := ical.ParseCalendar(data)
	if err != nil {
		t.Fatal(err)
	}
	events := cal.Events()
	if len(events) != 6 {
		t.Fatalf("expected 6 events, got %d", len(events))
	}
	// Event UID:3 (4th event) has PERIOD RDATE
	event3 := events[3]
	periods, ok := event3.Periods(ical.PropRDate)
	if !ok {
		t.Fatal("event UID:3 should have PERIOD RDATE")
	}
	if len(periods) != 2 {
		t.Fatalf("expected 2 periods in event UID:3, got %d", len(periods))
	}
}

func TestParseMultipleCalendars(t *testing.T) {
	data := readTestdata(t, "calendars/multiple_calendar_components.ics")
	cals, err := ical.ParseCalendars(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(cals) != 2 {
		t.Fatalf("expected 2 calendars, got %d", len(cals))
	}
	for _, cal := range cals {
		if cal.Name != ical.CompVCalendar {
			t.Fatalf("expected VCALENDAR, got %s", cal.Name)
		}
	}
	// Round-trip each calendar individually
	for i, cal := range cals {
		reparsed, err := ical.ParseCalendar([]byte(cal.String()))
		if err != nil {
			t.Fatalf("calendar %d round-trip failed: %v", i, err)
		}
		if len(reparsed.Events()) == 0 {
			t.Fatalf("calendar %d has no events after round-trip", i)
		}
	}
}

func TestParseCalendarWithTimezone(t *testing.T) {
	data := readTestdata(t, "calendars/timezoned.ics")
	cal, err := ical.ParseCalendar(data)
	if err != nil {
		t.Fatal(err)
	}
	events := cal.Events()
	if len(events) == 0 {
		t.Fatal("no events found")
	}
	// First event should have DTSTART
	dtstart, ok := events[0].DateTime(ical.PropDTStart)
	if !ok {
		t.Fatal("DTSTART not found")
	}
	if dtstart.IsZero() {
		t.Fatal("DTSTART is zero")
	}
}

func TestParseEventWithSimpleSummary(t *testing.T) {
	data := readTestdata(t, "events/issue_100_transformed_doctests_into_unittests.ics")
	event, err := ical.ParseComponent(data)
	if err != nil {
		t.Fatal(err)
	}
	summary := event.Text(ical.PropSummary)
	if summary != "te" {
		t.Fatalf("expected SUMMARY=te, got %q", summary)
	}
	// Check LANGUAGE parameter
	props := event.PropertiesByName(ical.PropSummary)
	if len(props) == 0 {
		t.Fatal("SUMMARY not found")
	}
	if lang := props[0].Params.Get("LANGUAGE"); lang != "ru" {
		t.Fatalf("expected LANGUAGE=ru, got %q", lang)
	}
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read testdata %s: %v", name, err)
	}
	return data
}
