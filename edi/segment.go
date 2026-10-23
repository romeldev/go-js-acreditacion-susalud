// Package edi implements the fixed-width ANSI ASC X12 segment encoding used by
// the SUSALUD transaction set, plus the separator-based decoder.
//
// The wire format is a hybrid: elements are separated by '*' and segments are
// terminated by '~', but every element is also padded to its declared width —
// numeric elements with leading '0', alphanumeric elements with trailing
// spaces. A receiver built against the original Java library depends on that
// padding, so the encoder reproduces it exactly.
//
// Decoding deliberately does not rely on the padding: it splits on the
// separators and trims. The reference implementation's positional decoder
// existed but was never called.
package edi

import (
	"strings"
	"unicode/utf8"
)

// Kind is the padding class of an element.
type Kind uint8

const (
	// KindAlpha is right-padded with spaces.
	KindAlpha Kind = iota
	// KindNumeric is left-padded with zeros.
	KindNumeric
	// KindRaw receives no padding and is emitted as-is.
	KindRaw
)

func (k Kind) String() string {
	switch k {
	case KindAlpha:
		return "alpha"
	case KindNumeric:
		return "numeric"
	default:
		return "raw"
	}
}

// Separators of the SUSALUD transaction set.
const (
	ElementSep   = "*"
	SegmentTerm  = "~"
	ComponentSep = ":"
)

// FieldDef describes one element of a segment.
type FieldDef struct {
	// Name is the X12 element identifier, e.g. "NM101".
	Name string
	// Kind selects the padding rule.
	Kind Kind
	// Min is the minimum declared length. Only checked by Validate.
	Min int
	// Max is the declared width. The encoder pads to exactly Max.
	Max int
	// Desc is the element description from the reference specification.
	Desc string
	// Default is the literal value the reference implementation starts from.
	Default string
}

// SegmentDef describes the layout of one segment type.
type SegmentDef struct {
	Name   string
	Prefix string
	Suffix string
	Fields []FieldDef
}

// Width returns the total encoded width of the segment body, separators
// excluded. This is the length the reference implementation asserts on input.
func (d *SegmentDef) Width() int {
	n := 0
	for _, f := range d.Fields {
		n += f.Max
	}
	return n
}

// Order returns the 1-based position of the named element, or 0 if absent.
func (d *SegmentDef) Order(name string) int {
	for i, f := range d.Fields {
		if f.Name == name {
			return i + 1
		}
	}
	return 0
}

// Segment is a live instance of a segment type.
//
// A Segment is not safe for concurrent use; create one per message. Because
// the type carries no package-level state, separate goroutines can marshal
// concurrently.
type Segment struct {
	Def    *SegmentDef
	Values []string
}

// New returns a segment initialised with the definition's literal defaults.
func New(def *SegmentDef) *Segment {
	s := &Segment{Def: def, Values: make([]string, len(def.Fields))}
	for i, f := range def.Fields {
		s.Values[i] = f.Default
	}
	return s
}

// NewNamed returns a fresh instance of the segment with the given X12 ID, such
// as "NM1". It panics if the segment is not defined; segment IDs are compile
// time constants of the library, so an unknown ID is a programming error.
func NewNamed(id string) *Segment {
	def, ok := Segments[id]
	if !ok {
		panic("edi: unknown segment " + id)
	}
	return New(def)
}

// Set writes the element at the given 1-based order and returns the segment for
// chaining. An order outside the definition panics: element positions are
// constants of the mapping, not data.
func (s *Segment) Set(order int, v string) *Segment {
	if order < 1 || order > len(s.Values) {
		panic("edi: element order out of range for " + s.Def.Name)
	}
	s.Values[order-1] = v
	return s
}

// SetName writes the element with the given X12 name, e.g. "NM108". It reports
// whether the element exists.
func (s *Segment) SetName(name, v string) bool {
	order := s.Def.Order(name)
	if order == 0 {
		return false
	}
	s.Values[order-1] = v
	return true
}

// Get returns the raw, unpadded value at the 1-based order.
func (s *Segment) Get(order int) string {
	if order < 1 || order > len(s.Values) {
		return ""
	}
	return s.Values[order-1]
}

// Encode renders the segment with the definition's default separators.
func (s *Segment) Encode() (string, error) {
	return s.EncodeWith(s.Def.Prefix, ElementSep, s.Def.Suffix)
}

// EncodeWith renders the segment using the given prefix, element separator and
// suffix. Composite sub-elements are rendered by passing an empty prefix and
// the component separator, which reproduces how the reference implementation
// nests REF04 and EB13 style components inside a parent element.
func (s *Segment) EncodeWith(prefix, sep, suffix string) (string, error) {
	var b strings.Builder
	b.Grow(len(prefix) + s.Def.Width() + len(sep)*len(s.Values) + len(suffix))
	b.WriteString(prefix)
	for i, f := range s.Def.Fields {
		if i > 0 {
			b.WriteString(sep)
		}
		v, err := pad(s.Values[i], f)
		if err != nil {
			return "", err
		}
		b.WriteString(v)
	}
	b.WriteString(suffix)
	return b.String(), nil
}

// pad applies the element's padding class.
//
// Unlike the reference implementation, a value wider than the declared maximum
// is an error rather than being silently replaced with "0" or "". Silent
// substitution both lost the value and produced a segment shorter than its
// declared width, which desynchronised any width-aware reader.
func pad(v string, f FieldDef) (string, error) {
	n := utf8.RuneCountInString(v)
	switch f.Kind {
	case KindNumeric, KindAlpha:
		if n > f.Max {
			return "", &FieldLengthError{Field: f.Name, Kind: f.Kind, Max: f.Max, Got: n}
		}
		if n == f.Max {
			return v, nil
		}
		if f.Kind == KindNumeric {
			return strings.Repeat("0", f.Max-n) + v, nil
		}
		return v + strings.Repeat(" ", f.Max-n), nil
	default:
		return v, nil
	}
}

// Validate checks the declared minimum lengths. The encoder does not call it;
// it exists so callers can reject incomplete messages before they reach the
// wire. Empty optional elements are skipped.
func (s *Segment) Validate() error {
	for i, f := range s.Def.Fields {
		if f.Min <= 0 {
			continue
		}
		v := strings.TrimSpace(s.Values[i])
		if v == "" {
			continue
		}
		if n := utf8.RuneCountInString(v); n < f.Min {
			return &FieldLengthError{Field: f.Name, Kind: f.Kind, Min: f.Min, Got: n}
		}
	}
	return nil
}

// ParseSegment splits a raw segment such as "NM1*IL*1*DIAZ~" into its segment
// ID and its elements, trimming the padding from every element. The trailing
// segment terminator, if present, is removed.
func ParseSegment(raw string) (id string, elems []string) {
	raw = strings.TrimSuffix(strings.TrimSpace(raw), SegmentTerm)
	parts := strings.Split(raw, ElementSep)
	id = strings.TrimSpace(parts[0])
	elems = make([]string, 0, len(parts)-1)
	for _, p := range parts[1:] {
		elems = append(elems, strings.TrimSpace(p))
	}
	return id, elems
}

// SplitSegments splits a full interchange into its raw segments.
func SplitSegments(s string) []string {
	var out []string
	for _, part := range strings.Split(s, SegmentTerm) {
		if strings.TrimSpace(part) == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

// elem returns elems[i-1], or "" when the sender omitted a trailing element.
func elem(elems []string, order int) string {
	if order < 1 || order > len(elems) {
		return ""
	}
	return elems[order-1]
}
