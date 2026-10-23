package edi

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// NewControlNumber returns a random decimal string of exactly digits digits
// with no leading zero, suitable for ISA13/GS06 and the other control fields.
//
// The reference implementation reseeded java.util.Random on every call and
// retried until the digit count matched. Those numbers collided by the
// birthday bound after roughly 37k messages, which breaks the interchange and
// group correlation the receiver relies on. Randomness here comes from
// crypto/rand and the range is uniform over [10^(digits-1), 10^digits).
func NewControlNumber(digits int) (string, error) {
	if digits < 1 || digits > 18 {
		return "", fmt.Errorf("edi: control number must have 1..18 digits, got %d", digits)
	}
	lo := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits-1)), nil)
	hi := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	hi.Sub(hi, lo)

	n, err := rand.Int(rand.Reader, hi)
	if err != nil {
		return "", fmt.Errorf("edi: control number: %w", err)
	}
	return n.Add(n, lo).String(), nil
}

// ISA carries the interchange header parameters.
type ISA struct {
	// SenderID is ISA06, maximum 15 characters, padded with spaces.
	SenderID string
	// ReceiverID is ISA08, maximum 15 characters, padded with spaces.
	ReceiverID string
	// Date is the interchange date as yyyyMMdd. ISA09 carries its last 6
	// digits, matching the reference implementation.
	Date string
	// Time is the interchange time; ISA10 carries its first 4 characters.
	Time string
	// Correlative is ISA13, 9 digits.
	Correlative string
	// Production selects ISA15. When false the interchange is marked "T"
	// (test), which is what the reference implementation always emitted.
	Production bool
}

// BuildISA renders the ISA segment.
func BuildISA(o ISA) (*Segment, error) {
	if len(o.Date) != 8 {
		return nil, &ParseError{Segment: "ISA", Reason: "date must be yyyyMMdd"}
	}
	if len(o.Time) < 4 {
		return nil, &ParseError{Segment: "ISA", Reason: "time must have at least 4 characters"}
	}
	if len(o.Correlative) != 9 {
		return nil, &ParseError{Segment: "ISA", Reason: "correlative must be 9 digits"}
	}

	seg := New(SegISA)
	seg.Set(6, o.SenderID)
	seg.Set(8, o.ReceiverID)
	seg.Set(9, o.Date[2:])
	seg.Set(10, o.Time[:4])
	seg.Set(13, o.Correlative)
	seg.Set(15, "T")
	if o.Production {
		seg.Set(15, "P")
	}
	return seg, nil
}

// Now returns the current local date as yyyyMMdd and time as HHmmss.
func Now() (date, clock string) {
	t := time.Now()
	return t.Format("20060102"), t.Format("150405")
}
