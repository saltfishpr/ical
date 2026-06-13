package ical

// Standard iCalendar component names defined by RFC 5545 Section 3.6.
const (
	CompVCalendar = "VCALENDAR"
	CompVEvent    = "VEVENT"
	CompVTodo     = "VTODO"
	CompVJournal  = "VJOURNAL"
	CompVFreeBusy = "VFREEBUSY"
	CompVTimezone = "VTIMEZONE"
	CompVAlarm    = "VALARM"
	CompStandard  = "STANDARD"
	CompDaylight  = "DAYLIGHT"
)

// Standard iCalendar property names defined by RFC 5545.
const (
	PropAction          = "ACTION"
	PropAttach          = "ATTACH"
	PropAttendee        = "ATTENDEE"
	PropCalScale        = "CALSCALE"
	PropCategories      = "CATEGORIES"
	PropClass           = "CLASS"
	PropComment         = "COMMENT"
	PropCompleted       = "COMPLETED"
	PropContact         = "CONTACT"
	PropCreated         = "CREATED"
	PropDescription     = "DESCRIPTION"
	PropDTEnd           = "DTEND"
	PropDTStamp         = "DTSTAMP"
	PropDTStart         = "DTSTART"
	PropDue             = "DUE"
	PropDuration        = "DURATION"
	PropExDate          = "EXDATE"
	PropFreeBusy        = "FREEBUSY"
	PropGeo             = "GEO"
	PropLastModified    = "LAST-MODIFIED"
	PropLocation        = "LOCATION"
	PropMethod          = "METHOD"
	PropOrganizer       = "ORGANIZER"
	PropPercentComplete = "PERCENT-COMPLETE"
	PropPriority        = "PRIORITY"
	PropProdID          = "PRODID"
	PropRDate           = "RDATE"
	PropRecurrenceID    = "RECURRENCE-ID"
	PropRelatedTo       = "RELATED-TO"
	PropRepeat          = "REPEAT"
	PropRequestStatus   = "REQUEST-STATUS"
	PropResources       = "RESOURCES"
	PropRRule           = "RRULE"
	PropSequence        = "SEQUENCE"
	PropStatus          = "STATUS"
	PropSummary         = "SUMMARY"
	PropTransp          = "TRANSP"
	PropTrigger         = "TRIGGER"
	PropTZID            = "TZID"
	PropTZName          = "TZNAME"
	PropTZOffsetFrom    = "TZOFFSETFROM"
	PropTZOffsetTo      = "TZOFFSETTO"
	PropTZURL           = "TZURL"
	PropUID             = "UID"
	PropURL             = "URL"
	PropVersion         = "VERSION"
)

// Common parameter names defined by RFC 5545 Section 3.2.
const (
	ParamAltRep        = "ALTREP"
	ParamCN            = "CN"
	ParamCUTYPE        = "CUTYPE"
	ParamDelegatedFrom = "DELEGATED-FROM"
	ParamDelegatedTo   = "DELEGATED-TO"
	ParamDir           = "DIR"
	ParamEncoding      = "ENCODING"
	ParamFBType        = "FBTYPE"
	ParamFmtType       = "FMTTYPE"
	ParamLanguage      = "LANGUAGE"
	ParamMember        = "MEMBER"
	ParamPartStat      = "PARTSTAT"
	ParamRange         = "RANGE"
	ParamRelated       = "RELATED"
	ParamRelType       = "RELTYPE"
	ParamRole          = "ROLE"
	ParamRSVP          = "RSVP"
	ParamSentBy        = "SENT-BY"
	ParamTZID          = "TZID"
	ParamValue         = "VALUE"
)

// Well-known parameter values defined by RFC 5545.
const (
	ValueTime   = "TIME"
	ValueDate   = "DATE"
	ValueBinary = "BINARY"
	ValueBase64 = "BASE64"
	ValuePeriod = "PERIOD"
)

// Internal parser tokens marking the start and end of RFC 5545 components.
const (
	TokenBegin = "BEGIN"
	TokenEnd   = "END"
)

// VALARM ACTION property values defined by RFC 5545 Section 3.6.6.
const (
	ActionAudio   = "AUDIO"
	ActionDisplay = "DISPLAY"
	ActionEmail   = "EMAIL"
)
