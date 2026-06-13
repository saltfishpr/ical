package ical_test

import (
	"strings"
	"testing"
	"time"

	"github.com/saltfishpr/ical"
)

func TestParseMultiValueParameter(t *testing.T) {
	raw := "BEGIN:VEVENT\r\n" +
		"UID:multi@example.com\r\n" +
		"DTSTAMP:20260612T010203Z\r\n" +
		"ATTENDEE;MEMBER=\"mailto:a@example.com\";MEMBER=\"mailto:b@example.com\":mailto:group@example.com\r\n" +
		"END:VEVENT\r\n"

	event, err := ical.ParseComponent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	prop := event.PropertiesByName(ical.PropAttendee)[0]
	members := prop.Params.Values("MEMBER")
	if len(members) != 2 {
		t.Fatalf("expected 2 MEMBER values, got %d: %v", len(members), members)
	}
	if members[0] != "mailto:a@example.com" || members[1] != "mailto:b@example.com" {
		t.Fatalf("unexpected MEMBER values: %v", members)
	}
}

func TestParseParameterWithQuote(t *testing.T) {
	raw := "BEGIN:VEVENT\r\n" +
		"UID:quote@example.com\r\n" +
		"DTSTAMP:20260612T010203Z\r\n" +
		"ATTENDEE;CN=\"Group: Marketing, Sales\":mailto:group@example.com\r\n" +
		"END:VEVENT\r\n"

	event, err := ical.ParseComponent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	got := event.PropertiesByName(ical.PropAttendee)[0].Params.Get("CN")
	want := "Group: Marketing, Sales"
	if got != want {
		t.Fatalf("CN = %q, want %q", got, want)
	}
}

func TestParseParameterWithCaretNewline(t *testing.T) {
	raw := "BEGIN:VEVENT\r\n" +
		"UID:caret@example.com\r\n" +
		"DTSTAMP:20260612T010203Z\r\n" +
		"ATTENDEE;CN=Line1^nLine2:mailto:a@example.com\r\n" +
		"END:VEVENT\r\n"

	event, err := ical.ParseComponent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	got := event.PropertiesByName(ical.PropAttendee)[0].Params.Get("CN")
	want := "Line1\nLine2"
	if got != want {
		t.Fatalf("CN = %q, want %q", got, want)
	}
}

func TestParseParameterWithCaretQuote(t *testing.T) {
	raw := "BEGIN:VEVENT\r\n" +
		"UID:caretquote@example.com\r\n" +
		"DTSTAMP:20260612T010203Z\r\n" +
		"ATTENDEE;CN=He said ^'Hello^':mailto:a@example.com\r\n" +
		"END:VEVENT\r\n"

	event, err := ical.ParseComponent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	got := event.PropertiesByName(ical.PropAttendee)[0].Params.Get("CN")
	want := `He said "Hello"`
	if got != want {
		t.Fatalf("CN = %q, want %q", got, want)
	}
}

func TestParseCaretCaretEscape(t *testing.T) {
	raw := "BEGIN:VEVENT\r\n" +
		"UID:caretcaret@example.com\r\n" +
		"DTSTAMP:20260612T010203Z\r\n" +
		"ATTENDEE;CN=Value^^WithCaret:mailto:a@example.com\r\n" +
		"END:VEVENT\r\n"

	event, err := ical.ParseComponent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	got := event.PropertiesByName(ical.PropAttendee)[0].Params.Get("CN")
	want := "Value^WithCaret"
	if got != want {
		t.Fatalf("CN = %q, want %q", got, want)
	}
}

func TestParseComponentWithLargeCount(t *testing.T) {
	// Repeat properties like ATTENDEE should all be preserved
	var b strings.Builder
	b.WriteString("BEGIN:VEVENT\r\n")
	b.WriteString("UID:many@example.com\r\n")
	b.WriteString("DTSTAMP:20260612T010203Z\r\n")
	for i := 0; i < 100; i++ {
		b.WriteString("ATTENDEE:mailto:person")
		b.WriteString(string(rune('0' + i%10)))
		b.WriteString("@example.com\r\n")
	}
	b.WriteString("END:VEVENT\r\n")

	event, err := ical.ParseComponent([]byte(b.String()))
	if err != nil {
		t.Fatal(err)
	}

	attendees := event.PropertiesByName(ical.PropAttendee)
	if len(attendees) != 100 {
		t.Fatalf("expected 100 ATTENDEE, got %d", len(attendees))
	}
}

func TestSerializePropertyRoundTrip(t *testing.T) {
	event := ical.NewEvent("roundtrip@example.com", time.Date(2026, 6, 12, 1, 2, 3, 0, time.UTC))
	event.AddText(ical.PropSummary, "Test , Event ; with \\ special chars")
	event.AddText(ical.PropDescription, "Line1\nLine2")

	parsed, err := ical.ParseComponent([]byte(event.String()))
	if err != nil {
		t.Fatal(err)
	}

	summary := parsed.Text(ical.PropSummary)
	want := "Test , Event ; with \\ special chars"
	if summary != want {
		t.Fatalf("SUMMARY = %q, want %q", summary, want)
	}

	desc := parsed.Text(ical.PropDescription)
	if desc != "Line1\nLine2" {
		t.Fatalf("DESCRIPTION = %q, want %q", desc, "Line1\nLine2")
	}
}

func TestParseParameterPreservesBackslashAndDecodesCaretEscapes(t *testing.T) {
	raw := "BEGIN:VEVENT\r\n" +
		"UID:param@example.com\r\n" +
		"DTSTAMP:20260612T010203Z\r\n" +
		"ATTENDEE;CN=\"C:\\Temp^'A^nB^^\":mailto:a@example.com\r\n" +
		"END:VEVENT\r\n"

	event, err := ical.ParseComponent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	got := event.PropertiesByName(ical.PropAttendee)[0].Params.Get("CN")
	want := "C:\\Temp\"A\nB^"
	if got != want {
		t.Fatalf("CN = %q, want %q", got, want)
	}
}

func TestSerializedParameterBackslashRoundTrips(t *testing.T) {
	event := ical.NewEvent("param-roundtrip@example.com", time.Date(2026, 6, 12, 1, 2, 3, 0, time.UTC))
	event.AddCalAddress(ical.PropAttendee, "a@example.com", ical.Params{}.Set("CN", "C:\\Temp"))

	parsed, err := ical.ParseComponent([]byte(event.String()))
	if err != nil {
		t.Fatal(err)
	}

	got := parsed.PropertiesByName(ical.PropAttendee)[0].Params.Get("CN")
	want := "C:\\Temp"
	if got != want {
		t.Fatalf("CN = %q, want %q", got, want)
	}
}
