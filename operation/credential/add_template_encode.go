package credential

import (
	"github.com/imfact-labs/credential-model/types"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *AddTemplateFact) unpack(enc encoder.Encoder,
	sAdr, cAdr, tmplID string,
	tmplName, svcDate, expDate string,
	tmplShr, ma bool,
	dpName, subjKey, desc, crAdr, cid string,
) error {
	fact.templateName = tmplName
	fact.serviceDate = types.Date(svcDate)
	fact.expirationDate = types.Date(expDate)
	fact.templateShare = types.Bool(tmplShr)
	fact.multiAudit = types.Bool(ma)
	fact.displayName = dpName
	fact.subjectKey = subjKey
	fact.description = desc
	fact.currency = ctypes.CurrencyID(cid)
	fact.templateID = tmplID

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

	switch a, err := base.DecodeAddress(crAdr, enc); {
	case err != nil:
		return err
	default:
		fact.creator = a
	}

	return nil
}
