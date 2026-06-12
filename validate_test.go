package ical_test

import (
	"testing"
	"time"

	"github.com/saltfishpr/ical"
)

func TestValidateTimezoneObservanceRequiresOffsetsAndStart(t *testing.T) {
	tz := ical.NewTimezone("America/New_York")
	tz.AddComponent(ical.NewComponent(ical.CompStandard))

	errs := tz.Validate()
	for _, property := range []string{ical.PropDTStart, ical.PropTZOffsetFrom, ical.PropTZOffsetTo} {
		if !hasValidationError(errs, ical.CompStandard, property, "required") {
			t.Fatalf("Validate() missing required error for %s: %#v", property, errs)
		}
	}
}

func TestValidateTimezoneObservanceAcceptsRequiredProperties(t *testing.T) {
	tz := ical.NewTimezone("America/New_York")
	standard := ical.NewComponent(ical.CompStandard)
	standard.AddDateTime(ical.PropDTStart, time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC))
	standard.AddUTCOffset(ical.PropTZOffsetFrom, -4*time.Hour)
	standard.AddUTCOffset(ical.PropTZOffsetTo, -5*time.Hour)
	tz.AddComponent(standard)

	if errs := tz.Validate(); len(errs) != 0 {
		t.Fatalf("Validate() returned errors: %#v", errs)
	}
}

func TestNewTimezoneStandard(t *testing.T) {
	standard := ical.NewTimezoneStandard(
		time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC),
		-4*time.Hour,
		-5*time.Hour,
	)
	if standard.Name != ical.CompStandard {
		t.Fatalf("expected STANDARD, got %s", standard.Name)
	}
	dtstart, ok := standard.DateTime(ical.PropDTStart)
	if !ok {
		t.Fatal("DTSTART not found")
	}
	if dtstart.Year() != 2026 {
		t.Fatalf("expected 2026, got %d", dtstart.Year())
	}
	offsetFrom, ok := standard.UTCOffset(ical.PropTZOffsetFrom)
	if !ok || offsetFrom != -4*time.Hour {
		t.Fatalf("expected TZOFFSETFROM=-4h, got %v", offsetFrom)
	}
	offsetTo, ok := standard.UTCOffset(ical.PropTZOffsetTo)
	if !ok || offsetTo != -5*time.Hour {
		t.Fatalf("expected TZOFFSETTO=-5h, got %v", offsetTo)
	}
}

func TestNewTimezoneDaylight(t *testing.T) {
	daylight := ical.NewTimezoneDaylight(
		time.Date(2026, 3, 8, 2, 0, 0, 0, time.UTC),
		-5*time.Hour,
		-4*time.Hour,
	)
	if daylight.Name != ical.CompDaylight {
		t.Fatalf("expected DAYLIGHT, got %s", daylight.Name)
	}
}

func TestTimezoneWithStandardAndDaylight(t *testing.T) {
	tz := ical.NewTimezone("America/New_York")
	standard := ical.NewTimezoneStandard(
		time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC),
		-4*time.Hour,
		-5*time.Hour,
	)
	daylight := ical.NewTimezoneDaylight(
		time.Date(2026, 3, 8, 2, 0, 0, 0, time.UTC),
		-5*time.Hour,
		-4*time.Hour,
	)
	tz.AddComponent(standard)
	tz.AddComponent(daylight)

	// Round-trip: serialize and re-parse
	parsed, err := ical.ParseComponent([]byte(tz.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Standard()) != 1 {
		t.Fatal("STANDARD lost in round-trip")
	}
	if len(parsed.Daylight()) != 1 {
		t.Fatal("DAYLIGHT lost in round-trip")
	}
	// Validate should pass with complete observance
	if errs := tz.Validate(); len(errs) != 0 {
		t.Fatalf("Validate() returned errors: %#v", errs)
	}
}

func hasValidationError(errs []ical.ValidationError, component, property, rule string) bool {
	for _, err := range errs {
		if err.Component == component && err.Property == property && err.Rule == rule {
			return true
		}
	}
	return false
}
