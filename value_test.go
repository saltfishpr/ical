package ical_test

import (
	"testing"
	"time"

	"github.com/saltfishpr/ical"
)

func TestDateTimeParsesTZIDParameter(t *testing.T) {
	raw := "BEGIN:VCALENDAR\r\n" +
		"VERSION:2.0\r\n" +
		"PRODID:-//example//EN\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:tzid@example.com\r\n" +
		"DTSTAMP:20260612T010203Z\r\n" +
		"DTSTART;TZID=Asia/Shanghai:20260612T093000\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"
	cal, err := ical.ParseCalendar([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	got, ok := cal.Events()[0].DateTime(ical.PropDTStart)
	if !ok {
		t.Fatal("DateTime returned ok=false")
	}
	want := time.Date(2026, 6, 12, 9, 30, 0, 0, location)
	if !got.Equal(want) || got.Location().String() != "Asia/Shanghai" {
		t.Fatalf("got %s in %s, want %s in %s", got, got.Location(), want, want.Location())
	}
}

func TestDateTimesParsesTZIDParameter(t *testing.T) {
	raw := "BEGIN:VEVENT\r\n" +
		"UID:tzid-list@example.com\r\n" +
		"DTSTAMP:20260612T010203Z\r\n" +
		"DTSTART;TZID=Asia/Shanghai:20260612T093000\r\n" +
		"RDATE;TZID=Asia/Shanghai:20260613T093000,20260614T093000\r\n" +
		"END:VEVENT\r\n"
	event, err := ical.ParseComponent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	got, ok := event.DateTimes(ical.PropRDate)
	if !ok {
		t.Fatal("DateTimes returned ok=false")
	}
	want := []time.Time{
		time.Date(2026, 6, 13, 9, 30, 0, 0, location),
		time.Date(2026, 6, 14, 9, 30, 0, 0, location),
	}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if !got[i].Equal(want[i]) || got[i].Location().String() != "Asia/Shanghai" {
			t.Fatalf("got[%d] = %s in %s, want %s in %s", i, got[i], got[i].Location(), want[i], want[i].Location())
		}
	}
}

func TestPeriodParsesStartDuration(t *testing.T) {
	period, err := ical.ParsePeriod("19970101T180000Z/PT5H30M")
	if err != nil {
		t.Fatal(err)
	}
	if period.Duration != 5*time.Hour+30*time.Minute {
		t.Fatalf("expected PT5H30M, got %v", period.Duration)
	}
	if period.Start.Year() != 1997 {
		t.Fatalf("expected 1997, got %d", period.Start.Year())
	}
}

func TestPeriodListParse(t *testing.T) {
	value := "19970101T180000Z/19970102T070000Z,19970109T180000Z/PT5H30M"
	periods, err := ical.ParsePeriodList(value)
	if err != nil {
		t.Fatal(err)
	}
	if len(periods) != 2 {
		t.Fatalf("expected 2 periods, got %d", len(periods))
	}
	if periods[1].Duration != 5*time.Hour+30*time.Minute {
		t.Fatalf("expected PT5H30M, got %v", periods[1].Duration)
	}
}

func TestDurationFormatAndParse(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{5*time.Hour + 30*time.Minute, "PT5H30M"},
		{90 * time.Minute, "PT1H30M"},
		{7 * 24 * time.Hour, "P1W"},
		{2 * 7 * 24 * time.Hour, "P2W"},
		{3 * 24 * time.Hour, "P3D"},
		{0, "PT0S"},
		{1 * time.Second, "PT1S"},
		{-1 * time.Hour, "-PT1H"},
	}
	for _, tt := range tests {
		formatted := ical.FormatDuration(tt.duration)
		if formatted != tt.expected {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, formatted, tt.expected)
		}
		parsed, err := ical.ParseDuration(formatted)
		if err != nil {
			t.Errorf("ParseDuration(%q): %v", formatted, err)
			continue
		}
		if parsed != tt.duration {
			t.Errorf("ParseDuration(%q) = %v, want %v", formatted, parsed, tt.duration)
		}
	}
}

func TestUTCOffsetWithSeconds(t *testing.T) {
	offset, err := ical.ParseUTCOffset("+053002")
	if err != nil {
		t.Fatal(err)
	}
	expected := 5*time.Hour + 30*time.Minute + 2*time.Second
	if offset != expected {
		t.Fatalf("expected %v, got %v", expected, offset)
	}
	formatted := ical.FormatUTCOffset(offset)
	if formatted != "+053002" {
		t.Fatalf("expected +053002, got %q", formatted)
	}
}

func TestParseUTCOffsetInvalid(t *testing.T) {
	_, err := ical.ParseUTCOffset("-0000")
	if err == nil {
		t.Fatal("expected error for -0000")
	}
	_, err = ical.ParseUTCOffset("+2400")
	if err == nil {
		t.Fatal("expected error for +2400")
	}
}

func TestParseBooleanInvalid(t *testing.T) {
	_, err := ical.ParseBoolean("YES")
	if err == nil {
		t.Fatal("expected error for YES")
	}
}

func TestParseFloatNaN(t *testing.T) {
	_, err := ical.ParseFloat("NaN")
	if err == nil {
		t.Fatal("expected error for NaN")
	}
	_, err = ical.ParseFloat("Inf")
	if err == nil {
		t.Fatal("expected error for Inf")
	}
}

func TestParseGeoInvalid(t *testing.T) {
	_, err := ical.ParseGeo("12.34")
	if err == nil {
		t.Fatal("expected error for missing semicolon")
	}
}

func TestParsePeriodInvalid(t *testing.T) {
	_, err := ical.ParsePeriod("not-a-period")
	if err == nil {
		t.Fatal("expected error for invalid period")
	}
}

func TestParseDurationInvalid(t *testing.T) {
	_, err := ical.ParseDuration("not-a-duration")
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
	_, err = ical.ParseDuration("P")
	if err == nil {
		t.Fatal("expected error for empty duration")
	}
}

func TestAddPeriodsRoundTripsPeriodList(t *testing.T) {
	freeBusy := ical.NewFreeBusy("freebusy@example.com", time.Date(2026, 6, 12, 1, 2, 3, 0, time.UTC))
	first := ical.Period{
		Start: time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC),
	}
	second := ical.Period{
		Start:    time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC),
		Duration: 90 * time.Minute,
	}
	freeBusy.AddPeriods(ical.PropFreeBusy, first, second)

	parsed, err := ical.ParseComponent([]byte(freeBusy.String()))
	if err != nil {
		t.Fatal(err)
	}
	prop := parsed.PropertiesByName(ical.PropFreeBusy)[0]
	if got := prop.Params.Get(ical.ParamValue); got != ical.ValuePeriod {
		t.Fatalf("FREEBUSY VALUE = %q, want %q", got, ical.ValuePeriod)
	}

	got, ok := parsed.Periods(ical.PropFreeBusy)
	if !ok {
		t.Fatal("Periods returned ok=false")
	}
	want := []ical.Period{first, second}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if !got[i].Start.Equal(want[i].Start) || !got[i].End.Equal(want[i].End) || got[i].Duration != want[i].Duration {
			t.Fatalf("got[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
