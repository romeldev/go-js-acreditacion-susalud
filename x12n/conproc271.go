package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalIn271ConProc renders an InConProc271: a 271 procedure-consultation
// reply.
func MarshalIn271ConProc(b *InConProc271, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InConProc271")
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
	}, "HI", ctrl, ctrlST, "", o.Production); err != nil {
		return "", err
	}

	e.addStd(edi.New(edi.SegBHT).Set(1, "0022").Set(2, b.TiFinalidad))

	e.addStd(edi.New(edi.SegHL).Set(1, "1").Set(3, "20").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "PR").Set(2, b.CaRemitente).Set(8, "PI").Set(9, b.IdRemitente))

	e.addStd(edi.New(edi.SegHL).Set(1, "2").Set(2, "1").Set(3, "21").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "1P").Set(2, b.CaReceptor).Set(8, "SV").Set(9, b.NuRucReceptor))

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "0"))
	e.addStd(nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "MI", b.CoAfPaciente, b.ApMaternoPaciente))

	emit := func(d *InConProc271Detalle) {
		e.addStd(edi.New(edi.SegEB).Set(1, "1").Set(9, d.CoInProcedimiento))
		e.addStd(edi.New(edi.SegEB).Set(1, "1").Set(9, d.CoInRestriccion))
		e.addNested(edi.New(edi.SegEB).
			Set(1, "PE").Set(3, d.CoProcedimiento).
			Set(7, d.ImDeducible).Set(8, d.PoCuExDecimal).
			Set(9, "5U").Set(10, d.NuFrecuencia), eb13(d.CoSexo))
		e.addStd(edi.New(edi.SegHSD).Set(1, "9S").Set(2, d.TiNuDias))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.TeMsgObservacion).Set(2, "LC").Set(3, d.IdFuentePE))

		e.addStd(edi.New(edi.SegEB).Set(1, "TE").Set(3, d.CoTiEspera))
		e.addStd(ref("82", "", d.DeTiEspera))
		e.addStd(edi.New(edi.SegDTP).Set(1, "327").Set(2, "D8").Set(3, d.FeFinVigencia))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.TeMsgTiEspera).Set(2, "LC").Set(3, d.IdFuenteTE))

		e.addStd(edi.New(edi.SegEB).Set(1, "EC").Set(3, d.CoExCarencia))
		e.addStd(ref("82", "", d.DeExCarencia))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.TeMsgExCarencia).Set(2, "LC").Set(3, d.IdFuenteEC))
	}

	for _, d := range b.InConProc271Detalles {
		emit(d)
	}
	if len(b.InConProc271Detalles) == 0 {
		emit(&InConProc271Detalle{})
	}

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalIn271ConProc parses a 271 procedure-consultation reply.
//
// The EB segments are dispatched by their EB01 qualifier: the two "1" rows open
// a detail block, then "PE", "TE" and "EC" populate its three sections.
func UnmarshalIn271ConProc(s string) (*InConProc271, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &InConProc271{}
	level, ones, refs, msgs := "", 0, 0, 0

	last := func() *InConProc271Detalle {
		if len(b.InConProc271Detalles) == 0 {
			b.InConProc271Detalles = append(b.InConProc271Detalles, &InConProc271Detalle{})
		}
		return b.InConProc271Detalles[len(b.InConProc271Detalles)-1]
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
				b.CoAfPaciente = at(el, 9)
				b.ApMaternoPaciente = at(el, 12)
			}
		case "EB":
			if level != "3" {
				continue
			}
			switch at(el, 1) {
			case "PE":
				d := last()
				d.CoProcedimiento = at(el, 3)
				d.ImDeducible = at(el, 7)
				d.PoCuExDecimal = at(el, 8)
				d.NuFrecuencia = at(el, 10)
				d.CoSexo = part(composite(el, 13), 1)
			case "TE":
				last().CoTiEspera = at(el, 3)
			case "EC":
				last().CoExCarencia = at(el, 3)
			default:
				// EB01 "1": the odd occurrence opens the block with the
				// procedure code, the even one carries the restriction.
				ones++
				if ones%2 == 1 {
					b.InConProc271Detalles = append(b.InConProc271Detalles, &InConProc271Detalle{
						CoInProcedimiento: at(el, 9),
					})
					refs, msgs = 0, 0
				} else {
					last().CoInRestriccion = at(el, 9)
				}
			}
		case "HSD":
			last().TiNuDias = at(el, 2)
		case "REF":
			d := last()
			refs++
			if refs == 1 {
				d.DeTiEspera = at(el, 3)
			} else {
				d.DeExCarencia = at(el, 3)
			}
		case "DTP":
			last().FeFinVigencia = at(el, 3)
		case "MSG":
			d := last()
			msgs++
			switch msgs {
			case 1:
				d.TeMsgObservacion = at(el, 1)
				d.IdFuentePE = at(el, 3)
			case 2:
				d.TeMsgTiEspera = at(el, 1)
				d.IdFuenteTE = at(el, 3)
			case 3:
				d.TeMsgExCarencia = at(el, 1)
				d.IdFuenteEC = at(el, 3)
			}
		}
	}
	b.InConProc271Detalles = prune(b.InConProc271Detalles, func(d *InConProc271Detalle) bool { return d.isZero() })
	return b, nil
}
