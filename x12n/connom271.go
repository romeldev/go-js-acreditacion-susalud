package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalConNom271 renders an InConNom271 as a 271 name enquiry reply.
func MarshalConNom271(b *InConNom271, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InConNom271")
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
	e.addStd(edi.New(edi.SegNM1).Set(1, "1P").Set(2, b.CaReceptor).Set(8, "FI").Set(9, b.NuRucReceptor))

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "0"))

	emit := func(d *InConNom271Detalle) {
		e.addStd(nm1Person("IL", d.CaPaciente, d.ApPaternoPaciente, d.NoPaciente, "MI", d.CoAfPaciente, d.ApMaternoPaciente))
		e.addStd(ref("ACC", d.CoEsPaciente))
		e.addNested(ref("DD", d.TiDoPaciente), ref4("4A", d.NuDoPaciente))
		e.addStd(ref("CT", d.NuContratoPaciente))
		e.addNested(ref("PRT", d.CoProducto, d.CoDescripcion), ref4("ZZ", d.NuSCTR))
		e.addStd(ref("ZZ", d.CoParentesco))
		e.addStd(ref("18", d.NuPlan))
		e.addStd(edi.New(edi.SegDMG).Set(1, "D8").Set(2, d.FeNacimiento).Set(3, d.Genero).Set(4, d.EsMarital))
		e.addStd(nm1Person("P5", d.TiCaContratante, d.NoPaContratante, d.NoContratante, "", "", d.NoMaContratante))
		e.addNested(ref("DD", d.TiDoContratante), ref4("XX5", d.CoReContratante))
	}

	for _, d := range b.InConNom271Detalles {
		emit(d)
	}
	if len(b.InConNom271Detalles) == 0 {
		emit(&InConNom271Detalle{})
	}

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalConNom271 parses a 271 name enquiry reply.
func UnmarshalConNom271(s string) (*InConNom271, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &InConNom271{}
	level, nm1, refs := "", 0, 0
	last := func() *InConNom271Detalle {
		if len(b.InConNom271Detalles) == 0 {
			b.InConNom271Detalles = append(b.InConNom271Detalles, &InConNom271Detalle{})
		}
		return b.InConNom271Detalles[len(b.InConNom271Detalles)-1]
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
					b.InConNom271Detalles = append(b.InConNom271Detalles, &InConNom271Detalle{
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
		case "REF":
			d := last()
			refs++
			c := composite(el, 4)
			switch refs {
			case 1:
				d.CoEsPaciente = at(el, 2)
			case 2:
				d.TiDoPaciente = at(el, 2)
				d.NuDoPaciente = part(c, 1)
			case 3:
				d.NuContratoPaciente = at(el, 2)
			case 4:
				d.CoProducto = at(el, 2)
				d.CoDescripcion = at(el, 3)
				d.NuSCTR = part(c, 1)
			case 5:
				d.CoParentesco = at(el, 2)
			case 6:
				d.NuPlan = at(el, 2)
			case 7:
				d.TiDoContratante = at(el, 2)
				d.IdReContratante = part(c, 0)
				d.CoReContratante = part(c, 1)
			}
		case "DMG":
			d := last()
			d.FeNacimiento = at(el, 2)
			d.Genero = at(el, 3)
			d.EsMarital = at(el, 4)
		}
	}
	b.InConNom271Detalles = prune(b.InConNom271Detalles, func(d *InConNom271Detalle) bool { return d.isZero() })
	return b, nil
}
