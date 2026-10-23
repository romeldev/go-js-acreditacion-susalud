package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalRegAfi271 renders an InRegAfi271 as a 271 affiliate registration
// response.
func MarshalRegAfi271(b *InRegAfi271, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InRegAfi271")
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

	e.addStd(edi.New(edi.SegBHT).Set(1, "0024").Set(2, b.TiFinalidad))

	e.addStd(edi.New(edi.SegHL).Set(1, "1").Set(3, "20").Set(4, "1"))
	e.addStd(nm1Person("RGA", b.CaRemitente, b.IdRemitente, "", "", "", ""))

	e.addStd(edi.New(edi.SegHL).Set(1, "2").Set(2, "1").Set(3, "21").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "PR").Set(2, b.CaReceptor).Set(8, "PI").Set(9, b.NuRucReceptor))

	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "0"))

	emit := func(d *InRegAfi271Detalle) {
		e.addStd(nm1Person("IL", d.CaPaciente, d.ApPaternoPaciente, d.NoPaciente, "MI", d.CoPaciente, d.ApMaternoPaciente))
		e.addNested(ref("DD", d.TiDocumentoPaciente), ref4("4A", d.NuDocumentoPaciente))
		e.addStd(ref("CT", d.CoContratoPaciente))
		e.addStd(ref("ACC", d.EsPaciente))
		e.addStd(ref("BLT", d.TiRegimenPaciente))
		e.addStd(ref("6P", d.TiPlanPaciente))
		e.addStd(ref("OZ", d.CoProductoPaciente))
		e.addStd(ref("18", d.CoPlanPaciente))
		e.addStd(ref("C6", d.NuCarnetPaciente))
		e.addStd(ref("ACD", d.CoVinculoFamPaciente))

		e.addStd(edi.New(edi.SegN2).Set(1, d.ApCasadaPaciente))
		e.addStd(edi.New(edi.SegN4).Set(5, "W").Set(6, d.DeUbigeoPaciente))
		e.addStd(edi.New(edi.SegDMG).
			Set(1, "D8").Set(2, d.FeNacePaciente).
			Set(3, d.GePaciente).Set(7, d.CoPaisPaciente))

		e.addStd(edi.New(edi.SegDTP).Set(1, "442").Set(2, "D8").Set(3, d.FeFallecidoPaciente))
		e.addStd(edi.New(edi.SegDTP).Set(1, "276").Set(2, "D8").Set(3, d.FeIniAfiliaPaciente))
		e.addStd(edi.New(edi.SegDTP).Set(1, "093").Set(2, "D8").Set(3, d.FeFinAfiliaPaciente))
		e.addStd(edi.New(edi.SegDTP).Set(1, "349").Set(2, "D8").Set(3, d.FeFinAtencionPaciente))

		e.addStd(edi.New(edi.SegNM1).Set(1, "PR").Set(2, d.CaAseguradora).Set(8, "PI").Set(9, d.CoAseguradora))
		e.addStd(nm1Person("C9", d.CaTitular, d.ApPaternoTitular, d.NoTitular, "MI", d.CoTitular, d.ApMaternoTitular))
		e.addNested(ref("DD", d.TiDocumentoTitular), ref4("A4", d.NuDocumentoTitular))
		e.addStd(ref("EI", d.CoEstablecimientoTitular))
		e.addStd(edi.New(edi.SegDMG).Set(1, "D8").Set(2, d.FeFallecidoTitular).Set(7, d.CoPaisTitular))

		e.addStd(nm1Person("C9", d.CaContratante, d.ApPaternoContratante, d.NoContratante, "", "", d.ApMaternoContratante))
		e.addNested(ref("DD", d.TiDocumentoContratante), ref4(d.IdReContratante, d.NuDocumentoContratante))
		e.addStd(edi.New(edi.SegDMG).Set(7, d.CoPaisContratante))
	}

	for _, d := range b.Regafi271Detalles {
		emit(d)
	}
	if len(b.Regafi271Detalles) == 0 {
		emit(&InRegAfi271Detalle{})
	}

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// UnmarshalRegAfi271 parses a 271 affiliate registration response.
//
// Each block carries four NM1 entities in a fixed order: the affiliate, the
// insurer, the policy holder and the contract holder.
func UnmarshalRegAfi271(s string) (*InRegAfi271, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}
	b := &InRegAfi271{}
	level, nm1, refs, dmgs := "", 0, 0, 0

	last := func() *InRegAfi271Detalle {
		if len(b.Regafi271Detalles) == 0 {
			b.Regafi271Detalles = append(b.Regafi271Detalles, &InRegAfi271Detalle{})
		}
		return b.Regafi271Detalles[len(b.Regafi271Detalles)-1]
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
				switch (nm1 - 1) % 4 {
				case 0:
					b.Regafi271Detalles = append(b.Regafi271Detalles, &InRegAfi271Detalle{
						CaPaciente:        at(el, 2),
						ApPaternoPaciente: at(el, 3),
						NoPaciente:        at(el, 4),
						CoPaciente:        at(el, 9),
						ApMaternoPaciente: at(el, 12),
					})
					refs, dmgs = 0, 0
				case 1:
					d := last()
					d.CaAseguradora = at(el, 2)
					d.CoAseguradora = at(el, 9)
				case 2:
					d := last()
					d.CaTitular = at(el, 2)
					d.ApPaternoTitular = at(el, 3)
					d.NoTitular = at(el, 4)
					d.CoTitular = at(el, 9)
					d.ApMaternoTitular = at(el, 12)
				case 3:
					d := last()
					d.CaContratante = at(el, 2)
					d.ApPaternoContratante = at(el, 3)
					d.NoContratante = at(el, 4)
					d.ApMaternoContratante = at(el, 12)
				}
			}
		case "REF":
			d := last()
			refs++
			c := composite(el, 4)
			switch refs {
			case 1:
				d.TiDocumentoPaciente = at(el, 2)
				d.NuDocumentoPaciente = part(c, 1)
			case 2:
				d.CoContratoPaciente = at(el, 2)
			case 3:
				d.EsPaciente = at(el, 2)
			case 4:
				d.TiRegimenPaciente = at(el, 2)
			case 5:
				d.TiPlanPaciente = at(el, 2)
			case 6:
				d.CoProductoPaciente = at(el, 2)
			case 7:
				d.CoPlanPaciente = at(el, 2)
			case 8:
				d.NuCarnetPaciente = at(el, 2)
			case 9:
				d.CoVinculoFamPaciente = at(el, 2)
			case 10:
				d.TiDocumentoTitular = at(el, 2)
				d.NuDocumentoTitular = part(c, 1)
			case 11:
				d.CoEstablecimientoTitular = at(el, 2)
			case 12:
				d.TiDocumentoContratante = at(el, 2)
				d.IdReContratante = part(c, 0)
				d.NuDocumentoContratante = part(c, 1)
			}
		case "N2":
			last().ApCasadaPaciente = at(el, 1)
		case "N4":
			last().DeUbigeoPaciente = at(el, 6)
		case "DMG":
			d := last()
			dmgs++
			switch dmgs {
			case 1:
				d.FeNacePaciente = at(el, 2)
				d.GePaciente = at(el, 3)
				d.CoPaisPaciente = at(el, 7)
			case 2:
				d.FeFallecidoTitular = at(el, 2)
				d.CoPaisTitular = at(el, 7)
			case 3:
				d.CoPaisContratante = at(el, 7)
			}
		case "DTP":
			d := last()
			switch at(el, 1) {
			case "442":
				d.FeFallecidoPaciente = at(el, 3)
			case "276":
				d.FeIniAfiliaPaciente = at(el, 3)
			case "093":
				d.FeFinAfiliaPaciente = at(el, 3)
			case "349":
				d.FeFinAtencionPaciente = at(el, 3)
			}
		}
	}
	b.Regafi271Detalles = prune(b.Regafi271Detalles, func(d *InRegAfi271Detalle) bool { return d.isZero() })
	return b, nil
}
