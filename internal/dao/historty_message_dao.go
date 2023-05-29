package dao

import (
	"sync"
)

type HistoryMessageDao struct {
}

var (
	msgDao *HistoryMessageDao
	once   sync.Once
)

func GetHistoryMessageDao() *HistoryMessageDao {
	once.Do(func() {
		msgDao = &HistoryMessageDao{}
	})

	return msgDao
}

/*
func (d *HistoryMessageDao) QueryMessages(ctx context.Context, uid int64, sessionID string, page, size int) error {
	if ctx == nil {
		ctx = context.Background()
	}

	result := db.GetHBaseFromCtx(ctx).
		Table("message_history").
		Range().
		Options(hrpc.Filters(filter.NewColumnPaginationFilter())).
		Scan()
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}
*/
