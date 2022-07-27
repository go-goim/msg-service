package service

import (
	"context"
	"encoding/json"
	"strconv"

	redisv8 "github.com/go-redis/redis/v8"

	responsepb "github.com/go-goim/api/transport/response"

	"github.com/go-goim/core/pkg/consts"
	"github.com/go-goim/core/pkg/log"

	messagev1 "github.com/go-goim/api/message/v1"

	"github.com/go-goim/msg-service/internal/app"
)

type OfflineMessageService struct {
	messagev1.UnimplementedOfflineMessageServiceServer
}

func (o *OfflineMessageService) QueryOfflineMessage(ctx context.Context, req *messagev1.QueryOfflineMessageReq) (
	*messagev1.QueryOfflineMessageResp, error) {
	rsp := &messagev1.QueryOfflineMessageResp{
		Response: responsepb.Code_OK.BaseResponse(),
	}

	log.Info("req=", req.String())

	cnt, err := app.GetApplication().Redis.ZCount(ctx,
		consts.GetUserOfflineQueueKey(req.Uid),
		// offset add 1 to skip the message user last online msg
		strconv.FormatInt(req.GetLastMsgId()+1, 10),
		"+inf").Result()
	if err != nil {
		rsp.Response = responsepb.NewBaseResponseWithError(err)
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
		rsp.Response = responsepb.NewBaseResponseWithError(err)
		return rsp, nil
	}

	rsp.Messages = make([]*messagev1.Message, len(results))
	for i, result := range results {
		str := result.Member.(string)
		msg := new(messagev1.Message)
		if err = json.Unmarshal([]byte(str), msg); err != nil {
			rsp.Response = responsepb.NewBaseResponseWithError(err)
			return rsp, nil
		}

		rsp.Messages[i] = msg
	}

	return rsp, nil
}
