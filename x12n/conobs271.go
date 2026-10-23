package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalIn271ConObs renders an In271ConObs as a 271 with free observations.
func MarshalIn271ConObs(b *In271ConObs, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil In271ConObs")
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
	}, "HB", ctrl, ctrlST, "", o.Production); err != nil {
		return "", err
	}

	e.addStd(edi.New(edi.SegBHT).Set(1, "0022").Set(2, b.TiFinalidad))

	e.addStd(edi.New(edi.SegHL).Set(1, "1").Set(3, "20").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "PR").Set(2, b.CaRemitente).Set(8, "PI").Set(9, b.IdRemitente))

	e.addStd(edi.New(edi.SegHL).Set(1, "2").Set(2, "1").Set(3, "21").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "1P").Set(2, b.CaReceptor).Set(8, "SV").Set(9, b.NuRucReceptor))

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "1"))
	e.addStd(nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "MI", b.CoAfPaciente, b.ApMaternoPaciente))

	e.addStd(edi.New(edi.SegEB).Set(1, "1").Set(5, b.CoSubCobertura))
	e.addStd(edi.New(edi.SegMSG).Set(1, b.TeMsgLibre1))
	e.addStd(edi.New(edi.SegMSG).Set(1, b.TeMsgLibre2))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalIn271ConObs parses a 271 with free observations.
func UnmarshalIn271ConObs(s string) (*In271ConObs, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &In271ConObs{}
	level := ""
	msgs := 0
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
				b.CaRemitente = at(el, 2)
			case "2":
				b.CaReceptor, b.NuRucReceptor = at(el, 2), at(el, 9)
			case "3":
				b.CaPaciente = at(el, 2)
				b.ApPaternoPaciente = at(el, 3)
				b.NoPaciente = at(el, 4)
				b.CoAfPaciente = at(el, 9)
				b.ApMaternoPaciente = at(el, 12)
			}
		case "EB":
			b.CoSubCobertura = at(el, 5)
		case "MSG":
			msgs++
			if msgs == 1 {
				b.TeMsgLibre1 = at(el, 1)
			} else if msgs == 2 {
				b.TeMsgLibre2 = at(el, 1)
			}
		}
	}
	return b, nil
}
