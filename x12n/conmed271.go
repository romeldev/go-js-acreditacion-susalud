package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalConMed271 renders an InConMed271 as a 271 medical-consultation reply.
//
// The reference emits the subscriber NM1 once per detail block and repeats the
// placeholder block when the detail list is empty, so the output is reproduced
// the same way.
func MarshalConMed271(b *InConMed271, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InConMed271")
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

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "0"))

	subscriber := func() *edi.Segment {
		return nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "MI", b.CoPaciente, b.ApMaternoPaciente)
	}

	for _, d := range b.InConMed271Detalles {
		e.addStd(subscriber())
		e.addNested(edi.New(edi.SegEB).Set(1, "W").Set(3, d.CoSeCIE10), eb13(d.CoRestriccion))
		e.addStd(ref("82", "", d.DePreexistencia))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.MsgObsPr).Set(2, "LC").Set(3, d.IdFuenteRE))
		e.addStd(edi.New(edi.SegEB).Set(1, d.EsCobertura).Set(8, d.MoDiagnostico))
		e.addStd(edi.New(edi.SegDTP).Set(1, "150").Set(2, "D8").Set(3, d.TiEspera))
	}
	if len(b.InConMed271Detalles) == 0 {
		e.addStd(subscriber())
		e.addNested(edi.New(edi.SegEB).Set(1, "W"), eb13(""))
		e.addStd(ref("82", "", ""))
		e.addStd(edi.New(edi.SegMSG))
		e.addStd(edi.New(edi.SegEB))
		e.addStd(edi.New(edi.SegDTP).Set(1, "150").Set(2, "D8"))
	}

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalConMed271 parses a 271 medical-consultation reply.
//
// The first EB of each block opens a detail and the second closes it, which is
// the alternating counter the reference uses.
func UnmarshalConMed271(s string) (*InConMed271, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &InConMed271{}
	level := ""
	eb := 1
	last := func() *InConMed271Detalle {
		if len(b.InConMed271Detalles) == 0 {
			b.InConMed271Detalles = append(b.InConMed271Detalles, &InConMed271Detalle{})
		}
		return b.InConMed271Detalles[len(b.InConMed271Detalles)-1]
	}

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
				b.CoPaciente = at(el, 9)
				b.ApMaternoPaciente = at(el, 12)
			}
		case "EB":
			if level != "3" {
				continue
			}
			if eb == 1 {
				d := &InConMed271Detalle{
					CoSeCIE10:     at(el, 3),
					CoRestriccion: part(composite(el, 13), 1),
				}
				b.InConMed271Detalles = append(b.InConMed271Detalles, d)
				eb = 2
			} else {
				d := last()
				d.EsCobertura = at(el, 1)
				d.MoDiagnostico = at(el, 8)
				eb = 1
			}
		case "REF":
			last().DePreexistencia = at(el, 3)
		case "MSG":
			d := last()
			d.MsgObsPr = at(el, 1)
			d.IdFuenteRE = at(el, 3)
		case "DTP":
			last().TiEspera = at(el, 3)
		}
	}
	b.InConMed271Detalles = prune(b.InConMed271Detalles, func(d *InConMed271Detalle) bool { return d.isZero() })
	return b, nil
}
