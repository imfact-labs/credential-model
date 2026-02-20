package digest

import (
	"context"

	"github.com/imfact-labs/credential-model/state"
	"github.com/imfact-labs/credential-model/types"
	cdigest "github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/currency-model/digest/util"
	"github.com/imfact-labs/mitum2/base"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	DefaultColNameDIDCredentialService = "digest_did_issuer"
	DefaultColNameDIDCredential        = "digest_did_credential"
	DefaultColNameHolder               = "digest_did_holder_did"
	DefaultColNameTemplate             = "digest_did_template"
)

var maxLimit int64 = 50

func CredentialService(st *cdigest.Database, contract string) (*types.Design, error) {
	filter := util.NewBSONFilter("contract", contract)

	var design *types.Design
	var sta base.State
	var err error
	if err := st.MongoClient().GetByFilter(
		DefaultColNameDIDCredentialService,
		filter.D(),
		func(res *mongo.SingleResult) error {
			sta, err = cdigest.LoadState(res.Decode, st.Encoders())
			if err != nil {
				return err
			}

			de, err := state.StateDesignValue(sta)
			if err != nil {
				return err
			}
			design = &de

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		return nil, err
	}

	return design, nil
}

func Credential(st *cdigest.Database, contract, templateID, credentialID string) (*types.Credential, bool, error) {
	filter := util.NewBSONFilter("contract", contract)
	filter = filter.Add("template", templateID)
	filter = filter.Add("credential_id", credentialID)

	var credential *types.Credential
	var isActive bool
	var sta base.State
	var err error
	if err = st.MongoClient().GetByFilter(
		DefaultColNameDIDCredential,
		filter.D(),
		func(res *mongo.SingleResult) error {
			sta, err = cdigest.LoadState(res.Decode, st.Encoders())
			if err != nil {
				return err
			}
			cre, active, err := state.StateCredentialValue(sta)
			if err != nil {
				return err
			}
			credential = &cre
			isActive = active
			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		return nil, false, err
	}

	return credential, isActive, nil
}

func Template(st *cdigest.Database, contract, templateID string) (*types.Template, error) {
	filter := util.NewBSONFilter("contract", contract)
	filter = filter.Add("template", templateID)

	var template *types.Template
	var sta base.State
	var err error
	if err = st.MongoClient().GetByFilter(
		DefaultColNameTemplate,
		filter.D(),
		func(res *mongo.SingleResult) error {
			sta, err = cdigest.LoadState(res.Decode, st.Encoders())
			if err != nil {
				return err
			}
			te, err := state.StateTemplateValue(sta)
			if err != nil {
				return err
			}
			template = &te
			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		return nil, err
	}

	return template, nil
}

func HolderDID(st *cdigest.Database, contract, holder string) (string, error) {
	filter := util.NewBSONFilter("contract", contract)
	filter = filter.Add("holder", holder)

	var did string
	var sta base.State
	var err error
	if err = st.MongoClient().GetByFilter(
		DefaultColNameHolder,
		filter.D(),
		func(res *mongo.SingleResult) error {
			sta, err = cdigest.LoadState(res.Decode, st.Encoders())
			if err != nil {
				return err
			}
			did, err = state.StateHolderDIDValue(sta)
			if err != nil {
				return err
			}

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		return "", err
	}

	return did, nil
}

func CredentialsByServiceTemplate(
	st *cdigest.Database,
	contract,
	templateID string,
	reverse bool,
	offset string,
	limit int64,
	callback func(types.Credential, bool, base.State) (bool, error),
) error {
	sortDir := 1
	cmpOp := "$gt"
	if reverse {
		sortDir = -1
		cmpOp = "$lt"
	}

	match := bson.D{
		{Key: "contract", Value: contract},
		{Key: "template", Value: templateID},
	}

	if offset != "" {
		match = append(match, bson.E{
			Key:   "credential_id",
			Value: bson.D{{Key: cmpOp, Value: offset}},
		})
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "credential_id", Value: sortDir},
			{Key: "height", Value: -1},
			{Key: "_id", Value: -1},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$credential_id"},
			{Key: "doc", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
		}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{
			{Key: "newRoot", Value: "$doc"},
		}}},
	}

	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}

	return st.MongoClient().Aggregate(
		context.Background(),
		DefaultColNameDIDCredential,
		pipeline,
		func(cursor *mongo.Cursor) (bool, error) {
			st, err := cdigest.LoadState(cursor.Decode, st.Encoders())
			if err != nil {
				return false, err
			}
			credential, isActive, err := state.StateCredentialValue(st)
			if err != nil {
				return false, err
			}
			return callback(credential, isActive, st)
		},
	)
}

func CredentialsByServiceHolder(
	st *cdigest.Database,
	contract, holder string,
	callback func(types.Credential, bool, base.State) (bool, error),
) error {
	match := bson.D{
		{Key: "contract", Value: contract},
		{Key: "d.value.credential.holder", Value: holder},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "credential_id", Value: 1},
			{Key: "template", Value: 1},
			{Key: "height", Value: -1},
			{Key: "_id", Value: -1},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "credential_id", Value: "$credential_id"},
				{Key: "template", Value: "$template"},
			}},
			{Key: "doc", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
		}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{
			{Key: "newRoot", Value: "$doc"},
		}}},
	}

	return st.MongoClient().Aggregate(
		context.Background(),
		DefaultColNameDIDCredential,
		pipeline,
		func(cursor *mongo.Cursor) (bool, error) {
			st, err := cdigest.LoadState(cursor.Decode, st.Encoders())
			if err != nil {
				return false, err
			}
			credential, isActive, err := state.StateCredentialValue(st)
			if err != nil {
				return false, err
			}
			return callback(credential, isActive, st)
		},
	)
}

func buildCredentialFilterByServiceHolder(contract, holder string) (bson.D, error) {
	filterA := bson.A{}

	// filter fot matching collection
	filterContract := bson.D{{"contract", bson.D{{"$in", []string{contract}}}}}
	filterHolder := bson.D{{"d.value.credential.holder", holder}}
	filterA = append(filterA, filterContract)
	filterA = append(filterA, filterHolder)

	filter := bson.D{}
	if len(filterA) > 0 {
		filter = bson.D{
			{"$and", filterA},
		}
	}

	return filter, nil
}
