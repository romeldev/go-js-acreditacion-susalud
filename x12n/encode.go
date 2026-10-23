package x12n

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// Options tunes marshalling. The zero value reproduces the reference
// implementation's layout and rejects elements that do not fit their declared
// width.
type Options struct {
	// Control overrides the interchange control number carried by ISA13,
	// GS06 and GE02. When empty a random 8-digit number is generated.
	Control string
	// ControlST overrides the transaction set control number carried by ST02
	// and SE02. When empty a random 8-digit number is generated.
	ControlST string
	// Production marks the interchange as production (ISA15 = "P"). The
	// reference implementation always emitted "T".
	Production bool
	// Compat reproduces the reference implementation's handling of elements
	// that exceed their declared width: instead of returning an error it
	// blanks the element ("" for alphanumeric, "0" for numeric) and emits the
	// segment shorter than its declared width.
	//
	// That behaviour silently discards data and desynchronises any width-aware
	// reader, so it is off by default. Enable it only when a peer requires
	// byte-identical output for input the reference accepted.
	Compat bool
}

// control resolves the interchange control number, storing it on both the
// options and the caller-visible value so a caller can log it.
func (o *Options) control() (string, error) {
	if o.Control == "" {
		n, err := edi.NewControlNumber(8)
		if err != nil {
			return "", err
		}
		o.Control = n
	}
	return o.Control, nil
}

func (o *Options) controlST() (string, error) {
	if o.ControlST == "" {
		n, err := edi.NewControlNumber(8)
		if err != nil {
			return "", err
		}
		o.ControlST = n
	}
	return o.ControlST, nil
}

// encoder accumulates the segments of one transaction.
type encoder struct {
	sb     strings.Builder
	err    error
	compat bool
}

func (e *encoder) fail(err error) {
	if e.err == nil {
		e.err = err
	}
}

// render trims every element and encodes the segment. The reference
// implementation trims on every bean setter, so trimming here reproduces the
// same effective values without repeating it at each call site.
func (e *encoder) render(s *edi.Segment, prefix, sep, suffix string) (string, error) {
	values := make([]string, len(s.Values))
	for i, v := range s.Values {
		values[i] = strings.TrimSpace(v)
	}
	if e.compat {
		for i, f := range s.Def.Fields {
			if f.Max > 0 && utf8.RuneCountInString(values[i]) > f.Max {
				switch f.Kind {
				case edi.KindNumeric:
					values[i] = "0"
				case edi.KindAlpha:
					values[i] = ""
				}
			}
		}
	}
	return (&edi.Segment{Def: s.Def, Values: values}).EncodeWith(prefix, sep, suffix)
}

// add appends a segment rendered with an explicit prefix, separator and
// suffix. It mirrors the reference `returnComoString(prefix, sep, suffix)`.
func (e *encoder) add(prefix, sep, suffix string, s *edi.Segment) {
	if e.err != nil {
		return
	}
	out, err := e.render(s, prefix, sep, suffix)
	if err != nil {
		e.fail(err)
		return
	}
	e.sb.WriteString(out)
}

// addStd appends a segment with the default separators, e.g. "NM1*...~".
func (e *encoder) addStd(s *edi.Segment) {
	e.add(s.Def.Prefix, edi.ElementSep, edi.SegmentTerm, s)
}

// addNested appends a parent segment whose last element carries a composite
// sub-segment, rendered with the component separator and closed by "~".
// This is the REF04 / EB13 idiom of the reference implementation.
func (e *encoder) addNested(parent, child *edi.Segment) {
	if e.err != nil {
		return
	}
	sub, err := e.render(child, "", edi.ComponentSep, edi.SegmentTerm)
	if err != nil {
		e.fail(err)
		return
	}
	e.add(parent.Def.Prefix, edi.ElementSep, edi.ElementSep+sub, parent)
}

// header carries the interchange identifiers shared by every transaction.
type header struct {
	SenderID    string // ISA06 / GS02
	ReceiverID  string // ISA08 / GS03
	Date        string // yyyyMMdd, GS04
	Time        string // HHmmss, GS05
	Correlative string // ISA13 / IEA02
	Transaction string // ST01
}

// open emits the ISA, GS and ST segments.
func (e *encoder) open(h header, gsType, control, controlST, stCount string, production bool) error {
	if len(h.Date) < 2 {
		return &edi.ParseError{Segment: "ISA", Reason: "Date must be yyyyMMdd"}
	}
	if len(h.Time) < 4 {
		return &edi.ParseError{Segment: "ISA", Reason: "Time must have at least 4 characters"}
	}

	isa := edi.New(edi.SegISA).
		Set(6, h.SenderID).
		Set(8, h.ReceiverID).
		Set(9, h.Date[2:]).
		Set(10, h.Time[:4]).
		Set(13, h.Correlative)
	if production {
		isa.Set(15, "P")
	}
	e.addStd(isa)

	e.addStd(edi.New(edi.SegGS).
		Set(1, gsType).
		Set(2, h.SenderID).
		Set(3, h.ReceiverID).
		Set(4, h.Date).
		Set(5, h.Time).
		Set(6, control))

	e.addStd(edi.New(edi.SegST).
		Set(1, h.Transaction).
		Set(2, controlST).
		Set(3, stCount))
	return e.err
}

// close emits SE, GE and IEA. SE01 counts every segment written so far, which
// is what the reference implementation does: it excludes SE itself.
func (e *encoder) close(control, controlST, correlative string) error {
	e.addStd(edi.New(edi.SegSE).Set(1, strconv.Itoa(e.segments())).Set(2, controlST))
	e.addStd(edi.New(edi.SegGE).Set(2, control))
	e.addStd(edi.New(edi.SegIEA).Set(2, correlative))
	return e.err
}

func (e *encoder) string() string { return e.sb.String() }

// segments counts the segments written so far, used for SE01.
func (e *encoder) segments() int { return strings.Count(e.sb.String(), edi.SegmentTerm) }
