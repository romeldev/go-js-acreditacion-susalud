package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalIn271ConDtad renders an In271ConDtad: a 271 carrying the subscriber's
// address and contact details.
func MarshalIn271ConDtad(b *In271ConDtad, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil In271ConDtad")
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

	// N4 carries the values in N405/N406.
	e.addStd(edi.New(edi.SegN3).Set(1, b.DeDirPaciente1).Set(2, b.DeDirPaciente2))
	e.addStd(edi.New(edi.SegN4).Set(5, "W").Set(6, b.CoUbigeoPaciente))

	e.addStd(edi.New(edi.SegPER).Set(1, "IC").Set(2, b.NoContacto))
	// PER carries the contact role in PER01, then EM/type/value/TE/number.
	e.addStd(edi.New(edi.SegPER).
		Set(1, "IP").Set(3, "EM").Set(4, b.EmContacto).
		Set(5, "TE").Set(6, b.NuTeContacto))

	e.addStd(nm1Person("CA", b.TiCaCalificador, b.ApPaNoEmCalificador, b.NoEmCalificador, "", "", b.ApMaCalificador))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalIn271ConDtad parses a 271 with subscriber address and contacts.
func UnmarshalIn271ConDtad(s string) (*In271ConDtad, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &In271ConDtad{}
	level, nm1, pers := "", 0, 0
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
			if level == "1" {
				b.CaRemitente = at(el, 2)
				continue
			}
			if level == "2" {
				b.CaReceptor, b.NuRucReceptor = at(el, 2), at(el, 9)
				continue
			}
			nm1++
			if nm1 == 1 {
				b.CaPaciente = at(el, 2)
				b.ApPaternoPaciente = at(el, 3)
				b.NoPaciente = at(el, 4)
				b.CoAfPaciente = at(el, 9)
				b.ApMaternoPaciente = at(el, 12)
			} else {
				b.TiCaCalificador = at(el, 2)
				b.ApPaNoEmCalificador = at(el, 3)
				b.NoEmCalificador = at(el, 4)
				b.ApMaCalificador = at(el, 12)
			}
		case "N3":
			b.DeDirPaciente1, b.DeDirPaciente2 = at(el, 1), at(el, 2)
		case "N4":
			b.CoUbigeoPaciente = at(el, 6)
		case "PER":
			pers++
			if pers == 1 {
				b.NoContacto = at(el, 2)
			} else {
				b.EmContacto = at(el, 4)
				b.NuTeContacto = at(el, 6)
			}
		}
	}
	return b, nil
}
