// Package ical parses, edits, and serializes RFC 5545 iCalendar data.
//
// # Overview
//
// iCalendar is the MIME type text/calendar, used for exchanging calendar and
// scheduling information. This package models the RFC 5545 object model with
// two core types: [Component] (VEVENT, VTODO, etc.) and [Property] (SUMMARY,
// DTSTART, etc.). A [Component] is a named container with zero or more
// [Property] values and zero or more child [Component] values.
//
// # Parsing
//
// Use [ParseCalendar] to parse a single VCALENDAR from an RFC 5545 byte stream,
// or [ParseCalendars] when the stream may contain multiple calendars.
//
// # Building
//
// Helper constructors ([NewCalendar], [NewEvent], [NewTodo], [NewTimezone], etc.)
// create components pre-populated with required properties. Each constructor
// accepts the minimum mandatory arguments so callers cannot forget them.
//
// Typed Add methods ([Component.AddText], [Component.AddDateTime],
// [Component.AddDuration], etc.) handle value formatting according to RFC 5545
// rules. Corresponding typed accessors ([Component.Text], [Component.DateTime],
// [Component.Duration], etc.) parse the first matching property.
//
// # Serialization
//
// Call [Component.String] to serialize any component tree back to RFC 5545 text
// with proper CRLF line endings and content-line folding at 75 octets.
//
// # Validation
//
// [Component.Validate] checks RFC 5545 property constraints — required
// properties, singleton enforcement, mutual exclusion, and co-occurrence —
// returning a list of [ValidationError] values.
//
// # Recurrence
//
// The [RecurrenceRule] type and [ParseRecurrenceRule] handle RRULE values.
// [RecurrenceRule.Expand] produces an iterator ([iter.Seq]) over recurrence
// instances bounded by COUNT or UNTIL, supporting common BYDAY, BYMONTHDAY, and
// BYSETPOS patterns.
//
// # Values
//
// The package provides Format/Parse pairs for each RFC 5545 data type:
// [FormatBoolean]/[ParseBoolean], [FormatDateTime]/[ParseDateTime],
// [FormatDuration]/[ParseDuration], and so on. Custom value types ([Geo],
// [Time], [Period]) implement [fmt.Stringer] for round-tripping.
//
// # Low-level access
//
// [Component.AddProperty], [Component.PropertiesByName], and
// [Component.AddComponent] offer direct manipulation for properties and
// components not covered by typed helpers. Unknown or custom X-* entities are
// preserved through parse → edit → serialize round-trips.
package ical
