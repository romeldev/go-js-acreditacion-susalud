package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalIn271ResSctr renders an In271ResSctr: a 271 SCTR determination.
func MarshalIn271ResSctr(b *In271ResSctr, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil In271ResSctr")
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

	// The patient REF carries the document value in REF02 and repeats it in
	// the composite.
	e.addNested(ref("DD", b.CoTiDoPaciente, b.NuDocPaciente), ref4("4A", b.NuDocPaciente))

	emit := func(d *In271ResSctrDetalle) {
		e.addStd(nm1Person("C9", d.TiCaContratante, d.NoEmApPaContratante, "", "MI", "", ""))
		e.addStd(ref("EI", d.CoEmContratante, ""))
		e.addNested(ref("DD", "6", ""), ref4(d.IdCaReContratante, d.ReIdContratante))

		e.addStd(nm1Person("C9", d.TiCaLuAtencion, d.NoEmLuAtencion, "", "", "", ""))
		e.addStd(ref("EI", d.CoEmReLuAtencion, ""))
		e.addNested(ref("DD", "6", ""), ref4(d.IdCaReLuAtencion, d.ReIdLuAtencion))

		e.addStd(ref("BB", d.CoLuAtencion, ""))
		e.addStd(ref("3H", d.DeTiAccidente))
		e.addStd(edi.New(edi.SegDTP).Set(1, "382").Set(2, "D8").Set(3, d.FeAfiliacion))
		e.addStd(edi.New(edi.SegDTP).Set(1, "382").Set(2, "D8").Set(3, d.FeOcAccidente))
	}

	for _, d := range b.In271ResSctrDetalles {
		emit(d)
	}
	if len(b.In271ResSctrDetalles) == 0 {
		emit(&In271ResSctrDetalle{})
	}

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalIn271ResSctr parses a 271 SCTR determination.
func UnmarshalIn271ResSctr(s string) (*In271ResSctr, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &In271ResSctr{}
	level, nm1, refs, dtps := "", 0, 0, 0

	last := func() *In271ResSctrDetalle {
		if len(b.In271ResSctrDetalles) == 0 {
			b.In271ResSctrDetalles = append(b.In271ResSctrDetalles, &In271ResSctrDetalle{})
		}
		return b.In271ResSctrDetalles[len(b.In271ResSctrDetalles)-1]
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
			if level == "1" {
				b.CaRemitente = at(el, 2)
				continue
			}
			if level == "2" {
				b.CaReceptor, b.NuRucReceptor = at(el, 2), at(el, 9)
				continue
			}
			nm1++
			switch {
			case nm1 == 1:
				b.CaPaciente = at(el, 2)
				b.ApPaternoPaciente = at(el, 3)
				b.NoPaciente = at(el, 4)
				b.CoAfPaciente = at(el, 9)
				b.ApMaternoPaciente = at(el, 12)
			case (nm1-2)%2 == 0:
				// Each block opens with the contract holder, then the
				// attending physician.
				b.In271ResSctrDetalles = append(b.In271ResSctrDetalles, &In271ResSctrDetalle{
					TiCaContratante:     at(el, 2),
					NoEmApPaContratante: at(el, 3),
				})
				refs, dtps = 0, 0
			default:
				d := last()
				d.TiCaLuAtencion = at(el, 2)
				d.NoEmLuAtencion = at(el, 3)
			}
		case "REF":
			// The patient REF precedes every detail block.
			if len(b.In271ResSctrDetalles) == 0 {
				b.CoTiDoPaciente = at(el, 2)
				b.NuDocPaciente = at(el, 3)
				continue
			}
			d := last()
			refs++
			c := composite(el, 4)
			switch refs {
			case 1:
				d.CoEmContratante = at(el, 2)
			case 2:
				d.IdCaReContratante = part(c, 0)
				d.ReIdContratante = part(c, 1)
			case 3:
				d.CoEmReLuAtencion = at(el, 2)
			case 4:
				d.IdCaReLuAtencion = part(c, 0)
				d.ReIdLuAtencion = part(c, 1)
			case 5:
				d.CoLuAtencion = at(el, 2)
			case 6:
				d.DeTiAccidente = at(el, 2)
			}
		case "DTP":
			if level != "3" {
				continue
			}
			d := last()
			dtps++
			if dtps == 1 {
				d.FeAfiliacion = at(el, 3)
			} else {
				d.FeOcAccidente = at(el, 3)
			}
		}
	}
	b.In271ResSctrDetalles = prune(b.In271ResSctrDetalles, func(d *In271ResSctrDetalle) bool { return d.isZero() })
	return b, nil
}
