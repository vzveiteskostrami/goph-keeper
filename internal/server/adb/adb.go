package adb

import (
	"context"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/server/pdb"
)

var lockWrite sync.Mutex

var store GSStorage

func init() {
	var s pdb.PGStorage
	store = &s
}

type GSStorage interface {
	Init()
	Close()
	UserIDExists(userID int64) (ok bool, err error)
	Register(login *string, password *string) (code int, err error)
	Authent(login *string, password *string, until time.Time) (token string, code int, err error)
	WriteUserDataList(ctx context.Context, userID int64, data []co.Udata) (newdata []co.Udata, err error)
	GetUserDataList(ctx context.Context, userID int64, info co.RequestList) (newdata []co.Udata, err error)
	DeleteDataList(ctx context.Context, data []co.Udata) (err error)
}

func Init() {
	store.Init()
}

func Close() {
	store.Close()
}

func UserIDExists(userID int64) (ok bool, err error) {
	ok, err = store.UserIDExists(userID)
	return
}

func Register(login *string, password *string) (code int, err error) {
	code, err = store.Register(login, password)
	return
}

func Authent(login *string, password *string, until time.Time) (token string, code int, err error) {
	token, code, err = store.Authent(login, password, until)
	return
}

func GetUserDataList(ctx context.Context, userID int64, info co.RequestList) (data []co.Udata, err error) {
	data, err = store.GetUserDataList(ctx, userID, info)
	return
}

func WriteUserDataList(ctx context.Context, userID int64, data []co.Udata) (newdata []co.Udata, err error) {
	newdata, err = store.WriteUserDataList(ctx, userID, data)
	return
}

func DeleteDataList(ctx context.Context, data []co.Udata) (err error) {
	err = store.DeleteDataList(ctx, data)
	return
}
