package adb

import (
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/goph-keeper/internal/pdb"
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
	Authent(login *string, password *string) (token string, code int, err error)
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

func Authent(login *string, password *string) (token string, code int, err error) {
	token, code, err = store.Authent(login, password)
	return
}

type Order struct {
	oid        *int64
	userid     *int64
	Number     *string `json:"number,omitempty"`
	Status     *string `json:"status,omitempty"`
	status     *int16
	Accrual    *float32 `json:"accrual,omitempty"`
	UploadedAt *string  `json:"uploaded_at,omitempty"`
	uploadedAt *time.Time
	deleteFlag *bool
}

type Balance struct {
	Current   *float32 `json:"current,omitempty"`
	Withdrawn *float32 `json:"withdrawn,omitempty"`
}

type Withdraw struct {
	Order        *string  `json:"order,omitempty"`
	Sum          *float32 `json:"sum,omitempty"`
	ProcessedAt  *string  `json:"processed_at,omitempty"`
	withdrawDate *time.Time
}
