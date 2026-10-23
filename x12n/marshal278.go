package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// MarshalConEntVinc278 renders an InConEntVinc278 as a 278 enquiry for entity
// linkage.
//
// It is the Go equivalent of ConEntVinc278_ToX12N.traducirEstructura271 (the
// reference names every 278 translator traducirEstructura271).
func MarshalConEntVinc278(b *InConEntVinc278, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InConEntVinc278")
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
		SenderID:    b.IdRemitente,
		ReceiverID:  b.IdReceptor,
		Date:        b.FeTransaccion,
		Time:        b.HoTransaccion,
		Correlative: b.IdCorrelativo,
		Transaction: b.IdTransaccion,
	}, "HB", ctrl, ctrlST, "", o.Production); err != nil {
		return "", err
	}

	e.addStd(edi.New(edi.SegBHT).Set(1, "0020").Set(2, b.TiFinalidad))
	e.addStd(edi.New(edi.SegHL).Set(1, "1").Set(3, "20").Set(4, "0"))
	e.addStd(nm1Person("IL", b.CaIPRESS, "", b.NoIPRESS, "", "", ""))
	e.addNested(ref("DD", b.TiDoIPRESS), ref4("XX5", b.NuRucIPRESS))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// MarshalResEntVinc278 renders an InResEntVinc278 as the 278 reply.
//
// It is the Go equivalent of ResEntVinc278_ToX12N.traducirEstructura271.
func MarshalResEntVinc278(b *InResEntVinc278, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil InResEntVinc278")
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
		SenderID:    b.IdRemitente,
		ReceiverID:  b.IdReceptor,
		Date:        b.FeTransaccion,
		Time:        b.HoTransaccion,
		Correlative: b.IdCorrelativo,
		Transaction: b.IdTransaccion,
	}, "HB", ctrl, ctrlST, "", o.Production); err != nil {
		return "", err
	}

	e.addStd(edi.New(edi.SegBHT).Set(1, "0020").Set(2, b.TiFinalidad))
	e.addStd(edi.New(edi.SegHL).Set(1, "1").Set(3, "20").Set(4, "0"))
	e.addStd(edi.New(edi.SegCRC).Set(1, "ZZ").Set(2, b.Respuesta))
	e.addStd(edi.New(edi.SegMSG).Set(1, b.MsgRespuesta))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}

// MarshalIn278SolCG renders an In278SolCG as a 278 certification request.
//
// It is the Go equivalent of In278SolCG_ToX12N.traducirEstructura278Sol.
func MarshalIn278SolCG(b *In278SolCG, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil In278SolCG")
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
		SenderID:    b.IdRemitente,
		ReceiverID:  b.IdReceptor,
		Date:        b.FeTransaccion,
		Time:        b.HoTransaccion,
		Correlative: b.IdCorrelativo,
		Transaction: b.IdTransaccion,
	}, "HI", ctrl, ctrlST, "", o.Production); err != nil {
		return "", err
	}

	e.addStd(edi.New(edi.SegBHT).Set(1, "0020").Set(2, b.TiFinalidad))
	e.addStd(edi.New(edi.SegHL).Set(1, "1").Set(3, "20").Set(4, "0"))
	e.addStd(nm1Person("IL", b.CaPaciente, b.ApPaternoPaciente, b.NoPaciente, "", "", b.ApMaternoPaciente))
	e.addNested(ref("DD", b.CoTiDoPaciente), ref4("4A", b.NuDoPaciente))
	e.addNested(ref("BB", b.NuCarGarantia), ref4("ACC", b.EsCarGarantia))
	e.addNested(ref("5F", ""), ref4("ZZ", b.NuSoCarGarantia))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}
