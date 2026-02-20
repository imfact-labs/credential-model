package types

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

var DesignHint = hint.MustNewHint("mitum-credential-design-v0.0.1")

type Design struct {
	hint.BaseHinter
	policy Policy
}

func NewDesign(policy Policy) Design {
	return Design{
		BaseHinter: hint.NewBaseHinter(DesignHint),
		policy:     policy,
	}
}

func (de Design) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false,
		de.BaseHinter,
		de.policy,
	); err != nil {
		return common.ErrValueInvalid.Wrap(errors.Errorf("design: %v", err))
	}

	return nil
}

func (de Design) Bytes() []byte {
	return util.ConcatBytesSlice(
		de.policy.Bytes(),
	)
}

func (de Design) Policy() Policy {
	return de.policy
}
