package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalIn271ResDeriva renders an In271ResDeriva: a 271 claim referral reply.
func MarshalIn271ResDeriva(b *In271ResDeriva, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil In271ResDeriva")
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

	emit := func(d *In271ResDerivaDetalle) {
		e.addStd(nm1Person("IL", d.CaPaciente, d.ApPaternoPaciente, d.NoPaciente, "MI", d.CoAfPaciente, d.ApMaternoPaciente))
		e.addNested(ref("DD", d.CoTiDoPaciente), ref4("4A", d.NuDoPaciente))
		e.addStd(ref("ZZ", d.CoParentesco))
		e.addNested(ref("DD", d.TiDoTrabajoAfiliado), ref4("XX5", d.NuDoTrabajoAfiliado))
		// This REF carries free text in REF03 and no composite.
		e.addStd(ref("BB", d.NuAtencion, d.TeMsgLibre1))
		e.addStd(ref("PRT", d.CoTiProducto, d.DeProducto))
		e.addNested(ref("D7", d.CoTiCobertura), ref4("ZZ", d.CoSubTiCobertura))
		e.addStd(edi.New(edi.SegDTP).Set(1, "382").Set(2, "D8").Set(3, d.FeAtSalud))

		e.addStd(nm1Person("C9", "2", d.NoLuAtencion, "", "MI", "", ""))
		e.addStd(ref("EI", d.CoLuAtencion, ""))
		e.addNested(ref("DD", d.TiDoContratante), ref4(d.IdCaReferencia, d.ReIdContratante))
	}

	for _, d := range b.In271ResDerivaDetalles {
		emit(d)
	}
	if len(b.In271ResDerivaDetalles) == 0 {
		emit(&In271ResDerivaDetalle{})
	}

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalIn271ResDeriva parses a 271 claim referral reply.
func UnmarshalIn271ResDeriva(s string) (*In271ResDeriva, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &In271ResDeriva{}
	level, nm1, refs := "", 0, 0

	last := func() *In271ResDerivaDetalle {
		if len(b.In271ResDerivaDetalles) == 0 {
			b.In271ResDerivaDetalles = append(b.In271ResDerivaDetalles, &In271ResDerivaDetalle{})
		}
		return b.In271ResDerivaDetalles[len(b.In271ResDerivaDetalles)-1]
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
			default:
				nm1++
				if nm1%2 == 1 {
					// Each block opens with the subscriber then the attending
					// location.
					b.In271ResDerivaDetalles = append(b.In271ResDerivaDetalles, &In271ResDerivaDetalle{
						CaPaciente:        at(el, 2),
						ApPaternoPaciente: at(el, 3),
						NoPaciente:        at(el, 4),
						CoAfPaciente:      at(el, 9),
						ApMaternoPaciente: at(el, 12),
					})
					refs = 0
				} else {
					last().NoLuAtencion = at(el, 3)
				}
			}
		case "REF":
			d := last()
			refs++
			c := composite(el, 4)
			switch refs {
			case 1:
				d.CoTiDoPaciente = at(el, 2)
				d.NuDoPaciente = part(c, 1)
			case 2:
				d.CoParentesco = at(el, 2)
			case 3:
				d.TiDoTrabajoAfiliado = at(el, 2)
				d.NuDoTrabajoAfiliado = part(c, 1)
			case 4:
				d.NuAtencion = at(el, 2)
				d.TeMsgLibre1 = at(el, 3)
			case 5:
				d.CoTiProducto = at(el, 2)
				d.DeProducto = at(el, 3)
			case 6:
				d.CoTiCobertura = at(el, 2)
				d.CoSubTiCobertura = part(c, 1)
			case 7:
				d.CoLuAtencion = at(el, 2)
			case 8:
				d.TiDoContratante = at(el, 2)
				d.IdCaReferencia = part(c, 0)
				d.ReIdContratante = part(c, 1)
			}
		case "DTP":
			last().FeAtSalud = at(el, 3)
		}
	}
	b.In271ResDerivaDetalles = prune(b.In271ResDerivaDetalles, func(d *In271ResDerivaDetalle) bool { return d.isZero() })
	return b, nil
}
