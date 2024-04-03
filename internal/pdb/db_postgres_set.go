package pdb

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/co"
)

func (d *PGStorage) WriteUserDataList(ctx context.Context, userID int64, data []co.Udata) (newdata []co.Udata, err error) {
	trn, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer trn.Rollback()

	updateTime := time.Now()
	oid := int64(0)
	var raw []byte
	for _, da := range data {
		raw, err = base64.StdEncoding.DecodeString(*da.Data)
		if err != nil {
			return
		}

		insert := false
		if da.Oid == nil || *da.Oid == 0 {
			oid, err = d.nextOID()
			if err != nil {
				return
			}
			k := oid
			da.Oid = &k
			insert = true
		}

		sql := ""
		if insert {
			sql = "INSERT INTO DATUM (OID,USERID,DATA_TYPE,DATA_NAME,CREATE_TIME,UPDATE_TIME,DELETE_FLAG,DATA)  VALUES ($1,$2,$3,$4,$5,$6,false,$7);"
		} else {
			sql = "UPDATE DATUM SET USERID=$2,DATA_TYPE=$3,DATA_NAME=$4,CREATE_TIME=$5,UPDATE_TIME=$6,DATA=$7 WHERE OID=$1;"
		}

		_, err = trn.ExecContext(ctx, sql,
			*da.Oid,
			userID,
			*da.DataType,
			*da.DataName,
			*da.CreateTime,
			updateTime,
			raw)
		if err != nil {
			return
		}

		nd := co.Udata{Oid: da.Oid, DataName: da.DataName, UpdateTime: &updateTime}
		newdata = append(newdata, nd)
	}

	trn.Commit()
	return
}
