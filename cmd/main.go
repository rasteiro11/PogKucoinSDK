package main

import (
	"context"

	"github.com/rasteiro11/PogCore/pkg/logger"
	"github.com/rasteiro11/PogKucoinSDK/pkg/kucoin"
)

func main() {
	ctx := context.Background()

	kucoinClient, err := kucoin.NewClient()
	if err != nil {
		logger.Of(ctx).Debugf("ERR: %+v\n", err)
	}

	res, err := kucoinClient.GetWithdrawalsQuota(ctx, &kucoin.GetWithdrawalsQuotaRequest{
		Currency: "USDT",
	})
	if err != nil {
		logger.Of(ctx).Debugf("ERR: %+v\n", err)
	}

	logger.Of(ctx).Debugf("RES: %+v\n", res)

}
