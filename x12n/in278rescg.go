package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalIn278ResCG renders an In278ResCG: a 278 certification determination.
//
// Each detail block repeats the subscriber NM1 followed by the contract
// references, and an empty list still emits one placeholder block.
func MarshalIn278ResCG(b *In278ResCG, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil In278ResCG")
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

	emit := func(d *In278ResCGDetalle) {
		e.addStd(nm1Person("IL", d.CaPaciente, d.ApPaternoPaciente, d.NoPaciente, "MI", d.CoAfPaciente, d.ApMaternoPaciente))

		e.addNested(ref("DD", d.CoTiDoPaciente), ref4("4A", d.NuDoPaciente))

		// INS01 "Y" (subscriber), INS02 "F7" (relationship), INS17 the amount.
		e.addStd(edi.New(edi.SegINS).Set(1, "Y").Set(2, "F7").Set(17, d.MonPago))

		e.addStd(nm1Person("P5", d.TiCaContratante, d.NoPaContratante, d.NoContratante, "", "", d.NoMaContratante))
		e.addNested(ref("DD", d.TiDoContratante), ref4(d.IdCaReContratante, d.NuCaReContratante))

		e.addNested(ref("5F", "", d.DeCarGarantia), ref4("ZZ", d.NuSoCarGarantia))
		e.addNested(ref("BB", d.NuCarGarantia), ref4("ZZ", d.VeCarGarantia))
		e.addStd(ref("ACC", d.EsCarGarantia))
		e.addStd(ref("OZ", d.CoProducto))
		e.addStd(ref("ZZ", d.CoProcedimiento, d.DeProcedimiento))
		e.addNested(ref("18", d.NuPlan), ref4("IMP", d.TiPlanSalud, "ZZ", d.CoMoneda))
		e.addStd(edi.New(edi.SegDTP).Set(1, "447").Set(2, "DT").Set(3, d.FeCarGarantia))
	}

	for _, d := range b.In278ResCGDetalles {
		emit(d)
	}
	if len(b.In278ResCGDetalles) == 0 {
		emit(&In278ResCGDetalle{})
	}

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalIn278ResCG parses a 278 certification determination.
func UnmarshalIn278ResCG(s string) (*In278ResCG, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &In278ResCG{}
	level, nm1, refs := "", 0, 0
	last := func() *In278ResCGDetalle {
		if len(b.In278ResCGDetalles) == 0 {
			b.In278ResCGDetalles = append(b.In278ResCGDetalles, &In278ResCGDetalle{})
		}
		return b.In278ResCGDetalles[len(b.In278ResCGDetalles)-1]
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
			switch {
			case level == "1":
				b.CaRemitente = at(el, 2)
			case level == "2":
				b.CaReceptor, b.NuRucReceptor = at(el, 2), at(el, 9)
			default:
				nm1++
				if nm1%2 == 1 {
					// Each block opens with the subscriber and then the
					// contract holder, so odd occurrences start a new detail.
					b.In278ResCGDetalles = append(b.In278ResCGDetalles, &In278ResCGDetalle{
						CaPaciente:        at(el, 2),
						ApPaternoPaciente: at(el, 3),
						NoPaciente:        at(el, 4),
						CoAfPaciente:      at(el, 9),
						ApMaternoPaciente: at(el, 12),
					})
					refs = 0
				} else {
					d := last()
					d.TiCaContratante = at(el, 2)
					d.NoPaContratante = at(el, 3)
					d.NoContratante = at(el, 4)
					d.NoMaContratante = at(el, 12)
				}
			}
		case "INS":
			last().MonPago = at(el, 17)
		case "REF":
			d := last()
			refs++
			c := composite(el, 4)
			switch refs {
			case 1:
				d.CoTiDoPaciente = at(el, 2)
				d.NuDoPaciente = part(c, 1)
			case 2:
				d.TiDoContratante = at(el, 2)
				d.IdCaReContratante = part(c, 0)
				d.NuCaReContratante = part(c, 1)
			case 3:
				d.DeCarGarantia = at(el, 3)
				d.NuSoCarGarantia = part(c, 1)
			case 4:
				d.NuCarGarantia = at(el, 2)
				d.VeCarGarantia = part(c, 1)
			case 5:
				d.EsCarGarantia = at(el, 2)
			case 6:
				d.CoProducto = at(el, 2)
			case 7:
				d.CoProcedimiento = at(el, 2)
				d.DeProcedimiento = at(el, 3)
			case 8:
				d.NuPlan = at(el, 2)
				d.TiPlanSalud = part(c, 1)
				d.CoMoneda = part(c, 3)
			}
		case "DTP":
			last().FeCarGarantia = at(el, 3)
		}
	}
	b.In278ResCGDetalles = prune(b.In278ResCGDetalles, func(d *In278ResCGDetalle) bool { return d.isZero() })
	return b, nil
}
