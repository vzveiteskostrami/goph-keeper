package pdb

import (
	"context"
	"errors"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
)

func (d *PGStorage) UserIDExists(userID int64) (ok bool, err error) {
	ok = false
	rows, err := d.db.QueryContext(context.Background(), "SELECT 1 FROM UDATA WHERE USERID=$1;", userID)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()
	ok = rows.Next()
	return
}

func (d *PGStorage) nextOID() (oid int64, err error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT NEXTVAL('GEN_OID');")
	if err == nil || rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&oid)
		if err != nil {
			logging.S().Error(err)
		}
	} else {
		err = errors.New("не вышло получить значение счётчика")
		logging.S().Error(err)
	}
	return
}
