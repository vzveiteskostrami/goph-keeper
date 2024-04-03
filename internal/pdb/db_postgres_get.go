package pdb

import (
	"context"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
)

func (d *PGStorage) GetUserDataList(ctx context.Context, userID int64, info co.RequestList) (data []co.Udata, err error) {
	sql := "SELECT OID,DATA_TYPE,DATA_NAME,CREATE_TIME,UPDATE_TIME"
	if *info.Full {
		sql += ",DATA"
	}
	sql += " FROM DATUM WHERE USERID=$1"
	if !*info.All {
		if len(info.Data) > 0 {
			s := " AND OID IN ("
			for _, a := range info.Data {
				s += strconv.FormatInt(*a.Oid, 10) + ","
			}
			s = s[:len(s)-1]
			s += ")"
			sql += s
		}
	}
	sql += ";"

	rows, err := d.db.QueryContext(ctx, sql, userID)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		oid := int64(0)
		dataType := int16(0)
		dataName := ""
		createTime := time.Now()
		updateTime := time.Now()
		var da co.Udata
		if *info.Full {
			var arr []byte
			err = rows.Scan(&oid, &dataType, &dataName, &createTime, &updateTime, &arr)
			zs := base64.StdEncoding.EncodeToString(arr)
			da = co.Udata{Oid: &oid, DataType: &dataType, DataName: &dataName, CreateTime: &createTime, UpdateTime: &updateTime, Data: &zs}
		} else {
			err = rows.Scan(&oid, &dataType, &dataName, &createTime, &updateTime)
			da = co.Udata{Oid: &oid, DataType: &dataType, DataName: &dataName, CreateTime: &createTime, UpdateTime: &updateTime}
		}
		data = append(data, da)
		if err != nil {
			logging.S().Error(err)
			return
		}
	}
	return
}
