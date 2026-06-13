package ical

import (
	"sort"
	"strings"
)

// Property is one RFC 5545 content line after parsing.
//
// A Property consists of a case-insensitive name, optional parameters
// (e.g., VALUE=DATE, TZID=America/New_York), and a value string that is
// already unescaped and unfolded. Use the Component typed accessors
// (DateTime, Duration, etc.) to parse the Value field; use the typed
// Add methods to build Properties.
type Property struct {
	Name   string
	Params Params
	Value  string
}

func (p Property) normalized() Property {
	p.Name = strings.ToUpper(p.Name)
	params := make(Params, len(p.Params))
	for k, v := range p.Params {
		cp := make([]string, len(v))
		copy(cp, v)
		params[strings.ToUpper(k)] = cp
	}
	p.Params = params
	return p
}

func (p Property) contentLine() string {
	p = p.normalized()
	var b strings.Builder
	b.WriteString(p.Name)
	if len(p.Params) > 0 {
		keys := make([]string, 0, len(p.Params))
		for key := range p.Params {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			b.WriteByte(';')
			b.WriteString(key)
			b.WriteByte('=')
			values := p.Params[key]
			for i, value := range values {
				if i > 0 {
					b.WriteByte(',')
				}
				b.WriteString(encodeParamValue(value))
			}
		}
	}
	b.WriteByte(':')
	b.WriteString(p.Value)
	return b.String()
}

// Params stores RFC 5545 property parameters. Names are case-insensitive and
// serialized in uppercase.
type Params map[string][]string

// Get returns the first value for a parameter name, or an empty string when
// the parameter has no values.
func (p Params) Get(name string) string {
	values := p.Values(name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Values returns all values for a parameter name, or nil when the parameter
// has no values.
func (p Params) Values(name string) []string {
	if p == nil {
		return nil
	}
	values := p[strings.ToUpper(name)]
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

// Set replaces all values for a parameter name with the given values.
// Names are case-insensitive and normalized to uppercase.
// It returns p so calls can be chained: Params{}.Set(ParamValue, ValueTime).Set(ParamTZID, tzid).
func (p Params) Set(name string, values ...string) Params {
	if p == nil {
		p = make(Params)
	}
	cp := make([]string, len(values))
	copy(cp, values)
	p[strings.ToUpper(name)] = cp
	return p
}
