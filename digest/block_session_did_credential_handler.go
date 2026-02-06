package digest

import (
	"github.com/ProtoconNet/mitum-credential/state"
	cdigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum2/base"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func PrepareDIDCredential(bs *cdigest.BlockSession, st base.State) (string, []mongo.WriteModel, error) {
	switch {
	case state.IsStateDesignKey(st.Key()):
		j, err := handleDIDCredentialDesignState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameDIDCredentialService, j, nil
	case state.IsStateCredentialKey(st.Key()):
		j, err := handleCredentialState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameDIDCredential, j, nil

	case state.IsStateHolderDIDKey(st.Key()):
		j, err := handleHolderDIDState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameHolder, j, nil
	case state.IsStateTemplateKey(st.Key()):
		j, err := handleTemplateState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameTemplate, j, nil
	}

	return "", nil, nil
}

func handleDIDCredentialDesignState(bs *cdigest.BlockSession, st base.State) ([]mongo.WriteModel, error) {
	if issuerDoc, err := NewDIDCredentialDesignDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(issuerDoc),
		}, nil
	}
}

func handleCredentialState(bs *cdigest.BlockSession, st base.State) ([]mongo.WriteModel, error) {
	if credentialDoc, err := NewCredentialDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(credentialDoc),
		}, nil
	}
}

func handleHolderDIDState(bs *cdigest.BlockSession, st base.State) ([]mongo.WriteModel, error) {
	if holderDidDoc, err := NewHolderDIDDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(holderDidDoc),
		}, nil
	}
}

func handleTemplateState(bs *cdigest.BlockSession, st base.State) ([]mongo.WriteModel, error) {
	if templateDoc, err := NewTemplateDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(templateDoc),
		}, nil
	}
}
