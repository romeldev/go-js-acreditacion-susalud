package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalLogAcreInsert271 renders an InLogAcreInsert271: the accreditation
// insertion record built on a 271 shape.
func MarshalLogAcreInsert271(b *InLogAcreInsert271, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InLogAcreInsert271")
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
	e.addStd(edi.New(edi.SegNM1).Set(1, "PR").Set(2, b.CaRemitente).Set(8, "PI").Set(9, b.NuRucRemitente))

	e.addStd(edi.New(edi.SegHL).Set(1, "2").Set(2, "1").Set(3, "21").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "1P").Set(2, b.CaReceptor).Set(8, "FI").Set(9, b.IdReceptor))

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "0"))
	e.addStd(nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "MI", b.CoAfPaciente, b.ApMaternoPaciente))

	e.addStd(ref("ACC", b.CoEsPaciente))
	e.addNested(ref("DD", b.TiDoPaciente), ref4("4A", b.NuDoPaciente))
	e.addNested(ref("CT", b.NuContratoPaciente), ref4("TY", b.CoTiPolizaAfiliacion))
	e.addStd(ref("PRT", b.CoProducto))
	e.addStd(ref("18", b.NuPlan))
	e.addStd(ref("ZZ", b.CoParentesco))

	e.addStd(edi.New(edi.SegDMG).Set(1, "D8").Set(2, b.FeNacimiento).Set(3, b.Genero))
	e.addStd(edi.New(edi.SegDTP).Set(1, "356").Set(2, "D8").Set(3, b.FeIniVigencia))

	// EB01 "1", EB02 the coverage number, EB05 its description.
	e.addStd(edi.New(edi.SegEB).Set(1, "1").Set(2, b.NuCobertura).Set(5, b.DeCobertura))

	// The contract holder NM1 uses the four-argument shape, so the entity type
	// lands in NM102 and the qualifier elements stay blank.
	e.addStd(edi.New(edi.SegNM1).Set(1, "P5").Set(2, b.CaContratante).Set(8, "").Set(9, ""))
	e.addNested(ref("DD", b.TiDoContratante), ref4(b.IdReContratante, b.RucContratante))

	e.addStd(edi.New(edi.SegNM1).Set(1, "C9").Set(2, "1").Set(8, "MI").Set(9, b.CoAfiliadoTitular))
	e.addStd(nm1Person("IL", b.CaResponsableAut, b.NoPaResponsableAut, b.NoResponsableAut, "", "", b.NoMaResponsableAut))
	e.addNested(ref("DD", b.TiDoResponsableAut), ref4("4A", b.NuDoResponsableAut))

	e.addStd(ref("ZZ", b.NuAutorizacion))
	e.addStd(edi.New(edi.SegDMG).Set(1, "DT").Set(2, b.FeHoAutorizacion))

	e.addStd(edi.New(edi.SegEB).Set(1, "1").Set(7, b.BeMaxInicial))
	e.addStd(edi.New(edi.SegEB).Set(1, "C").Set(7, b.CoPagoFijo))
	e.addStd(edi.New(edi.SegEB).Set(1, "1").Set(8, b.CoPagoVariable))
	e.addStd(edi.New(edi.SegEB).Set(1, b.FlagCartaGarantia).Set(5, b.DeFlagCartaGarantia))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalLogAcreInsert271 parses an accreditation insertion record.
func UnmarshalLogAcreInsert271(s string) (*InLogAcreInsert271, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &InLogAcreInsert271{}
	level, nm1, refs, dmgs, ebs := "", 0, 0, 0, 0

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
				b.CaRemitente, b.NuRucRemitente = at(el, 2), at(el, 9)
				continue
			}
			if level == "2" {
				b.CaReceptor = at(el, 2)
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
				b.CaContratante = at(el, 2)
			case 3:
				b.CoAfiliadoTitular = at(el, 9)
			case 4:
				b.CaResponsableAut = at(el, 2)
				b.NoPaResponsableAut = at(el, 3)
				b.NoResponsableAut = at(el, 4)
				b.NoMaResponsableAut = at(el, 12)
			}
		case "REF":
			refs++
			c := composite(el, 4)
			switch refs {
			case 1:
				b.CoEsPaciente = at(el, 2)
			case 2:
				b.TiDoPaciente = at(el, 2)
				b.NuDoPaciente = part(c, 1)
			case 3:
				b.NuContratoPaciente = at(el, 2)
				b.CoTiPolizaAfiliacion = part(c, 1)
			case 4:
				b.CoProducto = at(el, 2)
			case 5:
				b.NuPlan = at(el, 2)
			case 6:
				b.CoParentesco = at(el, 2)
			case 7:
				b.TiDoContratante = at(el, 2)
				b.IdReContratante = part(c, 0)
				b.RucContratante = part(c, 1)
			case 8:
				b.TiDoResponsableAut = at(el, 2)
				b.NuDoResponsableAut = part(c, 1)
			case 9:
				b.NuAutorizacion = at(el, 2)
			}
		case "DMG":
			dmgs++
			if dmgs == 1 {
				b.FeNacimiento = at(el, 2)
				b.Genero = at(el, 3)
			} else {
				b.FeHoAutorizacion = at(el, 2)
			}
		case "DTP":
			b.FeIniVigencia = at(el, 3)
		case "EB":
			ebs++
			switch ebs {
			case 1:
				b.NuCobertura = at(el, 2)
				b.DeCobertura = at(el, 5)
			case 2:
				b.BeMaxInicial = at(el, 7)
			case 3:
				b.CoPagoFijo = at(el, 7)
			case 4:
				b.CoPagoVariable = at(el, 8)
			case 5:
				b.FlagCartaGarantia = at(el, 1)
				b.DeFlagCartaGarantia = at(el, 5)
			}
		}
	}
	return b, nil
}
