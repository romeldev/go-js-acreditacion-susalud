package x12n

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// fixture is one transaction captured from the reference Java library: the
// sample bean it built and the EDI it produced from it.
type fixture struct {
	Txn  string          `json:"txn"`
	Impl string          `json:"impl"`
	EDI  string          `json:"edi"`
	Bean json.RawMessage `json:"bean"`
}

func loadFixtures(t *testing.T) []fixture {
	t.Helper()
	raw, err := os.ReadFile("testdata/golden.json")
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var out []fixture
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("parse fixtures: %v", err)
	}
	return out
}

// beanControls reads the control numbers the reference implementation stored
// on the bean. They are the raw values, before the fixed-width padding, so
// replaying them reproduces the reference output exactly.
func beanControls(t *testing.T, raw json.RawMessage) (control, controlST string) {
	t.Helper()
	var probe struct {
		NuControl   string
		NuControlST string
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatalf("read control numbers: %v", err)
	}
	if probe.NuControl == "" || probe.NuControlST == "" {
		t.Fatalf("fixture carries no control numbers")
	}
	return probe.NuControl, probe.NuControlST
}

// marshals maps a transaction to its Go implementation. Transactions are added
// here as they are ported; the golden test skips the rest.
var marshals = map[string]func(json.RawMessage, *Options) (string, error){
	"SolAut271": func(raw json.RawMessage, o *Options) (string, error) {
		var b InSolAut271
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return Marshal271(&b, o)
	},
	"In997ResAut": func(raw json.RawMessage, o *Options) (string, error) {
		var b In997ResAut
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return Marshal997(&b, o)
	},
	"ConEntVinc278": func(raw json.RawMessage, o *Options) (string, error) {
		var b InConEntVinc278
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalConEntVinc278(&b, o)
	},
	"ResEntVinc278": func(raw json.RawMessage, o *Options) (string, error) {
		var b InResEntVinc278
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalResEntVinc278(&b, o)
	},
	"In278SolCG": func(raw json.RawMessage, o *Options) (string, error) {
		var b In278SolCG
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalIn278SolCG(&b, o)
	},
	"RegAfi270": func(raw json.RawMessage, o *Options) (string, error) {
		var b InRegAfi270
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalRegAfi270(&b, o)
	},
	"In271ConObs": func(raw json.RawMessage, o *Options) (string, error) {
		var b In271ConObs
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalIn271ConObs(&b, o)
	},
	"ConMed271": func(raw json.RawMessage, o *Options) (string, error) {
		var b InConMed271
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalConMed271(&b, o)
	},
	"In271ConDtad": func(raw json.RawMessage, o *Options) (string, error) {
		var b In271ConDtad
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalIn271ConDtad(&b, o)
	},
	"In278ResCG": func(raw json.RawMessage, o *Options) (string, error) {
		var b In278ResCG
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalIn278ResCG(&b, o)
	},
}

// TestGolden re-marshals every captured bean and requires byte-for-byte
// equality with the reference implementation's output.
func TestGolden(t *testing.T) {
	for _, f := range loadFixtures(t) {
		f := f
		t.Run(f.Txn, func(t *testing.T) {
			marshal, ok := marshals[f.Txn]
			if !ok {
				t.Skip("not ported yet")
			}
			control, controlST := beanControls(t, f.Bean)
			got, err := marshal(f.Bean, &Options{Control: control, ControlST: controlST, Compat: true})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if got == f.EDI {
				return
			}
			t.Fatalf("output differs from reference\n%s", firstDiff(f.EDI, got))
		})
	}
}

// roundtrips parses each captured interchange and re-encodes it. The result
// must reproduce the reference bytes exactly, which is what establishes that
// the parser captured everything the encoder emits.
var roundtrips = map[string]func(string, *Options) (string, error){
	"SolAut271": func(s string, o *Options) (string, error) {
		b, err := Unmarshal271(s)
		if err != nil {
			return "", err
		}
		return Marshal271(b, o)
	},
	"In997ResAut": func(s string, o *Options) (string, error) {
		b, err := Unmarshal997(s)
		if err != nil {
			return "", err
		}
		return Marshal997(b, o)
	},
	"ConEntVinc278": func(s string, o *Options) (string, error) {
		b, err := UnmarshalConEntVinc278(s)
		if err != nil {
			return "", err
		}
		return MarshalConEntVinc278(b, o)
	},
	"ResEntVinc278": func(s string, o *Options) (string, error) {
		b, err := UnmarshalResEntVinc278(s)
		if err != nil {
			return "", err
		}
		return MarshalResEntVinc278(b, o)
	},
	"In278SolCG": func(s string, o *Options) (string, error) {
		b, err := UnmarshalIn278SolCG(s)
		if err != nil {
			return "", err
		}
		return MarshalIn278SolCG(b, o)
	},
	"RegAfi270": func(s string, o *Options) (string, error) {
		b, err := UnmarshalRegAfi270(s)
		if err != nil {
			return "", err
		}
		return MarshalRegAfi270(b, o)
	},
	"In271ConObs": func(s string, o *Options) (string, error) {
		b, err := UnmarshalIn271ConObs(s)
		if err != nil {
			return "", err
		}
		return MarshalIn271ConObs(b, o)
	},
	"ConMed271": func(s string, o *Options) (string, error) {
		b, err := UnmarshalConMed271(s)
		if err != nil {
			return "", err
		}
		return MarshalConMed271(b, o)
	},
	"In271ConDtad": func(s string, o *Options) (string, error) {
		b, err := UnmarshalIn271ConDtad(s)
		if err != nil {
			return "", err
		}
		return MarshalIn271ConDtad(b, o)
	},
	"In278ResCG": func(s string, o *Options) (string, error) {
		b, err := UnmarshalIn278ResCG(s)
		if err != nil {
			return "", err
		}
		return MarshalIn278ResCG(b, o)
	},
}

// TestRoundTrip re-encodes every captured interchange through the Go parser and
// requires the original bytes back.
func TestRoundTrip(t *testing.T) {
	for _, f := range loadFixtures(t) {
		f := f
		t.Run(f.Txn, func(t *testing.T) {
			roundtrip, ok := roundtrips[f.Txn]
			if !ok {
				t.Skip("parser not ported yet")
			}
			control, controlST := beanControls(t, f.Bean)
			got, err := roundtrip(f.EDI, &Options{Control: control, ControlST: controlST, Compat: true})
			if err != nil {
				t.Fatalf("round trip: %v", err)
			}
			if got != f.EDI {
				t.Fatalf("round trip differs\n%s", firstDiff(f.EDI, got))
			}
		})
	}
}

// firstDiff reports the first differing segment with its index and rendered
// values, which is far more useful than a 20 KB diff.
func firstDiff(want, got string) string {
	ws := edi.SplitSegments(want)
	gs := edi.SplitSegments(got)
	for i := 0; i < len(ws) || i < len(gs); i++ {
		var w, g string
		if i < len(ws) {
			w = ws[i]
		}
		if i < len(gs) {
			g = gs[i]
		}
		if w != g {
			return "segment " + itoa(i+1) + "\n want: " + w + "\n  got: " + g
		}
	}
	return "no segment differs (length mismatch)"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// rt wires a transaction's typed bean, its parser and a fresh instance.
type rt struct {
	zero  func() any
	parse func(string) (any, error)
}

var beanRoundTrips = map[string]rt{
	"SolAut271":     {func() any { return new(InSolAut271) }, func(s string) (any, error) { return Unmarshal271(s) }},
	"In997ResAut":   {func() any { return new(In997ResAut) }, func(s string) (any, error) { return Unmarshal997(s) }},
	"ConEntVinc278": {func() any { return new(InConEntVinc278) }, func(s string) (any, error) { return UnmarshalConEntVinc278(s) }},
	"ResEntVinc278": {func() any { return new(InResEntVinc278) }, func(s string) (any, error) { return UnmarshalResEntVinc278(s) }},
	"In278SolCG":    {func() any { return new(In278SolCG) }, func(s string) (any, error) { return UnmarshalIn278SolCG(s) }},
	"RegAfi270":     {func() any { return new(InRegAfi270) }, func(s string) (any, error) { return UnmarshalRegAfi270(s) }},
	"In271ConObs":   {func() any { return new(In271ConObs) }, func(s string) (any, error) { return UnmarshalIn271ConObs(s) }},
	"ConMed271":     {func() any { return new(InConMed271) }, func(s string) (any, error) { return UnmarshalConMed271(s) }},
	"In271ConDtad":  {func() any { return new(In271ConDtad) }, func(s string) (any, error) { return UnmarshalIn271ConDtad(s) }},
	"In278ResCG":    {func() any { return new(In278ResCG) }, func(s string) (any, error) { return UnmarshalIn278ResCG(s) }},
}

// skippedFields are marshalling inputs the reference also sets on the bean
// while encoding, so they never appear in the interchange and cannot be
// recovered from it.
var skippedFields = map[string]bool{
	"NuControl":     true,
	"NuControlST":   true,
	"NoTransaccion": true,
}

// TestBeanRoundTrip requires the parsed bean to match the bean the reference
// implementation started from, field by field. Round-trip byte equality alone
// would not catch a field written to the wrong element if the same mistake
// existed on both sides.
func TestBeanRoundTrip(t *testing.T) {
	for _, f := range loadFixtures(t) {
		f := f
		t.Run(f.Txn, func(t *testing.T) {
			w, ok := beanRoundTrips[f.Txn]
			if !ok {
				t.Skip("parser not ported yet")
			}
			want := w.zero()
			if err := json.Unmarshal(f.Bean, want); err != nil {
				t.Fatalf("load fixture bean: %v", err)
			}
			got, err := w.parse(f.EDI)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			var diffs []string
			compareBeans("", reflect.ValueOf(want).Elem(), reflect.ValueOf(got).Elem(), &diffs)
			if len(diffs) > 0 {
				t.Fatalf("parsed bean differs in %d field(s):\n  %s", len(diffs), strings.Join(diffs, "\n  "))
			}
		})
	}
}

func compareBeans(path string, want, got reflect.Value, diffs *[]string) {
	switch want.Kind() {
	case reflect.String:
		if want.String() != got.String() {
			*diffs = append(*diffs, path+": want "+strconv.Quote(want.String())+", got "+strconv.Quote(got.String()))
		}
	case reflect.Slice:
		if want.Len() != got.Len() {
			*diffs = append(*diffs, path+": want "+strconv.Itoa(want.Len())+" items, got "+strconv.Itoa(got.Len()))
			return
		}
		for i := 0; i < want.Len(); i++ {
			w := want.Index(i)
			g := got.Index(i)
			if w.Kind() == reflect.Ptr {
				w = w.Elem()
				g = g.Elem()
			}
			compareBeans(path+"["+strconv.Itoa(i)+"]", w, g, diffs)
		}
	case reflect.Struct:
		for i := 0; i < want.NumField(); i++ {
			name := want.Type().Field(i).Name
			if skippedFields[name] {
				continue
			}
			compareBeans(path+"."+name, want.Field(i), got.Field(i), diffs)
		}
	}
}
