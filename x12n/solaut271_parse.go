package x12n

import (
	"strconv"
	"strings"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// at returns the 1-based element of a parsed segment, or "" when the sender
// omitted it. The reference parser indexes raw arrays and throws on short
// segments; tolerating absent trailing elements is what makes the Go parser
// safe on real traffic.
func at(el []string, n int) string {
	if n < 1 || n > len(el) {
		return ""
	}
	return el[n-1]
}

// composite splits a nested composite element such as "AZ:3134:ID:1:TY:1"
// into its trimmed parts.
func composite(el []string, n int) []string {
	raw := at(el, n)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, edi.ComponentSep)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// part returns the i-th part (0-based) of a composite, or "".
func part(parts []string, i int) string {
	if i < 0 || i >= len(parts) {
		return ""
	}
	return parts[i]
}

// counts271 parses the ST03 detail counter "pe#te#ec#re".
//
// The reference replaces a zero count with 1 before reading it back, which
// keeps its loops from running on an empty block. The same rule is applied
// here so the parser follows the reference state machine.
func counts271(segs []string) (pe, te, ec, re int) {
	for _, raw := range segs {
		id, el := edi.ParseSegment(raw)
		if id != "ST" {
			continue
		}
		fields := strings.Split(at(el, 3), "#")
		if len(fields) < 4 {
			return 1, 1, 1, 1
		}
		pe, te, ec, re = num(fields[0]), num(fields[1]), num(fields[2]), num(fields[3])
		if pe == 0 {
			pe = 1
		}
		if te == 0 {
			te = 1
		}
		if ec == 0 {
			ec = 1
		}
		if re == 0 {
			re = 1
		}
		return
	}
	return 1, 1, 1, 1
}

func num(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

// Unmarshal271 parses an X12N 271 transaction set into an InSolAut271.
//
// The reference parser keeps its progress in package-level static fields, so
// it cannot be used concurrently and leaks state between calls. This parser
// carries the same counters in a per-call struct.
//
// Detail items that carry no data are dropped: the encoder emits placeholder
// blocks for empty lists, and dropping the empties makes a marshal/unmarshal
// cycle reproduce the original bytes.
func Unmarshal271(s string) (*InSolAut271, error) {
	segs := edi.SplitSegments(s)
	if len(segs) == 0 {
		return nil, &edi.ParseError{Segment: "interchange", Reason: "empty input"}
	}

	b := &InSolAut271{}
	// The reference parser starts every occurrence counter at 1.
	st := &state271{tagNM1: 1, tagREF: 1, tagDMG: 1, tagDTP: 1, tagEB: 1, tagMSG: 1}
	st.canPE, st.canTE, st.canEC, st.canRE = counts271(segs)

	for _, raw := range segs {
		id, el := edi.ParseSegment(raw)
		switch id {
		case "ISA":
			b.NoTransaccion = "271_SOL_AUT"
			b.IdRemitente = at(el, 6)
			b.IdReceptor = at(el, 8)
			b.IdCorrelativo = at(el, 13)
		case "GS":
			b.FeTransaccion = at(el, 4)
			b.HoTransaccion = at(el, 5)
		case "ST":
			b.IdTransaccion = at(el, 1)
		case "BHT":
			b.TiFinalidad = at(el, 2)
		case "HL":
			st.tagHL = at(el, 1)
		case "NM1":
			st.nm1(b, el)
		case "REF":
			st.ref(b, el)
		case "DMG":
			st.dmg(b, el)
		case "DTP":
			st.dtp(b, el)
		case "EB":
			st.eb(b, el)
		case "MSG":
			st.msg(b, el)
		case "HSD":
			if d := st.lastPE(b, false); d != nil {
				d.CaConAmbulatoria = at(el, 2)
			}
		}
	}

	st.pruneEmpty(b)
	return b, nil
}

type state271 struct {
	tagHL  string
	tagNM1 int
	tagREF int
	tagDMG int
	tagDTP int
	tagEB  int
	tagMSG int

	conEBPE, conEBTE, conEBEC, conEB, conREF int
	canPE, canTE, canEC, canRE               int
}

func (st *state271) nm1(b *InSolAut271, el []string) {
	switch st.tagHL {
	case "1":
		b.CaRemitente = at(el, 2)
		b.NuRucRemitente = at(el, 9)
	case "2":
		b.CaReceptor = at(el, 2)
	case "3":
		switch st.tagNM1 {
		case 0, 1:
			b.CaPaciente = at(el, 2)
			b.ApPaternoPaciente = at(el, 3)
			b.NoPaciente = at(el, 4)
			b.CoAfPaciente = at(el, 9)
			b.ApMaternoPaciente = at(el, 12)
		case 2:
			b.CaContratante = at(el, 2)
			b.NoPaContratante = at(el, 3)
			b.NoContratante = at(el, 4)
			b.NoMaContratante = at(el, 12)
		case 3:
			b.CaTitular = at(el, 2)
			b.NoPaTitular = at(el, 3)
			b.NoTitular = at(el, 4)
			b.CoAfTitular = at(el, 9)
			b.NoMaTitular = at(el, 12)
		case 4:
			b.CaPaciente = at(el, 2)
			b.ApPaternoPaciente = at(el, 3)
			b.NoPaciente = at(el, 4)
			b.CoAfPaciente = at(el, 9)
			b.ApMaternoPaciente = at(el, 12)
		}
		st.tagNM1++
	case "4":
		b.CaRegafi = at(el, 2)
		b.NoPaRegafi = at(el, 3)
		b.NoRegafi = at(el, 4)
		b.CoAfRegafi = at(el, 9)
		b.NoMaRegafi = at(el, 12)
		st.tagHL = "3"
		st.tagNM1 = 1
	}
}

func (st *state271) ref(b *InSolAut271, el []string) {
	switch st.tagREF {
	case 0, 1:
		b.CoAdmisionista = at(el, 2)
		st.tagREF = 2
	case 2:
		b.CoEsPaciente = at(el, 2)
		st.tagREF = 3
	case 3:
		b.TiDoPaciente = at(el, 2)
		b.NuDoPaciente = part(composite(el, 4), 1)
		st.tagREF = 4
	case 4:
		b.NuIdenEmpleador = at(el, 2)
		st.tagREF = 5
	case 5:
		b.NuContratoPaciente = at(el, 2)
		c := composite(el, 4)
		b.NuPoliza = part(c, 1)
		b.NuCertificado = part(c, 3)
		b.CoTiPolizaAfiliacion = part(c, 5)
		st.tagREF = 6
	case 6:
		b.CoProducto = at(el, 2)
		b.DeProducto = at(el, 3)
		st.tagREF = 7
	case 7:
		b.NuPlan = at(el, 2)
		c := composite(el, 4)
		b.TiPlanSalud = part(c, 1)
		b.CoMoneda = part(c, 3)
		st.tagREF = 8
	case 8:
		b.CoParentesco = at(el, 2)
		st.tagREF = 9
	case 9:
		b.SoBeneficio = at(el, 2)
		c := composite(el, 4)
		b.NuSoBeneficio = part(c, 1)
		b.CoEspecialidad = part(c, 3)
		st.tagREF = 10
	case 10:
		b.DeTiAccidente = at(el, 2)
		st.tagREF = 11
	case 11:
		b.TiProducto = at(el, 2)
		b.DeProductoDeFarmacia = at(el, 3)
		st.tagREF = 12
	case 12:
		b.TiDoContratante = at(el, 2)
		c := composite(el, 4)
		b.IdReContratante = part(c, 0)
		b.CoReContratante = part(c, 1)
		st.tagREF = 13
	case 13:
		b.TiDoTitular = at(el, 2)
		c := composite(el, 4)
		b.IdReTitular = part(c, 0)
		b.NuDoTitular = part(c, 1)
		st.tagREF = 14
	case 14:
		b.CoTiCobertura = at(el, 2)
		b.CoSubTiCobertura = part(composite(el, 4), 1)
		st.tagREF = 15
	case 15:
		if d := st.lastTE(b, false); d != nil {
			d.DeTiEspera = at(el, 2)
		}
		if st.conEBTE == st.canTE {
			st.tagREF++
		}
	case 16:
		if d := st.lastEC(b, false); d != nil {
			d.DeExCarencia = at(el, 2)
		}
		if st.conEBEC == st.canEC {
			st.tagREF++
		}
	case 17:
		if d := st.lastRE(b, false); d != nil {
			d.ObsRestricciones = at(el, 2)
		}
		st.conREF++
		st.tagREF = 18
	case 18:
		if d := st.lastRE(b, false); d != nil {
			d.DeRestricciones = at(el, 2)
		}
		if st.conREF < st.canRE {
			st.tagREF = 17
		} else {
			st.tagREF = 19
		}
	default:
		b.TiDoRegafi = at(el, 2)
		c := composite(el, 4)
		b.IdReRegafi = part(c, 0)
		b.NuDoRegafi = part(c, 1)
		st.tagREF = 1
	}
}

func (st *state271) dmg(b *InSolAut271, el []string) {
	if st.tagDMG == 0 || st.tagDMG == 1 {
		b.FeNacimiento = at(el, 2)
		b.Genero = at(el, 3)
		b.EsMarital = at(el, 4)
		st.tagDMG = 2
		return
	}
	b.FeNaRegafi = at(el, 2)
	b.GeRegafi = at(el, 3)
	b.CoPaisRegafi = at(el, 7)
	st.tagDMG = 1
}

func (st *state271) dtp(b *InSolAut271, el []string) {
	v := at(el, 3)
	switch st.tagDTP {
	case 0, 1:
		b.FeIniVigencia = v
	case 2:
		b.FeFinVigencia = v
	case 3:
		b.FeAfiliacion = v
	case 4:
		b.FeOcuAccidente = v
	case 5:
		b.FeAtencion = v
	case 6:
		b.FeIncTitular = v
	case 7:
		b.FeFinCarencia = v
	case 8:
		b.FeFinEspera = v
	case 9:
		if d := st.lastTE(b, false); d != nil {
			d.FeFinVigenciaTiEspera = v
		}
		if st.conEBTE == st.canTE {
			st.tagDTP++
		}
		return
	case 10:
		if d := st.lastRE(b, false); d != nil {
			d.FeFinEsperaRestricciones = v
		}
		return
	}
	st.tagDTP++
}

func (st *state271) eb(b *InSolAut271, el []string) {
	c := composite(el, 13)
	switch st.tagEB {
	case 0, 1:
		b.EsCobertura = at(el, 1)
		b.NuDecAccidente = at(el, 5)
		b.IdInfAccidente = part(c, 1)
		st.tagEB = 2
	case 2:
		b.NuAtencion = at(el, 5)
		b.IdDerFarmacia = part(c, 1)
		st.tagEB = 3
	case 3:
		b.NuCobertura = at(el, 5)
		b.ObsCobertura = part(c, 1)
		st.tagEB = 4
	case 4:
		b.NuCobPreExistencia = at(el, 5)
		b.BeMaxInicial = at(el, 7)
		b.CanServicio = at(el, 10)
		b.IdDeProducto = part(c, 1)
		st.tagEB = 5
	case 5:
		b.CoTiMoneda = at(el, 5)
		b.CoPagoFijo = at(el, 7)
		b.CoCalServicio = at(el, 9)
		b.CanCalServicio = at(el, 10)
		st.tagEB = 6
	case 6:
		b.CoPagoVariable = at(el, 8)
		st.tagEB = 7
	case 7:
		b.FlagCG = at(el, 1)
		b.DeflagCG = at(el, 5)
		st.tagEB = 8
	case 8:
		b.InSolAutProEsp271Detalle = append(b.InSolAutProEsp271Detalle, &InSolAutProEsp271Detalle{
			CoInProcedimiento: at(el, 9),
		})
		st.conEBPE++
		st.tagEB = 9
	case 9:
		d := st.lastPE(b, true)
		d.CoTiProConAmbulatoria = at(el, 3)
		d.NuPlanConAmbulatoria = at(el, 5)
		d.ImDeducible = at(el, 7)
		d.PoConAmbulatoria = at(el, 8)
		d.FrConAmbulatoria = at(el, 10)
		d.GeConAmbulatoria = part(c, 1)
		if st.conEBPE < st.canPE {
			st.tagEB = 8
		} else {
			st.tagEB = 10
		}
	case 10:
		b.InSolAutTieEsp271Detalle = append(b.InSolAutTieEsp271Detalle, &InSolAutTieEsp271Detalle{
			CoTiEspera: at(el, 3),
			IdTiEspera: part(c, 1),
		})
		st.conEBTE++
		if st.conEBTE < st.canTE {
			st.tagEB = 10
		} else {
			st.tagEB = 11
		}
	case 11:
		b.InSolAutExeCar271Detalle = append(b.InSolAutExeCar271Detalle, &InSolAutExeCar271Detalle{
			CoExCarencia: at(el, 3),
			IdExCarencia: part(c, 1),
		})
		st.conEBEC++
		if st.conEBEC < st.canEC {
			st.tagEB = 11
		} else {
			st.tagEB = 12
		}
	case 12:
		b.InSolAut271Detalle = append(b.InSolAut271Detalle, &InSolAutRes271Detalle{
			CIE10Restricciones: at(el, 3),
			IdRestricciones:    part(c, 1),
		})
		st.conEB++
		st.tagEB = 13
	case 13:
		if d := st.lastRE(b, false); d != nil {
			d.MonTopeRestricciones = at(el, 8)
		}
		if st.conEB < st.canRE {
			st.tagEB = 12
		}
	}
}

func (st *state271) msg(b *InSolAut271, el []string) {
	v := at(el, 1)
	switch st.tagMSG {
	case 1:
		b.MsgObs = v
		st.tagMSG++
	case 2:
		b.MsgConEspeciales = v
		st.tagMSG++
	case 3:
		b.MsgObsPre = v
		st.tagMSG++
	case 4:
		b.MsgConEspecialesPre = v
		st.tagMSG++
	case 5:
		// MSG repeats once per outpatient item, so the counter only advances
		// after the last one.
		if d := st.lastPE(b, false); d != nil {
			d.MsgConAmbulatoria = v
		}
		if st.conEBPE == st.canPE || st.canPE == 0 {
			st.tagMSG++
		}
	case 6:
		if d := st.lastTE(b, false); d != nil {
			d.MsgTiEspera = v
		}
		if st.conEBTE == st.canTE || st.canTE == 0 {
			st.tagMSG++
		}
	case 7:
		if d := st.lastEC(b, false); d != nil {
			d.MsgExCarencia = v
		}
		if st.conEBEC == st.canEC || st.canEC == 0 {
			st.tagMSG++
		}
	case 8:
		if d := st.lastRE(b, false); d != nil {
			d.MsgRestricciones = v
		}
	}
}

// lastPE returns the most recent outpatient-procedure item, creating one when
// create is set.
func (st *state271) lastPE(b *InSolAut271, create bool) *InSolAutProEsp271Detalle {
	if len(b.InSolAutProEsp271Detalle) == 0 {
		if !create {
			return nil
		}
		b.InSolAutProEsp271Detalle = append(b.InSolAutProEsp271Detalle, &InSolAutProEsp271Detalle{})
	}
	return b.InSolAutProEsp271Detalle[len(b.InSolAutProEsp271Detalle)-1]
}

func (st *state271) lastTE(b *InSolAut271, create bool) *InSolAutTieEsp271Detalle {
	if len(b.InSolAutTieEsp271Detalle) == 0 {
		if !create {
			return nil
		}
		b.InSolAutTieEsp271Detalle = append(b.InSolAutTieEsp271Detalle, &InSolAutTieEsp271Detalle{})
	}
	return b.InSolAutTieEsp271Detalle[len(b.InSolAutTieEsp271Detalle)-1]
}

func (st *state271) lastEC(b *InSolAut271, create bool) *InSolAutExeCar271Detalle {
	if len(b.InSolAutExeCar271Detalle) == 0 {
		if !create {
			return nil
		}
		b.InSolAutExeCar271Detalle = append(b.InSolAutExeCar271Detalle, &InSolAutExeCar271Detalle{})
	}
	return b.InSolAutExeCar271Detalle[len(b.InSolAutExeCar271Detalle)-1]
}

func (st *state271) lastRE(b *InSolAut271, create bool) *InSolAutRes271Detalle {
	if len(b.InSolAut271Detalle) == 0 {
		if !create {
			return nil
		}
		b.InSolAut271Detalle = append(b.InSolAut271Detalle, &InSolAutRes271Detalle{})
	}
	return b.InSolAut271Detalle[len(b.InSolAut271Detalle)-1]
}

// pruneEmpty drops detail items that carry no data, so the placeholder blocks
// emitted for empty lists are restored as empty lists and re-encoding is
// byte-identical.
func (st *state271) pruneEmpty(b *InSolAut271) {
	b.InSolAutProEsp271Detalle = prune(b.InSolAutProEsp271Detalle, func(d *InSolAutProEsp271Detalle) bool { return d.isZero() })
	b.InSolAutTieEsp271Detalle = prune(b.InSolAutTieEsp271Detalle, func(d *InSolAutTieEsp271Detalle) bool { return d.isZero() })
	b.InSolAutExeCar271Detalle = prune(b.InSolAutExeCar271Detalle, func(d *InSolAutExeCar271Detalle) bool { return d.isZero() })
	b.InSolAut271Detalle = prune(b.InSolAut271Detalle, func(d *InSolAutRes271Detalle) bool { return d.isZero() })
}

func prune[T any](in []*T, zero func(*T) bool) []*T {
	out := in[:0]
	for _, v := range in {
		if !zero(v) {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
