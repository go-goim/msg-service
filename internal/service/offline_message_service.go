package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	redisv8 "github.com/go-redis/redis/v8"

	"github.com/go-goim/api/errors"
	messagev1 "github.com/go-goim/api/message/v1"
	"github.com/go-goim/core/pkg/consts"
	"github.com/go-goim/core/pkg/log"
	"github.com/go-goim/msg-service/internal/app"
)

type OfflineMessageService struct {
	messagev1.UnimplementedOfflineMessageServiceServer
}

func (o *OfflineMessageService) QueryOfflineMessage(ctx context.Context, req *messagev1.QueryOfflineMessageReq) (
	*messagev1.QueryOfflineMessageResp, error) {
	rsp := &messagev1.QueryOfflineMessageResp{
		Error: errors.ErrorOK(),
	}

	log.Info("req=", req.String())

	cnt, err := app.GetApplication().Redis.ZCount(ctx,
		consts.GetUserOfflineQueueKey(req.Uid),
		// offset add 1 to skip the message user last online msg
		strconv.FormatInt(req.GetLastMsgId()+1, 10),
		"+inf").Result()
	if err != nil {
		rsp.Error = errors.NewErrorWithError(err)
		return rsp, nil
	}

	rsp.Total = int32(cnt)
	if req.GetOnlyCount() {
		return rsp, nil
	}

	results, err := app.GetApplication().Redis.ZRangeByScoreWithScores(ctx,
		consts.GetUserOfflineQueueKey(req.Uid), &redisv8.ZRangeBy{
			// offset add 1 to skip the message user last online msg
			Min:    strconv.FormatInt(req.GetLastMsgId()+1, 10),
			Max:    "+inf",
			Offset: int64((req.GetPage() - 1) * req.GetPageSize()),
			Count:  int64(req.GetPageSize()),
		}).Result()
	if err != nil {
		rsp.Error = errors.NewErrorWithError(err)
		return rsp, nil
	}

	rsp.Messages = make([]*messagev1.Message, len(results))
	for i, result := range results {
		str := result.Member.(string)
		msg := new(messagev1.Message)
		if err = json.Unmarshal([]byte(str), msg); err != nil {
			rsp.Error = errors.NewErrorWithError(err)
			return rsp, nil
		}

		rsp.Messages[i] = msg
	}

	return rsp, nil
}

func (o *OfflineMessageService) ConfirmLastMstID(ctx context.Context, req *messagev1.ConfirmLastMsgIDReq) (
	*messagev1.ConfirmLastMsgIDResp, error) {
	rsp := &messagev1.ConfirmLastMsgIDResp{
		Error: errors.ErrorOK(),
	}

	log.Info("req=", req.String())
	id, err := zRemMsgAndGetLastOne(ctx, app.GetApplication().Redis,
		consts.GetUserOfflineQueueKey(req.Uid), strconv.FormatInt(req.GetLastMsgId(), 10))
	if err != nil {
		rsp.Error = errors.NewErrorWithError(err)
		return rsp, nil
	}

	rsp.LastMsgId = id
	return rsp, nil

}

func zRemMsgAndGetLastOne(ctx context.Context, rdb *redisv8.Client, key, max string) (int64, error) {
	results, err := rdb.TxPipelined(ctx, func(pp redisv8.Pipeliner) error {
		pp.ZRemRangeByScore(ctx,
			key,
			"-inf",
			max)
		pp.ZRangeWithScores(ctx, key, 0, 0)
		return nil
	})
	if err != nil {
		return 0, err
	}
	if len(results) != 2 {
		return 0, fmt.Errorf("invalid resp")
	}

	zsc, ok := results[1].(*redisv8.ZSliceCmd)
	if !ok {
		return 0, fmt.Errorf("invalid result")
	}

	zSliceResults, err := zsc.Result()
	if err != nil {
		return 0, err
	}

	if len(zSliceResults) == 0 {
		return 0, fmt.Errorf("empty result")
	}

	return int64(zSliceResults[0].Score), nil
}
