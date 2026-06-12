package ical_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/saltfishpr/ical"
)

func ExampleNewCalendar() {
	cal := ical.NewCalendar("-//example//booking//EN")
	event := ical.NewEvent("meeting-1@example.com", time.Date(2026, 6, 11, 1, 2, 3, 0, time.UTC))
	event.AddDateTimeWithTZID("DTSTART", time.Date(2026, 6, 11, 9, 30, 0, 0, time.FixedZone("Asia/Shanghai", 8*60*60)), "Asia/Shanghai")
	event.AddDuration("DURATION", time.Hour)
	event.AddText("SUMMARY", "Planning review")
	cal.AddComponent(event)

	fmt.Print(strings.ReplaceAll(cal.String(), "\r\n", "\n"))
	// Output:
	// BEGIN:VCALENDAR
	// VERSION:2.0
	// PRODID:-//example//booking//EN
	// BEGIN:VEVENT
	// UID:meeting-1@example.com
	// DTSTAMP:20260611T010203Z
	// DTSTART;TZID=Asia/Shanghai:20260611T093000
	// DURATION:PT1H
	// SUMMARY:Planning review
	// END:VEVENT
	// END:VCALENDAR
}

func ExampleParseCalendar() {
	raw := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//example//EN\r\nBEGIN:VEVENT\r\nUID:1@example.com\r\nDTSTAMP:20260611T010203Z\r\nSUMMARY:Demo\\, review\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	cal, _ := ical.ParseCalendar([]byte(raw))
	event := cal.ComponentsByName("VEVENT")[0]

	fmt.Println(event.Text("SUMMARY"))
	// Output: Demo, review
}

func ExampleNewEvent_withRecurrence() {
	event := ical.NewEvent("recur@example.com", time.Date(2026, 6, 12, 1, 2, 3, 0, time.UTC))
	event.AddDateTime(ical.PropDTStart, time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC))
	event.AddRecurrenceRule(ical.RecurrenceRule{
		Frequency: ical.Weekly,
		ByDay:     []ical.Weekday{ical.Monday, ical.Wednesday, ical.Friday},
		Count:     6,
	})
	event.AddText(ical.PropSummary, "Recurring standup")

	fmt.Print(strings.ReplaceAll(event.String(), "\r\n", "\n"))
	// Output:
	// BEGIN:VEVENT
	// UID:recur@example.com
	// DTSTAMP:20260612T010203Z
	// DTSTART:20260612T100000Z
	// RRULE:FREQ=WEEKLY;COUNT=6;BYDAY=MO,WE,FR
	// SUMMARY:Recurring standup
	// END:VEVENT
}

func ExampleNewTimezone() {
	tz := ical.NewTimezone("America/New_York")
	standard := ical.NewTimezoneStandard(
		time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC),
		-4*time.Hour,
		-5*time.Hour,
	)
	tz.AddComponent(standard)

	errs := tz.Validate()
	fmt.Println(len(errs))
	// Output: 0
}

func ExampleComponent_Validate() {
	// An event without required DTSTAMP is invalid
	event := ical.NewComponent(ical.CompVEvent)
	event.AddText(ical.PropUID, "test@example.com")

	errs := event.Validate()
	for _, err := range errs {
		fmt.Println(err.Rule, err.Property)
	}
	// Output:
	// required DTSTAMP
}

func ExampleParseRecurrenceRule() {
	rule, _ := ical.ParseRecurrenceRule("FREQ=MONTHLY;BYDAY=2TH;COUNT=3")
	start := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	for t := range rule.Expand(start) {
		fmt.Println(t.Format("2006-01-02"))
	}
	// Output:
	// 2026-06-11
	// 2026-07-09
	// 2026-08-13
}

func ExampleParseCalendars() {
	raw := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//one//EN\r\nBEGIN:VEVENT\r\nUID:one@example.com\r\nDTSTAMP:20260612T010203Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n" +
		"BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//two//EN\r\nBEGIN:VEVENT\r\nUID:two@example.com\r\nDTSTAMP:20260612T010203Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	cals, _ := ical.ParseCalendars([]byte(raw))
	fmt.Println(len(cals))
	// Output: 2
}
