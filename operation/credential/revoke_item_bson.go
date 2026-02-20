package credential // nolint:dupl

import (
	"github.com/imfact-labs/currency-model/common"
	bsonenc "github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (it RevokeItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":       it.Hint().String(),
			"contract":    it.contract,
			"holder":      it.holder,
			"template_id": it.templateID,
			"id":          it.credentialID,
			"currency":    it.currency,
		},
	)
}

type RevokeItemBSONUnmarshaler struct {
	Hint         string `bson:"_hint"`
	Contract     string `bson:"contract"`
	Holder       string `bson:"holder"`
	TemplateID   string `bson:"template_id"`
	CredentialID string `bson:"credential_id"`
	Currency     string `bson:"currency"`
}

func (it *RevokeItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var uit RevokeItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	if err := it.unpack(enc, ht,
		uit.Contract,
		uit.Holder,
		uit.TemplateID,
		uit.CredentialID,
		uit.Currency,
	); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	return nil
}
