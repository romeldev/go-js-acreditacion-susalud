package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalRegAfi270 renders an InRegAfi270 as a 270 affiliate registration.
func MarshalRegAfi270(b *InRegAfi270, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InRegAfi270")
	}
	if o == nil {
		o = &Options{}
	}
	ctrl, err := o.control()
	if err != nil {
		return "", err
	}
	ctrlST, err := o.controlST()
	if err != nil {
		return "", err
	}
	b.NuControl, b.NuControlST = ctrl, ctrlST

	e := &encoder{compat: o.Compat}
	if err := e.open(header{
		SenderID: b.IdRemitente, ReceiverID: b.IdReceptor,
		Date: b.FeTransaccion, Time: b.HoTransaccion,
		Correlative: b.IdCorrelativo, Transaction: b.IdTransaccion,
	}, "HS", ctrl, ctrlST, "", o.Production); err != nil {
		return "", err
	}

	e.addStd(edi.New(edi.SegBHT).Set(1, "0022").Set(2, b.TiFinalidad))

	e.addStd(edi.New(edi.SegHL).Set(1, "1").Set(3, "20").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "PR").Set(2, b.CaRemitente).Set(8, "PI").Set(9, b.NuRucRemitente))

	e.addStd(edi.New(edi.SegHL).Set(1, "2").Set(2, "1").Set(3, "21").Set(4, "1"))
	e.addStd(nm1Person("RGA", b.CaReceptor, b.IdReceptor, "", "", "", ""))

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "0"))
	e.addStd(nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "", "", b.ApMaternoPaciente))

	e.addNested(ref("DD", b.TiDocumento), ref4("4A", b.NuDocumento))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalRegAfi270 parses a 270 affiliate registration.
func UnmarshalRegAfi270(s string) (*InRegAfi270, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &InRegAfi270{}
	level := ""
	for _, raw := range segs {
		id, el := edi.ParseSegment(raw)
		switch id {
		case "ISA":
			b.IdRemitente, b.IdReceptor, b.IdCorrelativo = at(el, 6), at(el, 8), at(el, 13)
		case "GS":
			b.FeTransaccion, b.HoTransaccion = at(el, 4), at(el, 5)
		case "ST":
			b.IdTransaccion = at(el, 1)
		case "BHT":
			b.TiFinalidad = at(el, 2)
		case "HL":
			level = at(el, 1)
		case "NM1":
			switch level {
			case "1":
				b.CaRemitente, b.NuRucRemitente = at(el, 2), at(el, 9)
			case "2":
				b.CaReceptor = at(el, 2)
			case "3":
				b.CaPaciente = at(el, 2)
				b.ApPaternoPaciente = at(el, 3)
				b.NoPaciente = at(el, 4)
				b.ApMaternoPaciente = at(el, 12)
			}
		case "REF":
			b.TiDocumento = at(el, 2)
			b.NuDocumento = part(composite(el, 4), 1)
		}
	}
	return b, nil
}
