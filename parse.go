package ical

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// ParseCalendar parses one VCALENDAR object from an RFC 5545 byte stream.
// If the stream contains multiple VCALENDAR objects, only the first is returned.
// Use [ParseCalendars] to parse streams with multiple calendars.
func ParseCalendar(data []byte) (*Component, error) {
	c, err := ParseComponent(data)
	if err != nil {
		return nil, err
	}
	if c.Name != CompVCalendar {
		return nil, fmt.Errorf("ical: expected VCALENDAR, got %s", c.Name)
	}
	return c, nil
}

// ParseCalendars parses one or more VCALENDAR objects from an RFC 5545 byte stream.
// This supports streams that contain multiple calendars concatenated together.
// Returns an error if no valid calendar is found.
func ParseCalendars(data []byte) ([]*Component, error) {
	calendars, err := parseComponents(data)
	if err != nil {
		return nil, err
	}
	if len(calendars) == 0 {
		return nil, fmt.Errorf("ical: no calendar found")
	}
	return calendars, nil
}

// ParseComponent parses any single RFC 5545 component (VCALENDAR, VEVENT,
// VTIMEZONE, etc.) from a byte stream. Use [ParseCalendar] when you specifically
// expect a VCALENDAR object.
func ParseComponent(data []byte) (*Component, error) {
	roots, err := parseComponents(data)
	if err != nil {
		return nil, err
	}
	if len(roots) != 1 {
		return nil, fmt.Errorf("ical: expected one root component, got %d", len(roots))
	}
	return roots[0], nil
}

// parseComponents parses all root components from unfolded lines.
// It is the shared parsing core for ParseComponent and ParseCalendars.
func parseComponents(data []byte) ([]*Component, error) {
	lines, err := unfoldLines(data)
	if err != nil {
		return nil, err
	}
	var roots []*Component
	var stack []*Component
	for _, line := range lines {
		prop, err := parseContentLine(line)
		if err != nil {
			return nil, err
		}
		switch prop.Name {
		case TokenBegin:
			component := NewComponent(prop.Value)
			if len(stack) > 0 {
				stack[len(stack)-1].AddComponent(component)
			}
			stack = append(stack, component)
		case TokenEnd:
			if len(stack) == 0 {
				return nil, fmt.Errorf("ical: END:%s without BEGIN", prop.Value)
			}
			top := stack[len(stack)-1]
			if top.Name != strings.ToUpper(prop.Value) {
				return nil, fmt.Errorf("ical: END:%s closes %s", prop.Value, top.Name)
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				roots = append(roots, top)
			}
		default:
			if len(stack) == 0 {
				return nil, fmt.Errorf("ical: property %s outside component", prop.Name)
			}
			stack[len(stack)-1].AddProperty(prop)
		}
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("ical: unclosed component %s", stack[len(stack)-1].Name)
	}
	return roots, nil
}

func unfoldLines(data []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			continue
		}
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if len(lines) == 0 {
				return nil, fmt.Errorf("ical: folded continuation without content line")
			}
			lines[len(lines)-1] += line[1:]
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func parseContentLine(line string) (Property, error) {
	split := -1
	inQuote := false
	escaped := false
	for i, r := range line {
		if escaped {
			escaped = false
			continue
		}
		switch r {
		case '\\':
			escaped = true
		case '"':
			inQuote = !inQuote
		case ':':
			if !inQuote {
				split = i
				goto found
			}
		}
	}
found:
	if split <= 0 {
		return Property{}, fmt.Errorf("ical: invalid content line %q", line)
	}
	left := line[:split]
	value := line[split+1:]
	parts := splitQuoted(left, ';')
	if len(parts) == 0 || parts[0] == "" {
		return Property{}, fmt.Errorf("ical: content line has no property name")
	}
	prop := Property{Name: strings.ToUpper(strings.TrimSpace(parts[0])), Params: Params{}, Value: value}
	for _, raw := range parts[1:] {
		key, val, ok := strings.Cut(raw, "=")
		if !ok {
			return Property{}, fmt.Errorf("ical: parameter %q has no value", raw)
		}
		key = strings.ToUpper(strings.TrimSpace(key))
		for _, item := range splitQuoted(val, ',') {
			prop.Params[key] = append(prop.Params[key], decodeParamValue(item))
		}
	}
	return prop, nil
}

func splitQuoted(s string, sep rune) []string {
	var out []string
	var b strings.Builder
	inQuote := false
	escaped := false
	for _, r := range s {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			b.WriteRune(r)
			escaped = true
			continue
		}
		if r == '"' {
			inQuote = !inQuote
			b.WriteRune(r)
			continue
		}
		if r == sep && !inQuote {
			out = append(out, b.String())
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
	out = append(out, b.String())
	return out
}
