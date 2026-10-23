package x12n

// isZero reports whether the detail carries no data. The parser uses it to
// drop the placeholder blocks the encoder emits for empty lists, so a
// marshal/unmarshal cycle reproduces the original bytes.

func (d *InSolAutProEsp271Detalle) isZero() bool {
	return d.CoInProcedimiento == "" &&
		d.CoTiProConAmbulatoria == "" &&
		d.NuPlanConAmbulatoria == "" &&
		d.ImDeducible == "" &&
		d.PoConAmbulatoria == "" &&
		d.FrConAmbulatoria == "" &&
		d.GeConAmbulatoria == "" &&
		d.CaConAmbulatoria == "" &&
		d.MsgConAmbulatoria == ""
}

func (d *InSolAutTieEsp271Detalle) isZero() bool {
	return d.CoTiEspera == "" &&
		d.IdTiEspera == "" &&
		d.DeTiEspera == "" &&
		d.FeFinVigenciaTiEspera == "" &&
		d.MsgTiEspera == ""
}

func (d *InSolAutExeCar271Detalle) isZero() bool {
	return d.CoExCarencia == "" &&
		d.IdExCarencia == "" &&
		d.DeExCarencia == "" &&
		d.MsgExCarencia == ""
}

func (d *InSolAutRes271Detalle) isZero() bool {
	return d.CIE10Restricciones == "" &&
		d.IdRestricciones == "" &&
		d.ObsRestricciones == "" &&
		d.DeRestricciones == "" &&
		d.MsgRestricciones == "" &&
		d.MonTopeRestricciones == "" &&
		d.FeFinEsperaRestricciones == ""
}

func (d *InConMed271Detalle) isZero() bool {
	return d.CoSeCIE10 == "" &&
		d.CoRestriccion == "" &&
		d.DePreexistencia == "" &&
		d.MsgObsPr == "" &&
		d.IdFuenteRE == "" &&
		d.EsCobertura == "" &&
		d.MoDiagnostico == "" &&
		d.TiEspera == ""
}

func (d *In278ResCGDetalle) isZero() bool {
	return d.CaPaciente == "" && d.ApPaternoPaciente == "" && d.NoPaciente == "" &&
		d.CoAfPaciente == "" && d.ApMaternoPaciente == "" && d.CoTiDoPaciente == "" &&
		d.NuDoPaciente == "" && d.MonPago == "" && d.TiCaContratante == "" &&
		d.NoPaContratante == "" && d.NoContratante == "" && d.NoMaContratante == "" &&
		d.TiDoContratante == "" && d.IdCaReContratante == "" && d.NuCaReContratante == "" &&
		d.DeCarGarantia == "" && d.NuSoCarGarantia == "" && d.NuCarGarantia == "" &&
		d.VeCarGarantia == "" && d.EsCarGarantia == "" && d.CoProducto == "" &&
		d.CoProcedimiento == "" && d.DeProcedimiento == "" && d.NuPlan == "" &&
		d.TiPlanSalud == "" && d.CoMoneda == "" && d.FeCarGarantia == ""
}

func (d *In271ResSctrDetalle) isZero() bool {
	return d.TiCaContratante == "" && d.NoEmApPaContratante == "" && d.CoEmContratante == "" &&
		d.IdCaReContratante == "" && d.ReIdContratante == "" && d.TiCaLuAtencion == "" &&
		d.NoEmLuAtencion == "" && d.CoEmReLuAtencion == "" && d.IdCaReLuAtencion == "" &&
		d.ReIdLuAtencion == "" && d.CoLuAtencion == "" && d.DeTiAccidente == "" &&
		d.FeAfiliacion == "" && d.FeOcAccidente == ""
}

func (d *InConNom271Detalle) isZero() bool {
	return d.CaPaciente == "" && d.ApPaternoPaciente == "" && d.NoPaciente == "" &&
		d.CoAfPaciente == "" && d.ApMaternoPaciente == "" && d.CoEsPaciente == "" &&
		d.TiDoPaciente == "" && d.NuDoPaciente == "" && d.NuContratoPaciente == "" &&
		d.CoProducto == "" && d.CoDescripcion == "" && d.NuSCTR == "" &&
		d.CoParentesco == "" && d.NuPlan == "" && d.FeNacimiento == "" &&
		d.Genero == "" && d.EsMarital == "" && d.TiCaContratante == "" &&
		d.NoPaContratante == "" && d.NoContratante == "" && d.NoMaContratante == "" &&
		d.TiDoContratante == "" && d.IdReContratante == "" && d.CoReContratante == ""
}

func (d *InConCod271Detalle) isZero() bool {
	return d.InfBeneficio == "" && d.NuCobertura == "" && d.BeMaxInicial == "" &&
		d.MoCobertura == "" && d.CoInRestriccion == "" && d.CanServicio == "" &&
		d.IdProducto == "" && d.CoTiCobertura == "" && d.CoSubTiCobertura == "" &&
		d.MsgObs == "" && d.MsgConEspeciales == "" && d.CoTiMoneda == "" &&
		d.CoPagoFijo == "" && d.CoCalServicio == "" && d.CanCalServicio == "" &&
		d.CoPagoVariable == "" && d.FlagCaGarantia == "" && d.DeflagCaGarantia == "" &&
		d.FeFinCarencia == "" && d.FeFinEspera == ""
}

func (d *In271ResDerivaDetalle) isZero() bool {
	return d.CaPaciente == "" && d.ApPaternoPaciente == "" && d.NoPaciente == "" &&
		d.CoAfPaciente == "" && d.ApMaternoPaciente == "" && d.CoTiDoPaciente == "" &&
		d.NuDoPaciente == "" && d.CoParentesco == "" && d.TiDoTrabajoAfiliado == "" &&
		d.NuDoTrabajoAfiliado == "" && d.NuAtencion == "" && d.TeMsgLibre1 == "" &&
		d.CoTiProducto == "" && d.DeProducto == "" && d.CoTiCobertura == "" &&
		d.CoSubTiCobertura == "" && d.FeAtSalud == "" && d.NoLuAtencion == "" &&
		d.CoLuAtencion == "" && d.TiDoContratante == "" && d.IdCaReferencia == "" &&
		d.ReIdContratante == ""
}

func (d *InConProc271Detalle) isZero() bool {
	return d.CoInProcedimiento == "" && d.CoInRestriccion == "" && d.CoProcedimiento == "" &&
		d.ImDeducible == "" && d.PoCuExDecimal == "" && d.NuFrecuencia == "" &&
		d.CoSexo == "" && d.TiNuDias == "" && d.TeMsgObservacion == "" &&
		d.IdFuentePE == "" && d.CoTiEspera == "" && d.DeTiEspera == "" &&
		d.FeFinVigencia == "" && d.TeMsgTiEspera == "" && d.IdFuenteTE == "" &&
		d.CoExCarencia == "" && d.DeExCarencia == "" && d.TeMsgExCarencia == "" &&
		d.IdFuenteEC == ""
}

func (d *InRegAfi271Detalle) isZero() bool {
	return d.CaPaciente == "" && d.ApPaternoPaciente == "" && d.NoPaciente == "" &&
		d.CoPaciente == "" && d.ApMaternoPaciente == "" && d.TiDocumentoPaciente == "" &&
		d.NuDocumentoPaciente == "" && d.CoContratoPaciente == "" && d.EsPaciente == "" &&
		d.TiRegimenPaciente == "" && d.TiPlanPaciente == "" && d.CoProductoPaciente == "" &&
		d.CoPlanPaciente == "" && d.NuCarnetPaciente == "" && d.CoVinculoFamPaciente == "" &&
		d.ApCasadaPaciente == "" && d.DeUbigeoPaciente == "" && d.FeNacePaciente == "" &&
		d.GePaciente == "" && d.CoPaisPaciente == "" && d.FeFallecidoPaciente == "" &&
		d.FeIniAfiliaPaciente == "" && d.FeFinAfiliaPaciente == "" && d.FeFinAtencionPaciente == "" &&
		d.CaAseguradora == "" && d.CoAseguradora == "" && d.CaTitular == "" &&
		d.ApPaternoTitular == "" && d.NoTitular == "" && d.CoTitular == "" &&
		d.ApMaternoTitular == "" && d.TiDocumentoTitular == "" && d.NuDocumentoTitular == "" &&
		d.CoEstablecimientoTitular == "" && d.FeFallecidoTitular == "" && d.CoPaisTitular == "" &&
		d.CaContratante == "" && d.ApPaternoContratante == "" && d.NoContratante == "" &&
		d.ApMaternoContratante == "" && d.TiDocumentoContratante == "" && d.IdReContratante == "" &&
		d.NuDocumentoContratante == "" && d.CoPaisContratante == ""
}
