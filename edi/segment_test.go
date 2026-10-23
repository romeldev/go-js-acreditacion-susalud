package edi

import (
	"strings"
	"testing"
)

func TestPadNumericIsLeftFilled(t *testing.T) {
	seg := New(SegISA)
	seg.Set(1, "7") // ISA01, numeric, width 2
	got, err := seg.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "ISA*07*") {
		t.Fatalf("numeric element not left-filled with zeros: %q", got)
	}
}

func TestPadAlphaIsRightFilled(t *testing.T) {
	seg := New(SegNM1)
	seg.Set(1, "IL") // width 3
	seg.Set(8, "MI") // width 2
	got, err := seg.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "NM1*IL *") {
		t.Fatalf("alpha element not right-filled with spaces: %q", got)
	}
}

func TestOverflowIsAnError(t *testing.T) {
	seg := New(SegNM1)
	seg.Set(1, "TOOLONG") // width 3
	if _, err := seg.Encode(); err == nil {
		t.Fatal("encoding an over-wide element must fail")
	} else if _, ok := err.(*FieldLengthError); !ok {
		t.Fatalf("want *FieldLengthError, got %T", err)
	}
}

func TestRawElementsAreNotPadded(t *testing.T) {
	seg := New(SegISA)
	// ISA12 and ISA16 are declared with no padding class.
	got, err := seg.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "*00501*") {
		t.Fatalf("raw element ISA12 should keep its literal width: %q", got)
	}
	if !strings.HasSuffix(got, ":*~") && !strings.HasSuffix(got, ":~") {
		t.Fatalf("raw element ISA16 should not be padded: %q", got)
	}
}

func TestParseSegmentTrimsPadding(t *testing.T) {
	id, elems := ParseSegment("NM1*IL *1*DIAZ      *ANA                 ~")
	if id != "NM1" {
		t.Fatalf("id = %q", id)
	}
	if elems[0] != "IL" || elems[1] != "1" || elems[2] != "DIAZ" || elems[3] != "ANA" {
		t.Fatalf("elements not trimmed: %#v", elems)
	}
}

func TestParseSegmentToleratesMissingElements(t *testing.T) {
	_, elems := ParseSegment("EB*1~")
	if len(elems) != 1 || elems[0] != "1" {
		t.Fatalf("unexpected elements: %#v", elems)
	}
}

func TestNewControlNumber(t *testing.T) {
	for _, digits := range []int{1, 8, 9, 18} {
		got, err := NewControlNumber(digits)
		if err != nil {
			t.Fatalf("%d digits: %v", digits, err)
		}
		if len(got) != digits {
			t.Fatalf("%d digits: got %q of length %d", digits, got, len(got))
		}
		if got[0] == '0' {
			t.Fatalf("%d digits: %q has a leading zero", digits, got)
		}
	}
	if _, err := NewControlNumber(0); err == nil {
		t.Fatal("zero digits must fail")
	}
}

func TestBuildISA(t *testing.T) {
	seg, err := BuildISA(ISA{
		SenderID:    "980001C",
		ReceiverID:  "20001",
		Date:        "20261022",
		Time:        "121037",
		Correlative: "026483112",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := seg.Encode()
	if err != nil {
		t.Fatal(err)
	}
	// ISA09 carries the last 6 digits of the date, ISA10 the first 4 of the time.
	if !strings.Contains(got, "*261022*1210*") {
		t.Fatalf("date/time truncation wrong: %q", got)
	}
	// Test indicator is the default, matching the reference implementation.
	if !strings.Contains(got, "*0*T*:") {
		t.Fatalf("expected test indicator: %q", got)
	}
}

func TestBuildISAProductionFlag(t *testing.T) {
	seg, err := BuildISA(ISA{
		SenderID: "A", ReceiverID: "B", Date: "20261022", Time: "1210",
		Correlative: "000000001", Production: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := seg.Encode()
	if !strings.Contains(got, "*0*P*:") {
		t.Fatalf("expected production indicator: %q", got)
	}
}

func TestBuildISARejectsShortDate(t *testing.T) {
	if _, err := BuildISA(ISA{Date: "2026", Time: "1210", Correlative: "000000001"}); err == nil {
		t.Fatal("short date must fail")
	}
}
