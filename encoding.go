package ical

import "strings"

// EncodeText escapes special characters (\, ;, ,, newline) in an RFC 5545
// TEXT value. Use this when you need to manually create a TEXT property value;
// Component.AddText applies this encoding automatically.
func EncodeText(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case ';':
			b.WriteString(`\;`)
		case ',':
			b.WriteString(`\,`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// DecodeText unescapes RFC 5545 escape sequences (\n, \\, \;, \,) in a TEXT
// value. Component.Text applies this decoding automatically.
func DecodeText(value string) string {
	var b strings.Builder
	escaped := false
	for _, r := range value {
		if escaped {
			switch r {
			case 'n', 'N':
				b.WriteByte('\n')
			default:
				b.WriteRune(r)
			}
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		b.WriteRune(r)
	}
	if escaped {
		b.WriteByte('\\')
	}
	return b.String()
}

func encodeParamValue(value string) string {
	value = strings.ReplaceAll(value, "^", "^^")
	value = strings.ReplaceAll(value, "\n", "^n")
	value = strings.ReplaceAll(value, `"`, "^'")
	if strings.ContainsAny(value, ",;:") {
		return `"` + value + `"`
	}
	return value
}

func decodeParamValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	var b strings.Builder
	escaped := false
	for _, r := range value {
		if escaped {
			switch r {
			case 'n':
				b.WriteByte('\n')
			case '\'':
				b.WriteByte('"')
			case '^':
				b.WriteByte('^')
			default:
				b.WriteRune(r)
			}
			escaped = false
			continue
		}
		if r == '^' {
			escaped = true
			continue
		}
		b.WriteRune(r)
	}
	if escaped {
		b.WriteByte('\\')
	}
	return b.String()
}
