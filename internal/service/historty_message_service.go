package service

import (
	"sync"

	messagev1 "github.com/go-goim/api/message/v1"
	"github.com/go-goim/msg-service/internal/dao"
)

type HistoryMessageService struct {
	messagev1.UnimplementedHistoryMessageServiceServer
	msgDao *dao.HistoryMessageDao
}

var (
	historyMessageService *HistoryMessageService
	once                  sync.Once
)

func GetHistoryMessageService() *HistoryMessageService {
	once.Do(func() {
		historyMessageService = &HistoryMessageService{
			msgDao: dao.GetHistoryMessageDao(),
		}
	})

	return historyMessageService
}

//
//func (s *HistoryMessageService) QuerySessionHistoryMessage(ctx context.Context, req *messagev1.QuerySessionHistoryMessageReq) (
//	*messagev1.QuerySessionHistoryMessageResp, error) {
//
//}
//
//func (s *HistoryMessageService) SyncHistoryMessage(ctx context.Context, req *messagev1.SyncHistoryMessageReq) (
//	*messagev1.SyncHistoryMessageResp, error) {
//
//}
