package steps

import (
	"context"

	"github.com/imfact-labs/credential-model/operation/credential"
	"github.com/imfact-labs/credential-model/runtime/contracts"
	cprocessor "github.com/imfact-labs/currency-model/operation/processor"
	ctype "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/ps"
)

var PNameOperationProcessorsMap = ps.Name("mitum-credential-operation-processors-map")

type processorInfo struct {
	hint      hint.Hint
	processor ctype.GetNewProcessor
}

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database
	var opr *cprocessor.OperationProcessor
	var set *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
		contracts.OperationProcessorContextKey, &opr,
		launch.OperationProcessorsMapContextKey, &set,
	); err != nil {
		return pctx, err
	}

	err := opr.SetGetNewProcessorFunc(cprocessor.GetNewProcessor)
	if err != nil {
		return pctx, err
	}

	processors := []processorInfo{
		{credential.RegisterModelHint, credential.NewRegisterModelProcessor()},
		{credential.AddTemplateHint, credential.NewAddTemplateProcessor()},
		{credential.IssueHint, credential.NewIssueProcessor()},
		{credential.RevokeHint, credential.NewRevokeProcessor()},
	}

	for i := range processors {
		p := processors[i]

		if err := opr.SetProcessor(p.hint, p.processor); err != nil {
			return pctx, err
		}

		if err := set.Add(p.hint,
			func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
				return opr.New(
					height,
					getStatef,
					nil,
					nil,
				)
			},
		); err != nil {
			return pctx, err
		}
	}

	pctx = context.WithValue(pctx, contracts.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, set) //revive:disable-line:modifies-parameter

	return pctx, nil
}
