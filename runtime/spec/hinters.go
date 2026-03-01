package spec

import (
	"github.com/imfact-labs/credential-model/operation/credential"
	"github.com/imfact-labs/credential-model/state"
	"github.com/imfact-labs/credential-model/types"
	"github.com/imfact-labs/mitum2/util/encoder"
)

var AddedHinters = []encoder.DecodeDetail{
	// revive:disable-next-line:line-length-limit
	{Hint: types.CredentialHint, Instance: types.Credential{}},
	{Hint: types.DesignHint, Instance: types.Design{}},
	{Hint: types.HolderHint, Instance: types.Holder{}},
	{Hint: types.PolicyHint, Instance: types.Policy{}},
	{Hint: types.TemplateHint, Instance: types.Template{}},

	{Hint: credential.RegisterModelHint, Instance: credential.RegisterModel{}},
	{Hint: credential.AddTemplateHint, Instance: credential.AddTemplate{}},
	{Hint: credential.IssueItemHint, Instance: credential.IssueItem{}},
	{Hint: credential.IssueHint, Instance: credential.Issue{}},
	{Hint: credential.RevokeItemHint, Instance: credential.RevokeItem{}},
	{Hint: credential.RevokeHint, Instance: credential.Revoke{}},

	{Hint: state.CredentialStateValueHint, Instance: state.CredentialStateValue{}},
	{Hint: state.DesignStateValueHint, Instance: state.DesignStateValue{}},
	{Hint: state.HolderDIDStateValueHint, Instance: state.HolderDIDStateValue{}},
	{Hint: state.TemplateStateValueHint, Instance: state.TemplateStateValue{}},
}

var AddedSupportedHinters = []encoder.DecodeDetail{
	{Hint: credential.AddTemplateFactHint, Instance: credential.AddTemplateFact{}},
	{Hint: credential.IssueFactHint, Instance: credential.IssueFact{}},
	{Hint: credential.RegisterModelFactHint, Instance: credential.RegisterModelFact{}},
	{Hint: credential.RevokeFactHint, Instance: credential.RevokeFact{}},
}
