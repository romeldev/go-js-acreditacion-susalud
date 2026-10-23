package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// ref4third builds a REF4 whose third value lands in element 4, which is what
// the reference's three-argument overload does. The element in between stays
// blank, so a reader must look at component index 3 for the value.
func ref4third(a, b, c string) *edi.Segment {
	return edi.New(edi.SegREF4).Set(1, a).Set(2, b).Set(4, c)
}

// MarshalConAse270 renders an InConAse270 as a 270 coverage enquiry.
func MarshalConAse270(b *InConAse270, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InConAse270")
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
	e.addStd(edi.New(edi.SegPRV).Set(1, "OR").Set(2, "").Set(3, b.TxRequest))

	e.addStd(edi.New(edi.SegHL).Set(1, "2").Set(2, "1").Set(3, "21").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "1P").Set(2, b.CaRemitente).Set(8, "FI").Set(9, b.IdReceptor))

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "0"))
	e.addStd(nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "", b.CoAfPaciente, b.ApMaternoPaciente))

	// The first two references are emitted standalone, without a composite.
	e.addStd(ref("DD", b.TiDocumento))
	e.addStd(ref("4A", b.NuDocumento))
	e.addNested(ref("PRT", b.CoProducto, b.DeProducto), ref4("ZZ", b.CoInProducto))
	e.addNested(ref("D7", b.NuCobertura, b.DeCobertura), ref4third(b.CaServicio, b.CoCalservicio, b.BeMaxInicial))
	e.addNested(ref("D7", b.CoTiCobertura), ref4("ZZ", b.CoSuTiCobertura))
	e.addStd(ref("8X", b.CoAplicativoTx))
	e.addStd(ref("S2", b.CoEspecialidad))
	e.addStd(ref("ZZ", b.CoParentesco))
	e.addNested(ref("18", b.NuPlan), ref4("ZZ", b.NuAutOrigen))
	e.addStd(ref("PRT", b.TiAccidente))
	e.addStd(edi.New(edi.SegDTP).Set(1, "447").Set(2, "D8").Set(3, b.FeAccidente))

	e.addStd(nm1Person("P5", b.TiCaContratante, b.NoPaContratante, b.NoContratante, "", "", b.NoMaContratante))
	e.addNested(ref("DD", b.TiDoContratante), ref4(b.IdReContratante, b.CoReContratante))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalConAse270 parses a 270 coverage enquiry.
func UnmarshalConAse270(s string) (*InConAse270, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &InConAse270{}
	level, nm1, refs := "", 0, 0
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
		case "PRV":
			b.TxRequest = at(el, 3)
		case "NM1":
			switch level {
			case "1":
				b.CaRemitente, b.NuRucRemitente = at(el, 2), at(el, 9)
			case "2":
				b.CaReceptor = at(el, 2)
			default:
				nm1++
				if nm1 == 1 {
					b.CaPaciente = at(el, 2)
					b.ApPaternoPaciente = at(el, 3)
					b.NoPaciente = at(el, 4)
					b.CoAfPaciente = at(el, 9)
					b.ApMaternoPaciente = at(el, 12)
				} else {
					b.TiCaContratante = at(el, 2)
					b.NoPaContratante = at(el, 3)
					b.NoContratante = at(el, 4)
					b.NoMaContratante = at(el, 12)
				}
			}
		case "REF":
			refs++
			c := composite(el, 4)
			switch refs {
			case 1:
				b.TiDocumento = at(el, 2)
			case 2:
				b.NuDocumento = at(el, 2)
			case 3:
				b.CoProducto = at(el, 2)
				b.DeProducto = at(el, 3)
				b.CoInProducto = part(c, 1)
			case 4:
				b.NuCobertura = at(el, 2)
				b.DeCobertura = at(el, 3)
				b.CaServicio = part(c, 0)
				b.CoCalservicio = part(c, 1)
				b.BeMaxInicial = part(c, 3)
			case 5:
				b.CoTiCobertura = at(el, 2)
				b.CoSuTiCobertura = part(c, 1)
			case 6:
				b.CoAplicativoTx = at(el, 2)
			case 7:
				b.CoEspecialidad = at(el, 2)
			case 8:
				b.CoParentesco = at(el, 2)
			case 9:
				b.NuPlan = at(el, 2)
				b.NuAutOrigen = part(c, 1)
			case 10:
				b.TiAccidente = at(el, 2)
			case 11:
				b.TiDoContratante = at(el, 2)
				b.IdReContratante = part(c, 0)
				b.CoReContratante = part(c, 1)
			}
		case "DTP":
			b.FeAccidente = at(el, 3)
		}
	}
	return b, nil
}
