package ical

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Component is an RFC 5545 component such as VCALENDAR, VEVENT, VTODO,
// VTIMEZONE, or a custom X-* component.
type Component struct {
	Name       string
	Properties []Property
	Components []*Component
}

// NewComponent creates an empty component with the given RFC 5545 name.
// Use the Comp* constants for standard component names, or provide a custom
// name for X-* extension components.
func NewComponent(name string) *Component {
	return &Component{Name: strings.ToUpper(name)}
}

// NewCalendar creates a VCALENDAR with required VERSION and PRODID values.
func NewCalendar(prodid string) *Component {
	return NewComponent(CompVCalendar).
		AddText(PropVersion, "2.0").
		AddText(PropProdID, prodid)
}

// NewEvent creates a VEVENT with required UID and DTSTAMP values.
func NewEvent(uid string, stamp time.Time) *Component {
	return NewComponent(CompVEvent).
		AddText(PropUID, uid).
		AddDateTime(PropDTStamp, stamp.UTC())
}

// NewTodo creates a VTODO with required UID and DTSTAMP values.
func NewTodo(uid string, stamp time.Time) *Component {
	return NewComponent(CompVTodo).
		AddText(PropUID, uid).
		AddDateTime(PropDTStamp, stamp.UTC())
}

// NewJournal creates a VJOURNAL with required UID and DTSTAMP values.
func NewJournal(uid string, stamp time.Time) *Component {
	return NewComponent(CompVJournal).
		AddText(PropUID, uid).
		AddDateTime(PropDTStamp, stamp.UTC())
}

// NewFreeBusy creates a VFREEBUSY with required UID and DTSTAMP values.
func NewFreeBusy(uid string, stamp time.Time) *Component {
	return NewComponent(CompVFreeBusy).
		AddText(PropUID, uid).
		AddDateTime(PropDTStamp, stamp.UTC())
}

// NewTimezone creates a VTIMEZONE with the required TZID value.
func NewTimezone(tzid string) *Component {
	return NewComponent(CompVTimezone).
		AddText(PropTZID, tzid)
}

// NewTimezoneStandard creates a STANDARD observance component with the required
// DTSTART, TZOFFSETFROM, and TZOFFSETTO values.
func NewTimezoneStandard(dtstart time.Time, offsetFrom, offsetTo time.Duration) *Component {
	return NewComponent(CompStandard).
		AddDateTime(PropDTStart, dtstart).
		AddUTCOffset(PropTZOffsetFrom, offsetFrom).
		AddUTCOffset(PropTZOffsetTo, offsetTo)
}

// NewTimezoneDaylight creates a DAYLIGHT observance component with the required
// DTSTART, TZOFFSETFROM, and TZOFFSETTO values.
func NewTimezoneDaylight(dtstart time.Time, offsetFrom, offsetTo time.Duration) *Component {
	return NewComponent(CompDaylight).
		AddDateTime(PropDTStart, dtstart).
		AddUTCOffset(PropTZOffsetFrom, offsetFrom).
		AddUTCOffset(PropTZOffsetTo, offsetTo)
}

// AddProperty appends a raw property. Unknown RFC/IANA/X-* properties are
// intentionally preserved by this generic model.
func (c *Component) AddProperty(p Property) *Component {
	c.Properties = append(c.Properties, p.normalized())
	return c
}

// AddText appends a TEXT property, escaping the value for RFC 5545 output.
func (c *Component) AddText(name, value string) *Component {
	return c.AddProperty(Property{Name: name, Value: EncodeText(value)})
}

// AddBoolean appends a BOOLEAN property.
func (c *Component) AddBoolean(name string, value bool) *Component {
	return c.AddProperty(Property{Name: name, Value: FormatBoolean(value)})
}

// AddInteger appends an INTEGER property.
func (c *Component) AddInteger(name string, value int) *Component {
	return c.AddProperty(Property{Name: name, Value: FormatInteger(value)})
}

// AddFloat appends a FLOAT property.
func (c *Component) AddFloat(name string, value float64) *Component {
	return c.AddProperty(Property{Name: name, Value: FormatFloat(value)})
}

// AddGeo appends a GEO property.
func (c *Component) AddGeo(value Geo) *Component {
	return c.AddProperty(Property{Name: PropGeo, Value: value.String()})
}

// AddURI appends a URI property.
func (c *Component) AddURI(name, value string) *Component {
	return c.AddProperty(Property{Name: name, Value: value})
}

// AddTime appends a TIME property with VALUE=TIME.
func (c *Component) AddTime(name string, value Time) *Component {
	return c.AddProperty(Property{Name: name, Params: Params{}.Set(ParamValue, ValueTime), Value: value.String()})
}

// AddTimeWithTZID appends a local TIME property with a TZID parameter.
func (c *Component) AddTimeWithTZID(name string, value Time, tzid string) *Component {
	return c.AddProperty(Property{Name: name, Params: Params{}.Set(ParamValue, ValueTime).Set(ParamTZID, tzid), Value: value.String()})
}

// AddBinary appends a BINARY property using BASE64 inline encoding.
func (c *Component) AddBinary(name string, value []byte) *Component {
	return c.AddProperty(Property{Name: name, Params: Params{}.Set(ParamEncoding, ValueBase64).Set(ParamValue, ValueBinary), Value: FormatBinary(value)})
}

// AddCalAddress appends a CAL-ADDRESS property. A bare email address is
// normalized to a mailto URI.
func (c *Component) AddCalAddress(name, value string, params Params) *Component {
	return c.AddProperty(Property{Name: name, Params: params, Value: FormatCalAddress(value)})
}

// AddDateTime appends a DATE-TIME property.
func (c *Component) AddDateTime(name string, value time.Time) *Component {
	return c.AddProperty(Property{Name: name, Value: FormatDateTime(value)})
}

// AddDateTimes appends a comma-separated DATE-TIME property.
func (c *Component) AddDateTimes(name string, values ...time.Time) *Component {
	encoded := make([]string, len(values))
	for i, value := range values {
		encoded[i] = FormatDateTime(value)
	}
	return c.AddProperty(Property{Name: name, Value: strings.Join(encoded, ",")})
}

// AddDateTimeWithTZID appends a local DATE-TIME property with a TZID parameter.
func (c *Component) AddDateTimeWithTZID(name string, value time.Time, tzid string) *Component {
	return c.AddProperty(Property{Name: name, Params: Params{}.Set(ParamTZID, tzid), Value: value.Format("20060102T150405")})
}

// AddDate appends a DATE property with VALUE=DATE.
func (c *Component) AddDate(name string, value time.Time) *Component {
	return c.AddProperty(Property{Name: name, Params: Params{}.Set(ParamValue, ValueDate), Value: value.Format("20060102")})
}

// AddDates appends a comma-separated DATE property with VALUE=DATE.
func (c *Component) AddDates(name string, values ...time.Time) *Component {
	encoded := make([]string, len(values))
	for i, value := range values {
		encoded[i] = value.Format("20060102")
	}
	return c.AddProperty(Property{Name: name, Params: Params{}.Set(ParamValue, ValueDate), Value: strings.Join(encoded, ",")})
}

// AddDuration appends a DURATION property.
func (c *Component) AddDuration(name string, value time.Duration) *Component {
	return c.AddProperty(Property{Name: name, Value: FormatDuration(value)})
}

// AddUTCOffset appends a UTC-OFFSET property.
func (c *Component) AddUTCOffset(name string, value time.Duration) *Component {
	return c.AddProperty(Property{Name: name, Value: FormatUTCOffset(value)})
}

// AddPeriod appends a PERIOD property with VALUE=PERIOD.
func (c *Component) AddPeriod(name string, value Period) *Component {
	return c.AddProperty(Property{Name: name, Params: Params{}.Set(ParamValue, ValuePeriod), Value: value.String()})
}

// AddPeriods appends a comma-separated PERIOD property with VALUE=PERIOD.
func (c *Component) AddPeriods(name string, values ...Period) *Component {
	encoded := make([]string, len(values))
	for i, value := range values {
		encoded[i] = value.String()
	}
	return c.AddProperty(Property{Name: name, Params: Params{}.Set(ParamValue, ValuePeriod), Value: strings.Join(encoded, ",")})
}

// AddRecurrenceRule appends an RRULE property.
func (c *Component) AddRecurrenceRule(rule RecurrenceRule) *Component {
	return c.AddProperty(Property{Name: PropRRule, Value: rule.String()})
}

// AddComponent appends a child component.
func (c *Component) AddComponent(child *Component) *Component {
	c.Components = append(c.Components, child)
	return c
}

// PropertiesByName returns properties with the given case-insensitive name.
func (c *Component) PropertiesByName(name string) []Property {
	name = strings.ToUpper(name)
	var out []Property
	for _, p := range c.Properties {
		if p.Name == name {
			out = append(out, p)
		}
	}
	return out
}

// ComponentsByName recursively returns components with the given name.
func (c *Component) ComponentsByName(name string) []*Component {
	name = strings.ToUpper(name)
	var out []*Component
	var walk func(*Component)
	walk = func(cur *Component) {
		if cur.Name == name {
			out = append(out, cur)
		}
		for _, child := range cur.Components {
			walk(child)
		}
	}
	for _, child := range c.Components {
		walk(child)
	}
	return out
}

// Events returns all VEVENT components below this component.
func (c *Component) Events() []*Component {
	return c.ComponentsByName(CompVEvent)
}

// Todos returns all VTODO components below this component.
func (c *Component) Todos() []*Component {
	return c.ComponentsByName(CompVTodo)
}

// Journals returns all VJOURNAL components below this component.
func (c *Component) Journals() []*Component {
	return c.ComponentsByName(CompVJournal)
}

// FreeBusy returns all VFREEBUSY components below this component.
func (c *Component) FreeBusy() []*Component {
	return c.ComponentsByName(CompVFreeBusy)
}

// Timezones returns all VTIMEZONE components below this component.
func (c *Component) Timezones() []*Component {
	return c.ComponentsByName(CompVTimezone)
}

// Standard returns the STANDARD observance components below this VTIMEZONE.
func (c *Component) Standard() []*Component {
	return c.ComponentsByName(CompStandard)
}

// Daylight returns the DAYLIGHT observance components below this VTIMEZONE.
func (c *Component) Daylight() []*Component {
	return c.ComponentsByName(CompDaylight)
}

// Text returns the first decoded TEXT-like value for a property.
// It returns an empty string when the property is absent.
func (c *Component) Text(name string) string {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return ""
	}
	return DecodeText(props[0].Value)
}

// Boolean returns the first BOOLEAN value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Boolean(name string) (bool, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return false, false
	}
	value, err := ParseBoolean(props[0].Value)
	return value, err == nil
}

// Integer returns the first INTEGER value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Integer(name string) (int, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return 0, false
	}
	value, err := ParseInteger(props[0].Value)
	return value, err == nil
}

// Float returns the first FLOAT value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Float(name string) (float64, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return 0, false
	}
	value, err := ParseFloat(props[0].Value)
	return value, err == nil
}

// Geo returns the GEO value for a component.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Geo() (Geo, bool) {
	props := c.PropertiesByName(PropGeo)
	if len(props) == 0 {
		return Geo{}, false
	}
	value, err := ParseGeo(props[0].Value)
	return value, err == nil
}

// URI returns the first URI-like value for a property.
// It returns an empty string when the property is absent.
func (c *Component) URI(name string) string {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return ""
	}
	return props[0].Value
}

// Time returns the first TIME value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Time(name string) (Time, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return Time{}, false
	}
	value, err := ParseTime(props[0].Value)
	return value, err == nil
}

// Binary returns the first BINARY value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Binary(name string) ([]byte, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return nil, false
	}
	value, err := ParseBinary(props[0].Value)
	return value, err == nil
}

// DateTime returns the first DATE-TIME value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) DateTime(name string) (time.Time, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return time.Time{}, false
	}
	t, err := parseDateTimeProperty(props[0])
	return t, err == nil
}

// DateTimes returns all DATE-TIME or DATE values for matching properties.
// It handles comma-separated multi-value properties.
// The second result reports whether at least one property was found and all values
// parsed successfully.
func (c *Component) DateTimes(name string) ([]time.Time, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return nil, false
	}
	var values []time.Time
	for _, prop := range props {
		for _, raw := range strings.Split(prop.Value, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			var (
				t   time.Time
				err error
			)
			if prop.Params.Get(ParamValue) == ValueDate {
				t, err = time.Parse("20060102", raw)
			} else {
				t, err = parseDateTimeProperty(Property{
					Name:   prop.Name,
					Params: prop.Params,
					Value:  raw,
				})
			}
			if err != nil {
				return nil, false
			}
			values = append(values, t)
		}
	}
	return values, true
}

func parseDateTimeProperty(prop Property) (time.Time, error) {
	tzid := prop.Params.Get(ParamTZID)
	if tzid == "" || strings.HasSuffix(prop.Value, "Z") {
		return ParseDateTime(prop.Value)
	}
	location, err := time.LoadLocation(tzid)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation("20060102T150405", prop.Value, location)
}

// Date returns the first DATE value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Date(name string) (time.Time, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return time.Time{}, false
	}
	t, err := time.Parse("20060102", props[0].Value)
	return t, err == nil
}

// Duration returns the first DURATION value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Duration(name string) (time.Duration, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return 0, false
	}
	d, err := ParseDuration(props[0].Value)
	return d, err == nil
}

// UTCOffset returns the first UTC-OFFSET value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) UTCOffset(name string) (time.Duration, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return 0, false
	}
	d, err := ParseUTCOffset(props[0].Value)
	return d, err == nil
}

// Period returns the first PERIOD value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Period(name string) (Period, bool) {
	periods, ok := c.Periods(name)
	if !ok || len(periods) == 0 {
		return Period{}, false
	}
	return periods[0], true
}

// Periods returns all comma-separated PERIOD values from the first matching property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) Periods(name string) ([]Period, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return nil, false
	}
	periods, err := ParsePeriodList(props[0].Value)
	return periods, err == nil
}

// RecurrenceRule returns the first RRULE value for a property.
// The second result reports whether the property was present and successfully parsed.
func (c *Component) RecurrenceRule(name string) (RecurrenceRule, bool) {
	props := c.PropertiesByName(name)
	if len(props) == 0 {
		return RecurrenceRule{}, false
	}
	rule, err := ParseRecurrenceRule(props[0].Value)
	return rule, err == nil
}

// String serializes the component as RFC 5545 text with CRLF endings.
func (c *Component) String() string {
	var b strings.Builder
	c.writeTo(&b)
	return b.String()
}

func (c *Component) writeTo(b *strings.Builder) {
	for _, line := range []string{
		fmt.Sprintf("BEGIN:%s", strings.ToUpper(c.Name)),
	} {
		b.WriteString(foldLine(line))
	}
	for _, p := range c.Properties {
		b.WriteString(foldLine(p.contentLine()))
	}
	for _, child := range c.Components {
		child.writeTo(b)
	}
	b.WriteString(foldLine(fmt.Sprintf("END:%s", strings.ToUpper(c.Name))))
}

const contentLineLimit = 75

func foldLine(line string) string {
	if len([]byte(line)) <= contentLineLimit {
		return line + "\r\n"
	}
	var b strings.Builder
	remaining := line
	limit := contentLineLimit
	for len([]byte(remaining)) > limit {
		n := splitUTF8AtByteLimit(remaining, limit)
		b.WriteString(remaining[:n])
		b.WriteString("\r\n ")
		remaining = remaining[n:]
		limit = contentLineLimit - 1
	}
	b.WriteString(remaining)
	b.WriteString("\r\n")
	return b.String()
}

func splitUTF8AtByteLimit(s string, limit int) int {
	if len(s) <= limit {
		return len(s)
	}
	n := 0
	for i, r := range s {
		size := utf8.RuneLen(r)
		if n+size > limit {
			return i
		}
		n += size
	}
	return len(s)
}
