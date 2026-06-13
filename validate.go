package ical

import "fmt"

// ValidationError describes one RFC 5545 component validation failure detected
// by [Component.Validate].
//
// Fields:
//   - Component: the component name where the failure occurred.
//   - Property: the property name involved, or empty for nesting violations.
//   - Rule: the conformance rule that was violated (required, singleton,
//     exclusive, inclusive, nesting, required-child, action-required).
//   - Message: a human-readable description of the failure.
type ValidationError struct {
	Component string
	Property  string
	Rule      string
	Message   string
}

// Error returns a readable validation error.
func (e ValidationError) Error() string {
	return e.Message
}

type Validator interface {
	Validate(c *Component) []ValidationError
}

var validatorsByComponent = map[string][]Validator{}

func RegisterValidator(component string, v Validator) {
	validatorsByComponent[component] = append(validatorsByComponent[component], v)
}

// componentRules encodes RFC 5545 property constraints for a component type.
//
// Each field maps to a specific conformance rule:
//   - required: properties that MUST appear at least once (e.g., UID, DTSTAMP).
//   - singletons: properties that MUST NOT appear more than once (e.g., DTSTART).
//   - exclusive: mutually exclusive property pairs — at most one may appear (e.g., DTEND vs DURATION).
//   - inclusive: property pairs that MUST appear together — if one is present, the other is required
//     (e.g., DURATION and REPEAT in VALARM).
type componentRules struct {
	required   []string
	singletons []string
	exclusive  [][2]string
	inclusive  [][2]string
}

var rulesByComponent = map[string]componentRules{
	CompVCalendar: {
		required:   []string{PropProdID, PropVersion},
		singletons: []string{PropProdID, PropVersion, PropCalScale, PropMethod},
	},
	CompVEvent: {
		required:   []string{PropUID, PropDTStamp},
		singletons: []string{PropClass, PropCreated, PropDescription, PropDTEnd, PropDTStart, PropDuration, PropGeo, PropLastModified, PropLocation, PropOrganizer, PropPriority, PropDTStamp, PropSequence, PropStatus, PropSummary, PropTransp, PropUID, PropURL, PropRecurrenceID},
		exclusive:  [][2]string{{PropDTEnd, PropDuration}},
	},
	CompVTodo: {
		required:   []string{PropUID, PropDTStamp},
		singletons: []string{PropClass, PropCompleted, PropCreated, PropDescription, PropDTStamp, PropDTStart, PropDue, PropDuration, PropGeo, PropLastModified, PropLocation, PropOrganizer, PropPercentComplete, PropPriority, PropRecurrenceID, PropSequence, PropStatus, PropSummary, PropUID, PropURL},
		exclusive:  [][2]string{{PropDue, PropDuration}},
	},
	CompVJournal: {
		required:   []string{PropUID, PropDTStamp},
		singletons: []string{PropClass, PropCreated, PropDescription, PropDTStamp, PropDTStart, PropLastModified, PropOrganizer, PropRecurrenceID, PropSequence, PropStatus, PropSummary, PropUID, PropURL},
	},
	CompVFreeBusy: {
		required:   []string{PropUID, PropDTStamp},
		singletons: []string{PropContact, PropDTEnd, PropDTStamp, PropDTStart, PropOrganizer, PropUID, PropURL},
	},
	CompVTimezone: {
		required:   []string{PropTZID},
		singletons: []string{PropTZID, PropLastModified, PropTZURL},
	},
	CompStandard: {
		required:   []string{PropDTStart, PropTZOffsetFrom, PropTZOffsetTo},
		singletons: []string{PropDTStart, PropTZOffsetFrom, PropTZOffsetTo},
	},
	CompDaylight: {
		required:   []string{PropDTStart, PropTZOffsetFrom, PropTZOffsetTo},
		singletons: []string{PropDTStart, PropTZOffsetFrom, PropTZOffsetTo},
	},
	CompVAlarm: {
		required:   []string{PropAction, PropTrigger},
		singletons: []string{PropAttach, PropAction, PropDescription, PropDuration, PropRepeat, PropSummary, PropTrigger},
		inclusive:  [][2]string{{PropDuration, PropRepeat}, {PropSummary, PropAttendee}},
	},
}

// Validate checks the component tree against core RFC 5545 conformance rules:
// required properties, singleton enforcement, mutual exclusion (e.g., DTEND
// vs DURATION), property co-occurrence (e.g., DURATION/REPEAT in VALARM),
// component nesting, and alarm action-specific requirements.
// It returns nil when no violations are found.
func (c *Component) Validate() []ValidationError {
	return c.validateInto("")
}

func (c *Component) validateInto(parent string) []ValidationError {
	var errs []ValidationError

	if !validChild(parent, c.Name) {
		errs = append(errs, ValidationError{
			Component: c.Name,
			Rule:      "nesting",
			Message:   fmt.Sprintf("%s cannot be nested in %s", c.Name, parent),
		})
	}

	if rules, ok := rulesByComponent[c.Name]; ok {
		counts := c.propertyCounts()
		errs = append(errs, c.checkRequired(rules.required, counts)...)
		errs = append(errs, c.checkSingletons(rules.singletons, counts)...)
		errs = append(errs, c.checkExclusive(rules.exclusive, counts)...)
		errs = append(errs, c.checkInclusive(rules.inclusive, counts)...)
	}

	for _, v := range validatorsByComponent[c.Name] {
		errs = append(errs, v.Validate(c)...)
	}

	for _, child := range c.Components {
		errs = append(errs, child.validateInto(c.Name)...)
	}
	return errs
}

// propertyCounts returns the occurrence count of each property name in the component.
func (c *Component) propertyCounts() map[string]int {
	counts := make(map[string]int)
	for _, p := range c.Properties {
		counts[p.Name]++
	}
	return counts
}

func (c *Component) checkRequired(required []string, counts map[string]int) []ValidationError {
	var errs []ValidationError
	for _, name := range required {
		if counts[name] == 0 {
			errs = append(errs, ValidationError{
				Component: c.Name,
				Property:  name,
				Rule:      "required",
				Message:   fmt.Sprintf("%s requires %s", c.Name, name),
			})
		}
	}
	return errs
}

func (c *Component) checkSingletons(singletons []string, counts map[string]int) []ValidationError {
	var errs []ValidationError
	for _, name := range singletons {
		if counts[name] > 1 {
			errs = append(errs, ValidationError{
				Component: c.Name,
				Property:  name,
				Rule:      "singleton",
				Message:   fmt.Sprintf("%s allows only one %s", c.Name, name),
			})
		}
	}
	return errs
}

func (c *Component) checkExclusive(exclusive [][2]string, counts map[string]int) []ValidationError {
	var errs []ValidationError
	for _, pair := range exclusive {
		if counts[pair[0]] > 0 && counts[pair[1]] > 0 {
			errs = append(errs, ValidationError{
				Component: c.Name,
				Property:  pair[0] + "/" + pair[1],
				Rule:      "exclusive",
				Message:   fmt.Sprintf("%s cannot contain both %s and %s", c.Name, pair[0], pair[1]),
			})
		}
	}
	return errs
}

func (c *Component) checkInclusive(inclusive [][2]string, counts map[string]int) []ValidationError {
	var errs []ValidationError
	for _, pair := range inclusive {
		if counts[pair[0]] > 0 && counts[pair[1]] == 0 {
			errs = append(errs, ValidationError{
				Component: c.Name,
				Property:  pair[0] + "/" + pair[1],
				Rule:      "inclusive",
				Message:   fmt.Sprintf("%s requires %s when %s is present", c.Name, pair[1], pair[0]),
			})
		}
		if counts[pair[1]] > 0 && counts[pair[0]] == 0 {
			errs = append(errs, ValidationError{
				Component: c.Name,
				Property:  pair[0] + "/" + pair[1],
				Rule:      "inclusive",
				Message:   fmt.Sprintf("%s requires %s when %s is present", c.Name, pair[0], pair[1]),
			})
		}
	}
	return errs
}

func validChild(parent, child string) bool {
	if parent == "" {
		return true
	}
	switch parent {
	case CompVCalendar:
		return child == CompVEvent || child == CompVTodo || child == CompVJournal || child == CompVFreeBusy || child == CompVTimezone || isExtensionComponent(child)
	case CompVEvent, CompVTodo:
		return child == CompVAlarm || isExtensionComponent(child)
	case CompVTimezone:
		return child == CompStandard || child == CompDaylight || isExtensionComponent(child)
	default:
		return isExtensionComponent(child)
	}
}

func isExtensionComponent(name string) bool {
	return len(name) > 2 && name[:2] == "X-"
}
