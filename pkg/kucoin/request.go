package kucoin

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rasteiro11/PogCore/pkg/logger"
)

type requestConfig struct {
	payload any
	header  http.Header
	client  *http.Client
	authz   *Authz
}

type RequestOption func(*requestConfig)

func WithHTTPClient(client *http.Client) RequestOption {
	return func(rc *requestConfig) {
		rc.client = client
	}
}

func WithAuthz(authz *Authz) RequestOption {
	return func(rc *requestConfig) {
		rc.authz = authz
	}
}

func WithPayload(payload any) RequestOption {
	return func(o *requestConfig) {
		o.payload = payload
	}
}

func WithToken(token string) RequestOption {
	return func(o *requestConfig) {
		o.header.Add("Authorization", "Bearer "+token)
	}
}

func WithHeader(key string, value string) RequestOption {
	return func(rc *requestConfig) {
		rc.header.Add(key, value)
	}
}

func newRequestOption(opts ...RequestOption) *requestConfig {
	options := &requestConfig{
		header: make(http.Header),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
func passPhraseEncrypt(key, plain []byte) string {
	hm := hmac.New(sha256.New, key)
	hm.Write(plain)
	return base64.StdEncoding.EncodeToString(hm.Sum(nil))
}

func (r *requestConfig) addHeaders(req *http.Request) {
	timestamp := time.Now().UnixNano() / 1000000
	r.authz.GetSignature(req.Context(), getSignatureRequest{
		timestamp: timestamp,
		method:    req.Method,
		endpoint:  req.URL.Path + req.URL.RawQuery,
	})

	messageToSign := fmt.Sprintf("%d%s%s", timestamp, req.Method, "/api/v1/withdrawals/quotas?currency=USDT")

	req.Header.Add("KC-API-SIGN", passPhraseEncrypt([]byte(r.authz.secret), []byte(messageToSign)))
	req.Header.Add("KC-API-PASSPHRASE", passPhraseEncrypt([]byte(r.authz.secret), []byte(r.authz.passphrase)))
	req.Header.Add("KC-API-KEY", r.authz.key)
	req.Header.Add("KC-API-TIMESTAMP", fmt.Sprintf("%d", timestamp))
	req.Header.Add("KC-API-KEY-VERSION", "2")
}

func NewRequest[T any](ctx context.Context, url *url.URL, method string, opts ...RequestOption) (T, error) {
	var target T

	options := newRequestOption(opts...)

	var (
		token string
		err   error
	)

	var payload io.Reader
	if options.payload != nil && options.header.Get("Content-Type") == "" {
		if body, err := json.Marshal(options.payload); err == nil {
			payload = bytes.NewBuffer(body)
			logger.Of(ctx).Debugf("package=kucoin body=%s\n", string(body))
		}
	} else {
		if options.payload != nil {
			if r, ok := options.payload.(io.Reader); ok {
				payload = r
			}
		}
	}

	logger.Of(ctx).Debugf("package=kucoin method=%s url=%s", method, url)

	req, err := http.NewRequestWithContext(ctx, method, url.String(), payload)
	if err != nil {
		logger.Of(ctx).Errorf("package=kucoin call=http.NewRequestWithContext() error=%+v\n", err)
		return target, err
	}

	req.Header = options.header

	if v := req.Header.Get("Content-Type"); v == "" {
		req.Header.Add("Content-Type", "application/json")
	}

	req.Header.Add("api-version", "1.0")

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	options.addHeaders(req)

	for key, val := range req.Header {
		logger.Of(ctx).Debugf("KEY: %s VAL: %s\n", key, val[0])
	}

	res, err := options.client.Do(req)
	if err != nil {
		logger.Of(ctx).Errorf("package=kucoin call=http.DefaultClient.Do() error=%+v\n", err)
		return target, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Of(ctx).Errorf("package=kucoin call=http.DefaultClient.Do() error=%+v\n", err)
		return target, err
	}
	logger.Of(ctx).Debugf("RES: %+v\n", string(body))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return target, &Error{
			Code: res.StatusCode,
			Body: string(body),
		}
	}

	if res.ContentLength != 0 {
		if err := json.Unmarshal(body, &target); err != nil {
			logger.Of(ctx).Errorf("package=kucoin call=json.Unmarshal() error=%+v\n", err)
			return target, err
		}
	}

	return target, nil
}
