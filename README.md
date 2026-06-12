# ical

[![Go Reference](https://pkg.go.dev/badge/github.com/saltfishpr/ical.svg)](https://pkg.go.dev/github.com/saltfishpr/ical)

English | [中文](README.zh.md)

Go implementation of [RFC 5545](https://datatracker.ietf.org/doc/html/rfc5545) iCalendar parsing, building, validation, and serialization library.

## Installation

```bash
go get github.com/saltfishpr/ical
```

## Quick Start

### Parse

```go
// Read .ics file
data, _ := os.ReadFile("event.ics")
cal, err := ical.ParseCalendar(data)
```

### Build

```go
cal := ical.NewCalendar("//example.com//product")
event := ical.NewEvent("uid-1@example.com", time.Now())
event.AddDateTime(ical.PropDTStart, time.Now())
event.AddText(ical.PropSummary, "Weekly Team Meeting")
cal.AddComponent(event)

fmt.Println(cal) // Serialize to RFC 5545 text
```

### Validate

```go
errs := event.Validate()
for _, e := range errs {
    fmt.Println(e)
}
```

### Recurrence Expansion

```go
rulestr := "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;COUNT=10"
rule, err := ical.ParseRecurrenceRule(rulestr)

for dt := range rule.Expand(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)) {
    fmt.Println(dt)
}
```

## Core API

### Components

| Constructor               | Purpose   |
| ------------------------- | --------- |
| `NewCalendar(prodid)`     | VCALENDAR |
| `NewEvent(uid, stamp)`    | VEVENT    |
| `NewTodo(uid, stamp)`     | VTODO     |
| `NewJournal(uid, stamp)`  | VJOURNAL  |
| `NewFreeBusy(uid, stamp)` | VFREEBUSY |
| `NewTimezone(tzid)`       | VTIMEZONE |
| `NewAlarm()`              | VALARM    |

### Property Access

| Method                                 | RFC 5545 Type |
| -------------------------------------- | ------------- |
| `AddText` / `Text`                     | TEXT          |
| `AddDateTime` / `DateTime`             | DATE-TIME     |
| `AddDate` / `Date`                     | DATE          |
| `AddTime` / `Time`                     | TIME          |
| `AddDuration` / `Duration`             | DURATION      |
| `AddInteger` / `Integer`               | INTEGER       |
| `AddFloat` / `Float`                   | FLOAT         |
| `AddBoolean` / `Boolean`               | BOOLEAN       |
| `AddGeo` / `Geo`                       | GEO           |
| `AddPeriod` / `Period`                 | PERIOD        |
| `AddURI` / `URI`                       | URI           |
| `AddBinary` / `Binary`                 | BINARY        |
| `AddCalAddress`                        | CAL-ADDRESS   |
| `AddUTCOffset` / `UTCOffset`           | UTC-OFFSET    |
| `AddRecurrenceRule` / `RecurrenceRule` | RECUR         |

### Value Types

Full Format/Parse pairs:

- `EncodeText` / `DecodeText`
- `FormatBoolean` / `ParseBoolean`
- `FormatInteger` / `ParseInteger`
- `FormatFloat` / `ParseFloat`
- `FormatDateTime` / `ParseDateTime`
- `FormatDuration` / `ParseDuration`
- `FormatUTCOffset` / `ParseUTCOffset`
- `FormatPeriod` / `ParsePeriod` / `ParsePeriodList`
- `FormatGeo` / `ParseGeo`
- `FormatCalAddress`
- `FormatBinary` / `ParseBinary`
- `Time` / `ParseTime`

### Low-Level API

When the typed helpers aren't enough:

- `AddProperty` — manually add arbitrary properties
- `PropertiesByName` — retrieve all matching properties by name
- `AddComponent` — add child components (for nested structures)
- Unknown and custom `X-*` properties are fully preserved across parse → edit → serialize round-trips

## Acknowledgments

This project is inspired by and draws from [github.com/collective/icalendar](https://github.com/collective/icalendar).

## License

MIT
