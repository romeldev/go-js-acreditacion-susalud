package x12n

import "encoding/json"

// This file registers transactions with the verification harness. Keeping the
// registrations together avoids editing three separate maps for every port.
func init() {
	marshals["ConAse270"] = func(raw json.RawMessage, o *Options) (string, error) {
		var b InConAse270
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalConAse270(&b, o)
	}
	roundtrips["ConAse270"] = func(s string, o *Options) (string, error) {
		b, err := UnmarshalConAse270(s)
		if err != nil {
			return "", err
		}
		return MarshalConAse270(b, o)
	}
	beanRoundTrips["ConAse270"] = rt{
		zero:  func() any { return new(InConAse270) },
		parse: func(s string) (any, error) { return UnmarshalConAse270(s) },
	}
}

func init() {
	marshals["In271ResSctr"] = func(raw json.RawMessage, o *Options) (string, error) {
		var b In271ResSctr
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalIn271ResSctr(&b, o)
	}
	roundtrips["In271ResSctr"] = func(s string, o *Options) (string, error) {
		b, err := UnmarshalIn271ResSctr(s)
		if err != nil {
			return "", err
		}
		return MarshalIn271ResSctr(b, o)
	}
	beanRoundTrips["In271ResSctr"] = rt{
		zero:  func() any { return new(In271ResSctr) },
		parse: func(s string) (any, error) { return UnmarshalIn271ResSctr(s) },
	}
}

func init() {
	marshals["LogAcreInsert271"] = func(raw json.RawMessage, o *Options) (string, error) {
		var b InLogAcreInsert271
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalLogAcreInsert271(&b, o)
	}
	roundtrips["LogAcreInsert271"] = func(s string, o *Options) (string, error) {
		b, err := UnmarshalLogAcreInsert271(s)
		if err != nil {
			return "", err
		}
		return MarshalLogAcreInsert271(b, o)
	}
	beanRoundTrips["LogAcreInsert271"] = rt{
		zero:  func() any { return new(InLogAcreInsert271) },
		parse: func(s string) (any, error) { return UnmarshalLogAcreInsert271(s) },
	}
}

func init() {
	marshals["ConNom271"] = func(raw json.RawMessage, o *Options) (string, error) {
		var b InConNom271
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalConNom271(&b, o)
	}
	roundtrips["ConNom271"] = func(s string, o *Options) (string, error) {
		b, err := UnmarshalConNom271(s)
		if err != nil {
			return "", err
		}
		return MarshalConNom271(b, o)
	}
	beanRoundTrips["ConNom271"] = rt{
		zero:  func() any { return new(InConNom271) },
		parse: func(s string) (any, error) { return UnmarshalConNom271(s) },
	}
}

func init() {
	marshals["ConCod271"] = func(raw json.RawMessage, o *Options) (string, error) {
		var b InConCod271
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalConCod271(&b, o)
	}
	roundtrips["ConCod271"] = func(s string, o *Options) (string, error) {
		b, err := UnmarshalConCod271(s)
		if err != nil {
			return "", err
		}
		return MarshalConCod271(b, o)
	}
	beanRoundTrips["ConCod271"] = rt{
		zero:  func() any { return new(InConCod271) },
		parse: func(s string) (any, error) { return UnmarshalConCod271(s) },
	}
}

func init() {
	marshals["In271ResDeriva"] = func(raw json.RawMessage, o *Options) (string, error) {
		var b In271ResDeriva
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalIn271ResDeriva(&b, o)
	}
	roundtrips["In271ResDeriva"] = func(s string, o *Options) (string, error) {
		b, err := UnmarshalIn271ResDeriva(s)
		if err != nil {
			return "", err
		}
		return MarshalIn271ResDeriva(b, o)
	}
	beanRoundTrips["In271ResDeriva"] = rt{
		zero:  func() any { return new(In271ResDeriva) },
		parse: func(s string) (any, error) { return UnmarshalIn271ResDeriva(s) },
	}
}

func init() {
	marshals["In271ConProc"] = func(raw json.RawMessage, o *Options) (string, error) {
		var b InConProc271
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalIn271ConProc(&b, o)
	}
	roundtrips["In271ConProc"] = func(s string, o *Options) (string, error) {
		b, err := UnmarshalIn271ConProc(s)
		if err != nil {
			return "", err
		}
		return MarshalIn271ConProc(b, o)
	}
	beanRoundTrips["In271ConProc"] = rt{
		zero:  func() any { return new(InConProc271) },
		parse: func(s string) (any, error) { return UnmarshalIn271ConProc(s) },
	}
}

func init() {
	marshals["RegAfi271"] = func(raw json.RawMessage, o *Options) (string, error) {
		var b InRegAfi271
		if err := json.Unmarshal(raw, &b); err != nil {
			return "", err
		}
		return MarshalRegAfi271(&b, o)
	}
	roundtrips["RegAfi271"] = func(s string, o *Options) (string, error) {
		b, err := UnmarshalRegAfi271(s)
		if err != nil {
			return "", err
		}
		return MarshalRegAfi271(b, o)
	}
	beanRoundTrips["RegAfi271"] = rt{
		zero:  func() any { return new(InRegAfi271) },
		parse: func(s string) (any, error) { return UnmarshalRegAfi271(s) },
	}
}
