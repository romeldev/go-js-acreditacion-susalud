package x12n

import (
	"errors"
	"fmt"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// Marshal271 renders an InSolAut271 as an X12N 271 transaction set.
//
// It is the Go equivalent of SolAut271_ToX12N.traducirEstructura271. The
// output carries the fixed-width padding the reference implementation emits,
// so a counterpart built against that library accepts it unchanged.
func Marshal271(b *InSolAut271, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InSolAut271")
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

	counts := fmt.Sprintf("%d#%d#%d#%d",
		len(b.InSolAutProEsp271Detalle),
		len(b.InSolAutTieEsp271Detalle),
		len(b.InSolAutExeCar271Detalle),
		len(b.InSolAut271Detalle))

	e := &encoder{compat: o.Compat}

	// Interchange and transaction set headers.
	if err := e.open(header{
		SenderID:    b.IdRemitente,
		ReceiverID:  b.IdReceptor,
		Date:        b.FeTransaccion,
		Time:        b.HoTransaccion,
		Correlative: b.IdCorrelativo,
		Transaction: b.IdTransaccion,
	}, "HB", ctrl, ctrlST, counts, o.Production); err != nil {
		return "", err
	}

	e.addStd(edi.New(edi.SegBHT).Set(1, "0023").Set(2, b.TiFinalidad))

	// Submitter hierarchy level.
	e.addStd(edi.New(edi.SegHL).Set(1, "1").Set(3, "20").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "PR").Set(2, b.CaRemitente).Set(8, "PI").Set(9, b.NuRucRemitente))
	e.addStd(edi.New(edi.SegREF).Set(1, "OL").Set(2, b.CoAdmisionista))

	// Receiver hierarchy level.
	e.addStd(edi.New(edi.SegHL).Set(1, "2").Set(2, "1").Set(3, "21").Set(4, "1"))
	e.addStd(edi.New(edi.SegNM1).Set(1, "1P").Set(2, b.CaReceptor).Set(8, "FI").Set(9, b.IdReceptor))

	// Subscriber hierarchy level.
	e.addStd(edi.New(edi.SegHL).Set(1, "3").Set(2, "2").Set(3, "22").Set(4, "1"))
	e.addStd(nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "MI", b.CoAfPaciente, b.ApMaternoPaciente))
	e.addStd(edi.New(edi.SegREF).Set(1, "ACC").Set(2, b.CoEsPaciente))
	e.addNested(ref("DD", b.TiDoPaciente), ref4("4A", b.NuDoPaciente))
	e.addStd(edi.New(edi.SegREF).Set(1, "EI").Set(2, b.NuIdenEmpleador))
	e.addNested(ref("CT", b.NuContratoPaciente), ref4("AZ", b.NuPoliza, "ID", b.NuCertificado, "TY", b.CoTiPolizaAfiliacion))
	e.addStd(ref("PRT", b.CoProducto, b.DeProducto))
	e.addNested(ref("18", b.NuPlan), ref4("IMP", b.TiPlanSalud, "ZZ", b.CoMoneda))
	e.addStd(ref("ZZ", b.CoParentesco))
	e.addNested(ref("ZZ", b.SoBeneficio), ref4("3B", b.NuSoBeneficio, "ZZ", b.CoEspecialidad))

	e.addStd(edi.New(edi.SegDMG).
		Set(1, "D8").Set(2, b.FeNacimiento).
		Set(3, b.Genero).Set(4, b.EsMarital))

	e.addStd(edi.New(edi.SegDTP).Set(1, "356").Set(2, "D8").Set(3, b.FeIniVigencia))
	e.addStd(edi.New(edi.SegDTP).Set(1, "357").Set(2, "D8").Set(3, b.FeFinVigencia))

	// Accident coverage.
	e.addNested(ebRest(b.EsCobertura, b.NuDecAccidente), eb13(b.IdInfAccidente))
	e.addStd(ref("3H", b.DeTiAccidente))
	e.addStd(edi.New(edi.SegDTP).Set(1, "356").Set(2, "D8").Set(3, b.FeAfiliacion))
	e.addStd(edi.New(edi.SegDTP).Set(1, "357").Set(2, "D8").Set(3, b.FeOcuAccidente))

	// Pharmacy attention.
	e.addNested(ebRest(b.EsCobertura, b.NuAtencion), eb13(b.IdDerFarmacia))
	e.addStd(ref("PRT", b.TiProducto, b.DeProductoDeFarmacia))
	e.addStd(edi.New(edi.SegDTP).Set(1, "382").Set(2, "D8").Set(3, b.FeAtencion))

	// Coverage.
	e.addNested(ebRest(b.EsCobertura, b.NuCobertura), eb13(b.ObsCobertura))
	e.addStd(edi.New(edi.SegMSG).Set(1, b.MsgObs))
	e.addStd(edi.New(edi.SegMSG).Set(1, b.MsgConEspeciales))

	// Contract holder.
	e.addStd(nm1Person("P5", b.CaContratante, b.NoPaContratante, b.NoContratante, "", "", b.NoMaContratante))
	e.addNested(ref("DD", b.TiDoContratante, ""), ref4(b.IdReContratante, b.CoReContratante))

	// Policy holder.
	e.addStd(nm1Person("C9", b.CaTitular, b.NoPaTitular, b.NoTitular, "MI", b.CoAfTitular, b.NoMaTitular))
	e.addNested(ref("DD", b.TiDoTitular), ref4(b.IdReTitular, b.NuDoTitular))
	e.addStd(edi.New(edi.SegDTP).Set(1, "382").Set(2, "D8").Set(3, b.FeIncTitular))

	e.addNested(edi.New(edi.SegEB).
		Set(1, b.EsCobertura).Set(5, b.NuCobPreExistencia).
		Set(7, b.BeMaxInicial).Set(10, b.CanServicio),
		eb13(b.IdDeProducto))
	e.addNested(ref("D7", b.CoTiCobertura), ref4("ZZ", b.CoSubTiCobertura))
	e.addStd(edi.New(edi.SegMSG).Set(1, b.MsgObsPre))
	e.addStd(edi.New(edi.SegMSG).Set(1, b.MsgConEspecialesPre))

	// Monetary coverage lines.
	e.addStd(edi.New(edi.SegEB).
		Set(1, "C").Set(5, b.CoTiMoneda).Set(7, b.CoPagoFijo).
		Set(9, b.CoCalServicio).Set(10, b.CanCalServicio))
	e.addStd(edi.New(edi.SegEB).Set(1, b.EsCobertura).Set(8, b.CoPagoVariable))
	e.addStd(edi.New(edi.SegEB).Set(1, b.FlagCG).Set(5, b.DeflagCG))

	e.addStd(edi.New(edi.SegDTP).Set(1, "348").Set(2, "D8").Set(3, b.FeFinCarencia))
	e.addStd(edi.New(edi.SegDTP).Set(1, "338").Set(2, "D8").Set(3, b.FeFinEspera))

	// Outpatient procedures. An empty list still emits a placeholder block,
	// which the reference implementation relies on for positional parsing.
	// Only the second EB carries the EB13 composite.
	for _, d := range b.InSolAutProEsp271Detalle {
		e.addStd(edi.New(edi.SegEB).Set(1, "1").Set(9, d.CoInProcedimiento))
		e.addNested(edi.New(edi.SegEB).
			Set(1, "P").
			Set(3, d.CoTiProConAmbulatoria).
			Set(5, d.NuPlanConAmbulatoria).
			Set(7, d.ImDeducible).
			Set(8, d.PoConAmbulatoria).
			Set(9, "5U").
			Set(10, d.FrConAmbulatoria), eb13(d.GeConAmbulatoria))
		e.addStd(edi.New(edi.SegHSD).Set(1, "9S").Set(2, d.CaConAmbulatoria))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.MsgConAmbulatoria))
	}
	if len(b.InSolAutProEsp271Detalle) == 0 {
		e.addStd(edi.New(edi.SegEB).Set(1, "1"))
		e.addNested(edi.New(edi.SegEB).Set(1, "P").Set(9, "5U"), eb13(""))
		e.addStd(edi.New(edi.SegHSD).Set(1, "9S"))
		e.addStd(edi.New(edi.SegMSG))
	}

	// Waiting periods.
	for _, d := range b.InSolAutTieEsp271Detalle {
		e.addNested(edi.New(edi.SegEB).Set(1, "W").Set(3, d.CoTiEspera), eb13(d.IdTiEspera))
		e.addStd(ref("82", d.DeTiEspera))
		e.addStd(edi.New(edi.SegDTP).Set(1, "150").Set(2, "D8").Set(3, d.FeFinVigenciaTiEspera))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.MsgTiEspera))
	}
	if len(b.InSolAutTieEsp271Detalle) == 0 {
		e.addNested(edi.New(edi.SegEB).Set(1, "W"), eb13(""))
		e.addStd(ref("82"))
		e.addStd(edi.New(edi.SegDTP).Set(1, "150").Set(2, "D8"))
		e.addStd(edi.New(edi.SegMSG))
	}

	// Grace periods.
	for _, d := range b.InSolAutExeCar271Detalle {
		e.addNested(edi.New(edi.SegEB).Set(1, "W").Set(3, d.CoExCarencia), eb13(d.IdExCarencia))
		e.addStd(ref("82", d.DeExCarencia))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.MsgExCarencia))
	}
	if len(b.InSolAutExeCar271Detalle) == 0 {
		e.addNested(edi.New(edi.SegEB).Set(1, "W"), eb13(""))
		e.addStd(ref("82"))
		e.addStd(edi.New(edi.SegMSG))
	}

	// The reference implementation repeats the subscriber NM1 before the
	// restriction details.
	e.addStd(nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "MI", b.CoAfPaciente, b.ApMaternoPaciente))

	// Restrictions.
	for _, d := range b.InSolAut271Detalle {
		e.addNested(edi.New(edi.SegEB).Set(1, "W").Set(3, d.CIE10Restricciones), eb13(d.IdRestricciones))
		e.addStd(ref("82", d.ObsRestricciones))
		e.addStd(ref("82", d.DeRestricciones))
		e.addStd(edi.New(edi.SegMSG).Set(1, d.MsgRestricciones))
		e.addStd(edi.New(edi.SegEB).Set(1, b.EsCobertura).Set(8, d.MonTopeRestricciones))
		e.addStd(edi.New(edi.SegDTP).Set(1, "150").Set(2, "D8").Set(3, d.FeFinEsperaRestricciones))
	}
	if len(b.InSolAut271Detalle) == 0 {
		e.addNested(edi.New(edi.SegEB).Set(1, "W"), eb13(""))
		e.addStd(ref("82"))
		e.addStd(ref("82"))
		e.addStd(edi.New(edi.SegMSG))
		e.addStd(edi.New(edi.SegEB))
		e.addStd(edi.New(edi.SegDTP).Set(1, "150").Set(2, "D8"))
	}

	// Registered affiliate hierarchy level.
	e.addStd(edi.New(edi.SegHL).Set(1, "4").Set(2, "3").Set(3, "23").Set(4, "0"))
	e.addStd(nm1Person("IL", b.CaRegafi, b.NoPaRegafi, b.NoRegafi, "MI", b.CoAfRegafi, b.NoMaRegafi))
	e.addNested(ref("DD", b.TiDoRegafi), ref4(b.IdReRegafi, b.NuDoRegafi))
	e.addStd(edi.New(edi.SegDMG).
		Set(1, "D8").Set(2, b.FeNaRegafi).
		Set(3, b.GeRegafi).Set(7, b.CoPaisRegafi))

	if e.err != nil {
		return "", e.err
	}

	// Trailer. SE01 counts every segment emitted so far, matching the
	// reference implementation, which excludes SE itself from the count.
	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// nm1Person builds an NM1 segment in the reference seven-argument shape:
// entity identifier, entity type, surname, given name, ID qualifier, ID code
// and maternal surname.
func nm1Person(id, entityType, surname, given, idQualifier, idCode, maternal string) *edi.Segment {
	return edi.New(edi.SegNM1).
		Set(1, id).
		Set(2, entityType).
		Set(3, surname).
		Set(4, given).
		Set(8, idQualifier).
		Set(9, idCode).
		Set(12, maternal)
}

// ref builds a REF segment; empty trailing arguments leave the element blank.
func ref(code string, values ...string) *edi.Segment {
	s := edi.New(edi.SegREF).Set(1, code)
	for i, v := range values {
		if i+2 > len(s.Values) {
			break
		}
		s.Set(i+2, v)
	}
	return s
}

// ref4 builds the composite REF sub-segment nested inside a parent REF.
func ref4(values ...string) *edi.Segment {
	s := edi.New(edi.SegREF4)
	for i, v := range values {
		if i+1 > len(s.Values) {
			break
		}
		s.Set(i+1, v)
	}
	return s
}

// ebRest builds an EB segment carrying the coverage indicator in EB01 and a
// single value in EB05, the shape used for the accident, pharmacy and
// coverage lines.
func ebRest(coverage, value string) *edi.Segment {
	return edi.New(edi.SegEB).Set(1, coverage).Set(5, value)
}

// eb13 builds the EB13 composite sub-segment (service qualifier "ZZ" plus the
// identifier), which the reference always attaches as a composite.
func eb13(id string) *edi.Segment {
	return edi.New(edi.SegEB13).Set(1, "ZZ").Set(2, id)
}
