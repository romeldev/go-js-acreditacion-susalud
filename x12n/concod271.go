package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalConCod271 renders an InConCod271 as a 271 coding enquiry reply.
func MarshalConCod271(b *InConCod271, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InConCod271")
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
	e.addNested(ref("OL", b.UserRemitente), ref4("Y8", b.PassRemitente))
	e.addStd(edi.New(edi.SegDTP).Set(1, "163").Set(2, "D8").Set(3, b.FeUpFoto))

	e.addStd(edi.New(edi.SegHL).Set(1, "2").Set(2, "1").Set(3, "21").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "1P").Set(2, b.CaReceptor).Set(8, "FI").Set(9, b.NuRucReceptor))

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "0"))

	// The subscriber identifier is "3" here, not "IL".
	e.addStd(nm1Person("3", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "MI", b.CoAfPaciente, b.ApMaternoPaciente))
	e.addStd(ref("ACC", b.CoEsPaciente))
	e.addNested(ref("DD", b.TiDoPaciente), ref4("4A", b.NuDoPaciente))
	e.addStd(ref("EI", b.NuIdenPaciente))
	e.addNested(ref("CT", b.NuContratoPaciente),
		ref4("AZ", b.NuPoliza, "ID", b.NuCertificado, "TY", b.CoTiPoliza))
	e.addStd(ref("PRT", b.CoProducto, b.DeProducto))
	e.addNested(ref("18", b.NuPlan), ref4("IMP", b.TiPlanSalud, "ZZ", b.CoMoneda))
	e.addStd(ref("ZZ", b.CoParentesco))
	e.addNested(ref("ZZ", b.SoBeneficio), ref4("3B", b.NuSoBeneficio))

	e.addStd(edi.New(edi.SegDMG).Set(1, "D8").Set(2, b.FeNacimiento).Set(3, b.Genero).Set(4, b.EsMarital))
	e.addStd(edi.New(edi.SegDTP).Set(1, "356").Set(2, "D8").Set(3, b.FeIniVigencia))
	e.addStd(edi.New(edi.SegDTP).Set(1, "357").Set(2, "D8").Set(3, b.FeFinVigencia))

	e.addStd(nm1Person("P5", b.TiCaContratante, b.NoPaContratante, b.NoContratante, "", "", b.NoMaContratante))
	e.addNested(ref("DD", b.TiDoContratante), ref4("XX5", b.CoReContratante))

	e.addStd(nm1Person("C9", b.CaTitular, b.NoPaTitular, b.NoTitular, "MI", b.CoAfTitular, b.NoMaTitular))
	e.addNested(ref("DD", b.TiDoTitular), ref4("4A", b.NuDoTitular))
	e.addStd(edi.New(edi.SegDTP).Set(1, "382").Set(2, "D8").Set(3, b.FeInsTitular))

	emit := func(d *InConCod271Detalle) {
		e.addNested(edi.New(edi.SegEB).
			Set(1, d.InfBeneficio).Set(5, d.NuCobertura).Set(7, d.BeMaxInicial).
			Set(8, d.MoCobertura).Set(9, d.CoInRestriccion).Set(10, d.CanServicio),
			eb13(d.IdProducto))
		e.addNested(ref("D7", d.CoTiCobertura), ref4("ZZ", d.CoSubTiCobertura))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.MsgObs))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.MsgConEspeciales))
		e.addStd(edi.New(edi.SegEB).
			Set(1, "C").Set(5, d.CoTiMoneda).Set(7, d.CoPagoFijo).
			Set(9, d.CoCalServicio).Set(10, d.CanCalServicio))
		e.addStd(edi.New(edi.SegEB).Set(1, "1").Set(8, d.CoPagoVariable))
		e.addStd(edi.New(edi.SegEB).Set(1, d.FlagCaGarantia).Set(5, d.DeflagCaGarantia))
		e.addStd(edi.New(edi.SegDTP).Set(1, "327").Set(2, "D8").Set(3, d.FeFinCarencia))
		e.addStd(edi.New(edi.SegDTP).Set(1, "338").Set(2, "D8").Set(3, d.FeFinEspera))
	}

	for _, d := range b.InConCod271Detalles {
		emit(d)
	}
	if len(b.InConCod271Detalles) == 0 {
		emit(&InConCod271Detalle{})
	}

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalConCod271 parses a 271 coding enquiry reply.
func UnmarshalConCod271(s string) (*InConCod271, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &InConCod271{}
	level, nm1, refs, msgs, eb := "", 0, 0, 0, 0

	last := func() *InConCod271Detalle {
		if len(b.InConCod271Detalles) == 0 {
			b.InConCod271Detalles = append(b.InConCod271Detalles, &InConCod271Detalle{})
		}
		return b.InConCod271Detalles[len(b.InConCod271Detalles)-1]
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
			switch nm1 {
			case 1:
				b.CaPaciente = at(el, 2)
				b.ApPaternoPaciente = at(el, 3)
				b.NoPaciente = at(el, 4)
				b.CoAfPaciente = at(el, 9)
				b.ApMaternoPaciente = at(el, 12)
			case 2:
				b.TiCaContratante = at(el, 2)
				b.NoPaContratante = at(el, 3)
				b.NoContratante = at(el, 4)
				b.NoMaContratante = at(el, 12)
			case 3:
				b.CaTitular = at(el, 2)
				b.NoPaTitular = at(el, 3)
				b.NoTitular = at(el, 4)
				b.CoAfTitular = at(el, 9)
				b.NoMaTitular = at(el, 12)
			}
		case "REF":
			refs++
			c := composite(el, 4)
			if refs > 11 {
				d := last()
				d.CoTiCobertura = at(el, 2)
				d.CoSubTiCobertura = part(c, 1)
				continue
			}
			switch refs {
			case 1:
				b.UserRemitente = at(el, 2)
				b.PassRemitente = part(c, 1)
			case 2:
				b.CoEsPaciente = at(el, 2)
			case 3:
				b.TiDoPaciente = at(el, 2)
				b.NuDoPaciente = part(c, 1)
			case 4:
				b.NuIdenPaciente = at(el, 2)
			case 5:
				b.NuContratoPaciente = at(el, 2)
				b.NuPoliza = part(c, 1)
				b.NuCertificado = part(c, 3)
				b.CoTiPoliza = part(c, 5)
			case 6:
				b.CoProducto = at(el, 2)
				b.DeProducto = at(el, 3)
			case 7:
				b.NuPlan = at(el, 2)
				b.TiPlanSalud = part(c, 1)
				b.CoMoneda = part(c, 3)
			case 8:
				b.CoParentesco = at(el, 2)
			case 9:
				b.SoBeneficio = at(el, 2)
				b.NuSoBeneficio = part(c, 1)
			case 10:
				b.TiDoContratante = at(el, 2)
				b.IdReContratante = part(c, 0)
				b.CoReContratante = part(c, 1)
			case 11:
				b.TiDoTitular = at(el, 2)
				b.NuDoTitular = part(c, 1)
			}
		case "DMG":
			b.FeNacimiento, b.Genero, b.EsMarital = at(el, 2), at(el, 3), at(el, 4)
		case "EB":
			if level != "3" {
				continue
			}
			eb++
			switch eb {
			case 1:
				b.InConCod271Detalles = append(b.InConCod271Detalles, &InConCod271Detalle{
					InfBeneficio:    at(el, 1),
					NuCobertura:     at(el, 5),
					BeMaxInicial:    at(el, 7),
					MoCobertura:     at(el, 8),
					CoInRestriccion: at(el, 9),
					CanServicio:     at(el, 10),
					IdProducto:      part(composite(el, 13), 1),
				})
			case 2:
				d := last()
				d.CoTiMoneda = at(el, 5)
				d.CoPagoFijo = at(el, 7)
				d.CoCalServicio = at(el, 9)
				d.CanCalServicio = at(el, 10)
			case 3:
				last().CoPagoVariable = at(el, 8)
			case 4:
				d := last()
				d.FlagCaGarantia = at(el, 1)
				d.DeflagCaGarantia = at(el, 5)
				eb = 0
			}
		case "MSG":
			msgs++
			d := last()
			if msgs%2 == 1 {
				d.MsgObs = at(el, 1)
			} else {
				d.MsgConEspeciales = at(el, 1)
			}
		case "DTP":
			switch at(el, 1) {
			case "163":
				b.FeUpFoto = at(el, 3)
			case "356":
				b.FeIniVigencia = at(el, 3)
			case "357":
				b.FeFinVigencia = at(el, 3)
			case "382":
				b.FeInsTitular = at(el, 3)
			case "327":
				last().FeFinCarencia = at(el, 3)
			case "338":
				last().FeFinEspera = at(el, 3)
			}
		}
	}
	b.InConCod271Detalles = prune(b.InConCod271Detalles, func(d *InConCod271Detalle) bool { return d.isZero() })
	return b, nil
}
