package kucoin

import (
	"context"
	"github.com/rasteiro11/PogCore/pkg/config"
	"net/http"
	"net/url"
	"time"
)

type KucoinClient interface {
	GetWithdrawalsQuota(ctx context.Context, req *GetWithdrawalsQuotaRequest) (*GetWithdrawalsQuotaResponse, error)
}

type Kucoin struct {
	authz  *Authz
	client *http.Client
	url    *url.URL
}

type GetWithdrawalsQuotaRequest struct {
	Currency string
}

type GetWithdrawalsQuotaResponseData struct {
	Currency                 string `json:"currency"`
	LimitBTCAmount           string `json:"limitBTCAmount"`
	UsedBTCAmount            string `json:"usedBTCAmount"`
	QuotaCurrency            string `json:"quotaCurrency"`
	LimitQuotaCurrencyAmount string `json:"limitQuotaCurrencyAmount"`
	UsedQuotaCurrencyAmount  string `json:"usedQuotaCurrencyAmount"`
	RemainAmount             string `json:"remainAmount"`
	AvailableAmount          string `json:"availableAmount"`
	WithdrawMinFee           string `json:"withdrawMinFee"`
	InnerWithdrawMinFee      string `json:"innerWithdrawMinFee"`
	WithdrawMinSize          string `json:"withdrawMinSize"`
	IsWithdrawEnabled        bool   `json:"isWithdrawEnabled"`
	Precision                int    `json:"precision"`
	Chain                    string `json:"chain"`
	Reason                   string `json:"reason"`
	LockedAmount             string `json:"lockedAmount"`
}

type GetWithdrawalsQuotaResponse struct {
	Code string                          `json:"code"`
	Data GetWithdrawalsQuotaResponseData `json:"data"`
}

func (k *Kucoin) GetWithdrawalsQuota(ctx context.Context, req *GetWithdrawalsQuotaRequest) (*GetWithdrawalsQuotaResponse, error) {
	url := k.url.JoinPath("/api/v1/withdrawals/quotas")

	q := url.Query()
	q.Add("currency", req.Currency)

	url.RawQuery = q.Encode()

	return NewRequest[*GetWithdrawalsQuotaResponse](ctx, url, http.MethodGet,
		WithHTTPClient(k.client),
		WithAuthz(k.authz))
}

type Option func(*Kucoin) Option

func NewClient(opts ...Option) (KucoinClient, error) {
	kucoin := &Kucoin{}

	url, err := url.Parse(config.Instance().RequiredString("KUCOIN_URL"))
	if err != nil {
		return nil, err
	}

	kucoin.url = url

	if kucoin.client == nil {
		kucoin.client = &http.Client{
			Timeout: time.Second * 30,
		}
	}

	authz, err := NewAuthz(WithAuthzHTTPClient(kucoin.client))
	if err != nil {
		return nil, err
	}

	kucoin.authz = authz

	return kucoin, nil
}
