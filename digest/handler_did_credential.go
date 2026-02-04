package digest

import (
	"net/http"
	"time"

	"github.com/ProtoconNet/mitum-credential/types"
	cdigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var (
	HandlerPathDIDService     = `/did/{contract:(?i)` + ctypes.REStringAddressString + `}`
	HandlerPathDIDCredential  = `/did/{contract:(?i)` + ctypes.REStringAddressString + `}/template/{template_id:` + ctypes.ReSpecialCh + `}/credential/{credential_id:` + ctypes.ReSpecialCh + `}`
	HandlerPathDIDTemplate    = `/did/{contract:(?i)` + ctypes.REStringAddressString + `}/template/{template_id:` + ctypes.ReSpecialCh + `}`
	HandlerPathDIDCredentials = `/did/{contract:(?i)` + ctypes.REStringAddressString + `}/template/{template_id:` + ctypes.ReSpecialCh + `}/credentials`
	HandlerPathDIDHolder      = `/did/{contract:(?i)` + ctypes.REStringAddressString + `}/holder/{holder:(?i)` + ctypes.REStringAddressString + `}` // revive:disable-line:line-length-limit
)

func SetHandlers(hd *cdigest.Handlers) {
	get := 1000
	_ = hd.SetHandler(HandlerPathDIDService, HandleCredentialService, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathDIDCredentials, HandleCredentials, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathDIDCredential, HandleCredential, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathDIDHolder, HandleHolderCredential, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathDIDTemplate, HandleTemplate, true, get, get).
		Methods(http.MethodOptions, "GET")
}

func HandleCredentialService(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)

		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleCredentialServiceInGroup(hd, contract)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleCredentialServiceInGroup(hd *cdigest.Handlers, contract string) (interface{}, error) {
	switch design, err := CredentialService(hd.Database(), contract); {
	case err != nil:
		return nil, util.ErrNotFound.WithMessage(err, "credential design, contract %s", contract)
	case design == nil:
		return nil, util.ErrNotFound.Errorf("credential design, contract %s", contract)
	default:
		hal, err := buildCredentialServiceHal(hd, contract, *design)
		if err != nil {
			return nil, err
		}
		return hd.Encoder().Marshal(hal)
	}
}

func buildCredentialServiceHal(hd *cdigest.Handlers, contract string, design types.Design) (cdigest.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDIDService, "contract", contract)
	if err != nil {
		return nil, err
	}

	hal := cdigest.NewBaseHal(design, cdigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleCredential(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	templateID, err, status := cdigest.ParseRequest(w, r, "template_id")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	credentialID, err, status := cdigest.ParseRequest(w, r, "credential_id")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleCredentialInGroup(hd, contract, templateID, credentialID)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleCredentialInGroup(hd *cdigest.Handlers, contract, templateID, credentialID string) (interface{}, error) {
	switch credential, isActive, err := Credential(hd.Database(), contract, templateID, credentialID); {
	case err != nil:
		return nil, util.ErrNotFound.WithMessage(err, "credential by contract %s, template %s, id %s", contract, templateID, credentialID)
	case credential == nil:
		return nil, util.ErrNotFound.Errorf("credential by contract %s, template %s, id %s", contract, templateID, credentialID)
	default:
		hal, err := buildCredentialHal(hd, contract, *credential, isActive)
		if err != nil {
			return nil, err
		}
		return hd.Encoder().Marshal(hal)
	}
}

func buildCredentialHal(
	hd *cdigest.Handlers,
	contract string,
	credential types.Credential,
	isActive bool,
) (cdigest.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDIDCredential,
		"contract", contract,
		"template_id", credential.TemplateID(),
		"credential_id", credential.CredentialID(),
	)
	if err != nil {
		return nil, err
	}

	hal := cdigest.NewBaseHal(
		struct {
			Credential types.Credential `json:"credential"`
			IsActive   bool             `json:"is_active"`
		}{Credential: credential, IsActive: isActive},
		cdigest.NewHalLink(h, nil),
	)

	return hal, nil
}

func HandleCredentials(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	limit := cdigest.ParseLimitQuery(r.URL.Query().Get("limit"))
	offset := cdigest.ParseStringQuery(r.URL.Query().Get("offset"))
	reverse := cdigest.ParseBoolQuery(r.URL.Query().Get("reverse"))

	cachekey := cdigest.CacheKey(
		r.URL.Path, cdigest.StringOffsetQuery(offset),
		cdigest.StringBoolQuery("reverse", reverse),
	)

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)

		return
	}

	templateID, err, status := cdigest.ParseRequest(w, r, "template_id")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	v, err, shared := hd.RG().Do(cachekey, func() (interface{}, error) {
		i, filled, err := handleCredentialsInGroup(hd, contract, templateID, offset, reverse, limit)

		return []interface{}{i, filled}, err
	})

	if err != nil {
		hd.Log().Err(err).Str("Issuer", contract).Msg("failed to get credentials")
		cdigest.HTTP2HandleError(w, err)

		return
	}

	var b []byte
	var filled bool
	{
		l := v.([]interface{})
		b = l[0].([]byte)
		filled = l[1].(bool)
	}

	cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, b, http.StatusOK)

	if !shared {
		expire := hd.ExpireNotFilled()
		if len(offset) > 0 && filled {
			expire = time.Minute
		}

		cdigest.HTTP2WriteCache(w, cachekey, expire)
	}
}

func handleCredentialsInGroup(
	hd *cdigest.Handlers,
	contract, templateID string,
	offset string,
	reverse bool,
	l int64,
) ([]byte, bool, error) {
	var limit int64
	if l < 0 {
		limit = hd.ItemsLimiter("service-credentials")
	} else {
		limit = l
	}

	var vas []cdigest.Hal
	if err := CredentialsByServiceTemplate(
		hd.Database(), contract, templateID, reverse, offset, limit,
		func(credential types.Credential, isActive bool, st base.State) (bool, error) {
			hal, err := buildCredentialHal(hd, contract, credential, isActive)
			if err != nil {
				return false, err
			}
			vas = append(vas, hal)

			return true, nil
		},
	); err != nil {
		return nil, false, util.ErrNotFound.WithMessage(err, "credentials by contract %s, template %s", contract, templateID)
	} else if len(vas) < 1 {
		return nil, false, util.ErrNotFound.Errorf("credentials by contract %s, template %s", contract, templateID)
	}

	i, err := buildCredentialsHal(hd, contract, templateID, vas, offset, reverse)
	if err != nil {
		return nil, false, err
	}

	b, err := hd.Encoder().Marshal(i)
	return b, int64(len(vas)) == limit, err
}

func buildCredentialsHal(
	hd *cdigest.Handlers,
	contract, templateID string,
	vas []cdigest.Hal,
	offset string,
	reverse bool,
) (cdigest.Hal, error) {
	baseSelf, err := hd.CombineURL(
		HandlerPathDIDCredentials,
		"contract", contract,
		"template_id", templateID,
	)
	if err != nil {
		return nil, err
	}

	self := baseSelf
	if len(offset) > 0 {
		self = cdigest.AddQueryValue(baseSelf, cdigest.StringOffsetQuery(offset))
	}
	if reverse {
		self = cdigest.AddQueryValue(baseSelf, cdigest.StringBoolQuery("reverse", reverse))
	}

	var hal cdigest.Hal
	hal = cdigest.NewBaseHal(vas, cdigest.NewHalLink(self, nil))

	h, err := hd.CombineURL(HandlerPathDIDService, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("service", cdigest.NewHalLink(h, nil))

	var nextOffset string

	if len(vas) > 0 {
		va, ok := vas[len(vas)-1].Interface().(struct {
			Credential types.Credential `json:"credential"`
			IsActive   bool             `json:"is_active"`
		})
		if !ok {
			return nil, errors.Errorf("failed to build credentials hal")
		}
		nextOffset = va.Credential.CredentialID()
	}

	if len(nextOffset) > 0 {
		next := baseSelf
		next = cdigest.AddQueryValue(next, cdigest.StringOffsetQuery(nextOffset))

		if reverse {
			next = cdigest.AddQueryValue(next, cdigest.StringBoolQuery("reverse", reverse))
		}

		hal = hal.AddLink("next", cdigest.NewHalLink(next, nil))
	}

	hal = hal.AddLink("reverse", cdigest.NewHalLink(cdigest.AddQueryValue(baseSelf, cdigest.StringBoolQuery("reverse", !reverse)), nil))

	return hal, nil
}

func HandleHolderCredential(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	holder, err, status := cdigest.ParseRequest(w, r, "holder")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleHolderCredentialsInGroup(hd, contract, holder)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleHolderCredentialsInGroup(hd *cdigest.Handlers, contract, holder string) (interface{}, error) {
	var did string
	switch d, err := HolderDID(hd.Database(), contract, holder); {
	case err != nil:
		return nil, util.ErrNotFound.WithMessage(err, "DID by contract %s, holder %s", contract, holder)
	case d == "":
		return nil, util.ErrNotFound.Errorf("DID by contract %s, holder %s", contract, holder)
	default:
		did = d
	}

	var vas []cdigest.Hal
	if err := CredentialsByServiceHolder(
		hd.Database(), contract, holder,
		func(credential types.Credential, isActive bool, st base.State) (bool, error) {
			hal, err := buildCredentialHal(hd, contract, credential, isActive)
			if err != nil {
				return false, err
			}
			vas = append(vas, hal)

			return true, nil
		},
	); err != nil {
		return nil, util.ErrNotFound.WithMessage(err, "credentials by contract %s, holder %s", contract, holder)
	} else if len(vas) < 1 {
		return nil, util.ErrNotFound.Errorf("credentials by contract %s, holder %s", contract, holder)
	}

	hal, err := buildHolderDIDCredentialsHal(hd, contract, holder, did, vas)
	if err != nil {
		return nil, err
	}
	return hd.Encoder().Marshal(hal)
}

func buildHolderDIDCredentialsHal(
	hd *cdigest.Handlers,
	contract, holder, did string,
	vas []cdigest.Hal,
) (cdigest.Hal, error) {
	baseSelf, err := hd.CombineURL(HandlerPathDIDHolder, "contract", contract, "holder", holder)
	if err != nil {
		return nil, err
	}

	self := baseSelf

	var hal cdigest.Hal
	hal = cdigest.NewBaseHal(
		struct {
			DID         string        `json:"did"`
			Credentials []cdigest.Hal `json:"credentials"`
		}{
			DID:         did,
			Credentials: vas,
		}, cdigest.NewHalLink(self, nil))

	h, err := hd.CombineURL(HandlerPathDIDService, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("service", cdigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleTemplate(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	templateID, err, status := cdigest.ParseRequest(w, r, "template_id")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleTemplateInGroup(hd, contract, templateID)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireLongLived())
		}
	}
}

func handleTemplateInGroup(hd *cdigest.Handlers, contract, templateID string) (interface{}, error) {
	switch template, err := Template(hd.Database(), contract, templateID); {
	case err != nil:
		return nil, util.ErrNotFound.WithMessage(err, "template by contract %s, template %s", contract, templateID)
	case template == nil:
		return nil, util.ErrNotFound.Errorf("template by contract %s, template %s", contract, templateID)
	default:
		hal, err := buildTemplateHal(hd, contract, templateID, *template)
		if err != nil {
			return nil, err
		}
		return hd.Encoder().Marshal(hal)
	}
}

func buildTemplateHal(
	hd *cdigest.Handlers,
	contract, templateID string,
	template types.Template,
) (cdigest.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDIDTemplate,
		"contract", contract,
		"template_id", templateID,
	)
	if err != nil {
		return nil, err
	}

	hal := cdigest.NewBaseHal(template, cdigest.NewHalLink(h, nil))

	return hal, nil
}
