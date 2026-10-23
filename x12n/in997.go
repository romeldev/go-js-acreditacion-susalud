package x12n

import (
	"errors"

	"github.com/romeldev/go-js-acreditacion-susalud/edi"
)

// Marshal997 renders an In997ResAut as a 997 acknowledgement transaction set.
//
// It is the Go equivalent of In997ResAut_ToXs12N.traducirEstructura997ResAut.
func Marshal997(b *In997ResAut, o *Options) (string, error) {
	if b == nil {
		return "", errors.New("x12n: nil In997ResAut")
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

	// The 997 omits the transaction-set count in ST03.
	if err := e.open(header{
		SenderID:    b.IdRemitente,
		ReceiverID:  b.IdReceptor,
		Date:        b.FeTransaccion,
		Time:        b.HoTransaccion,
		Correlative: b.IdCorrelativo,
		Transaction: b.IdTransaccion,
	}, "FA", ctrl, ctrlST, "", o.Production); err != nil {
		return "", err
	}

	e.addStd(edi.New(edi.SegAK1).
		Set(1, "PO").
		Set(2, ctrl).
		Set(3, b.NuAutorizacion))

	e.addStd(edi.New(edi.SegAK2).
		Set(1, b.IdTransaccion).
		Set(2, ctrlST).
		Set(3, b.CoSeguridad))

	e.addStd(edi.New(edi.SegAK5).
		Set(1, b.CoError).
		Set(2, b.InCoErrorEncontrado))

	e.addStd(edi.New(edi.SegAK9).
		Set(1, b.CoError).
		Set(2, "1").
		Set(3, "1").
		Set(4, "1"))

	if err := e.close(ctrl, ctrlST, b.IdCorrelativo); err != nil {
		return "", err
	}
	return e.string(), nil
}
