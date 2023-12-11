package kucoin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/rasteiro11/PogCore/pkg/config"
	"github.com/rasteiro11/PogCore/pkg/logger"
)

type Authz struct {
	client     *http.Client
	key        string
	secret     string
	passphrase string
}

type AuthzOption func(*Authz)

func WithAuthzHTTPClient(client *http.Client) AuthzOption {
	return func(a *Authz) {
		a.client = client
	}
}

func WithKey(key string) AuthzOption {
	return func(a *Authz) {
		a.key = key
	}
}

func WithPassphrase(passphrase string) AuthzOption {
	return func(a *Authz) {
		a.passphrase = passphrase
	}
}

func WithSecret(secret string) AuthzOption {
	return func(a *Authz) {
		a.secret = secret
	}
}

type getSignatureRequest struct {
	timestamp int64
	method    string
	endpoint  string
}

type getSignatureResponse struct {
	msgToSign []byte
}

func (a *Authz) GetSignature(ctx context.Context, req getSignatureRequest) getSignatureResponse {
	messageToSign := fmt.Sprintf("%d%s%s", req.timestamp, req.method, req.endpoint)

	logger.Of(ctx).Debugf("ENDPOINT: %s\n", req.endpoint)
	logger.Of(ctx).Debugf("MESSAGE TO SIGN: %s\n", messageToSign)

	logger.Of(ctx).Debugf("SECRET: %s", a.secret)
	logger.Of(ctx).Debugf("PASS: %s", a.passphrase)

	hm := hmac.New(sha256.New, []byte(a.secret))
	hm.Write([]byte(messageToSign))

	return getSignatureResponse{
		msgToSign: hm.Sum(nil),
	}
}

func NewAuthz(opts ...AuthzOption) (*Authz, error) {
	authz := &Authz{
		key:        config.Instance().String("KUCOIN_KEY"),
		secret:     config.Instance().String("KUCOIN_SECRET"),
		passphrase: config.Instance().String("KUCOIN_PASSPHRASE"),
	}

	for _, opt := range opts {
		opt(authz)
	}

	if authz.client == nil {
		authz.client = &http.Client{
			Timeout: time.Second * 30,
		}
	}

	return authz, nil
}
