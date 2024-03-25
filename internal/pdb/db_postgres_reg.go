package pdb

import (
	"context"
	"errors"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func (d *PGStorage) Register(login *string, password *string) (code int, err error) {
	code = http.StatusOK
	hashLogin := misc.Hash256(*login)
	rows, err := d.db.QueryContext(context.Background(), "SELECT 1 FROM UAUTH WHERE USER_NAME=$1;", hashLogin)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = errors.New("логин уже занят")
		logging.S().Infoln(*login, ":", err)
		code = http.StatusConflict
		return
	}

	var userID int64
	userID, err = d.nextOID()
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}

	hashPwd := misc.Hash256(*password)

	_, err = d.db.ExecContext(context.Background(),
		"INSERT INTO UAUTH (USERID,USER_NAME,USER_PWD,DELETE_FLAG) VALUES ($1,$2,$3,false);",
		userID, hashLogin, hashPwd)
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}

	return
}
func (d *PGStorage) Authent(login *string, password *string) (token string, code int, err error) {
	token = ""
	code = http.StatusOK
	hashLogin := misc.Hash256(*login)
	hashPwd := misc.Hash256(*password)

	rows, err := d.db.QueryContext(context.Background(),
		"SELECT USERID,USER_PWD FROM UAUTH WHERE USER_NAME=$1;",
		hashLogin)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}
	defer rows.Close()

	ok := false
	var userID int64
	var dtbPwd string
	if rows.Next() {
		err = rows.Scan(&userID, &dtbPwd)
		if err != nil {
			logging.S().Error(err)
			code = http.StatusInternalServerError
			return
		}
		ok = true
	}

	if ok {
		ok = hashPwd == dtbPwd
	}

	if ok {
		token, err = misc.MakeToken(userID)
		if err != nil {
			logging.S().Error(err)
			code = http.StatusInternalServerError
		}
	} else {
		err = errors.New("неверная пара логин/пароль")
		logging.S().Infoln(*login, *password, ":", err)
		code = http.StatusUnauthorized
	}
	return
}
