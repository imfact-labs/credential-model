package credential

import (
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *RegisterModelFact) unpack(enc encoder.Encoder, sAdr, cAdr, cid string) error {
	switch a, err := base.DecodeAddress(sAdr, enc); {
	case err != nil:
		return err
	default:
		fact.sender = a
	}

	switch a, err := base.DecodeAddress(cAdr, enc); {
	case err != nil:
		return err
	default:
		fact.contract = a
	}

	fact.currency = ctypes.CurrencyID(cid)

	return nil
}
