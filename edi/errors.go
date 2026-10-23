package edi

import "fmt"

// FieldLengthError reports an element that does not fit its declared width.
type FieldLengthError struct {
	// Field is the X12 element identifier, e.g. "NM103".
	Field string
	// Kind is the padding class of the element.
	Kind Kind
	// Min is the declared minimum when the error comes from Validate, else 0.
	Min int
	// Max is the declared width when the error comes from encoding, else 0.
	Max int
	// Got is the observed length in runes.
	Got int
}

func (e *FieldLengthError) Error() string {
	if e.Max > 0 {
		return fmt.Sprintf("edi: element %s exceeds its width of %d (got %d)", e.Field, e.Max, e.Got)
	}
	return fmt.Sprintf("edi: element %s is shorter than its minimum of %d (got %d)", e.Field, e.Min, e.Got)
}

// ParseError reports a segment that could not be interpreted.
type ParseError struct {
	Segment string
	Reason  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("edi: cannot parse %s: %s", e.Segment, e.Reason)
}
