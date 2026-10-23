package x12n

import (
	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// The parsers below read the interchange back into the same bean the encoder
// consumes. Control numbers (GS06/ST02) are intentionally not restored: the
// wire form is zero-padded, so the original value cannot be recovered, and
// they are marshalling inputs rather than business data. Supply them through
// Options when a caller needs to reproduce a specific interchange.

func segmentsOf(s string) ([]string, error) {
	segs := edi.SplitSegments(s)
	if len(segs) == 0 {
		return nil, &edi.ParseError{Segment: "interchange", Reason: "empty input"}
	}
	return segs, nil
}

// Unmarshal997 parses a 997 acknowledgement into an In997ResAut.
func Unmarshal997(s string) (*In997ResAut, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}

	b := &In997ResAut{}
	for _, raw := range segs {
		id, el := edi.ParseSegment(raw)
		switch id {
		case "ISA":
			b.NoTransaccion = "997_RES_AUT"
			b.IdRemitente = at(el, 6)
			b.IdReceptor = at(el, 8)
			b.IdCorrelativo = at(el, 13)
		case "GS":
			b.FeTransaccion = at(el, 4)
			b.HoTransaccion = at(el, 5)
		case "ST":
			b.IdTransaccion = at(el, 1)
		case "AK1":
			b.NuAutorizacion = at(el, 3)
		case "AK2":
			b.CoSeguridad = at(el, 3)
		case "AK5":
			b.CoError = at(el, 1)
			b.InCoErrorEncontrado = at(el, 2)
		case "AK9":
			b.CoError = at(el, 1)
		}
	}
	return b, nil
}

// UnmarshalConEntVinc278 parses a 278 entity-linkage enquiry.
func UnmarshalConEntVinc278(s string) (*InConEntVinc278, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}

	b := &InConEntVinc278{}
	for _, raw := range segs {
		id, el := edi.ParseSegment(raw)
		switch id {
		case "ISA":
			b.NoTransaccion = "278_CON_ENT_VINC"
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
		case "NM1":
			b.CaIPRESS = at(el, 2)
			b.NoIPRESS = at(el, 4)
		case "REF":
			b.TiDoIPRESS = at(el, 2)
			b.NuRucIPRESS = part(composite(el, 4), 1)
		}
	}
	return b, nil
}

// UnmarshalResEntVinc278 parses the 278 entity-linkage reply.
func UnmarshalResEntVinc278(s string) (*InResEntVinc278, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}

	b := &InResEntVinc278{}
	for _, raw := range segs {
		id, el := edi.ParseSegment(raw)
		switch id {
		case "ISA":
			b.NoTransaccion = "278_RES_ENT_VINC"
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
		case "CRC":
			b.Respuesta = at(el, 2)
		case "MSG":
			b.MsgRespuesta = at(el, 1)
		}
	}
	return b, nil
}

// UnmarshalIn278SolCG parses a 278 certification request.
func UnmarshalIn278SolCG(s string) (*In278SolCG, error) {
	segs, err := segmentsOf(s)
	if err != nil {
		return nil, err
	}

	b := &In278SolCG{}
	refs := 0
	for _, raw := range segs {
		id, el := edi.ParseSegment(raw)
		switch id {
		case "ISA":
			b.NoTransaccion = "278_SOL_CG"
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
		case "NM1":
			b.CaPaciente = at(el, 2)
			b.ApPaternoPaciente = at(el, 3)
			b.NoPaciente = at(el, 4)
			b.ApMaternoPaciente = at(el, 12)
		case "REF":
			refs++
			c := composite(el, 4)
			switch refs {
			case 1:
				b.CoTiDoPaciente = at(el, 2)
				b.NuDoPaciente = part(c, 1)
			case 2:
				b.NuCarGarantia = at(el, 2)
				b.EsCarGarantia = part(c, 1)
			case 3:
				b.NuSoCarGarantia = part(c, 1)
			}
		}
	}
	return b, nil
}
