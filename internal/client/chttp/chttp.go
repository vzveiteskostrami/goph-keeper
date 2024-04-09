package chttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vzveiteskostrami/goph-keeper/internal/co"
)

func CheckServerPresent() error {
	rsp, err := getResponce("GET", "ready", nil)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode == http.StatusOK {
		return nil
	} else {
		return fmt.Errorf("сервер вернул ошибку %d", rsp.StatusCode)
	}
}

func Registration(login string, password string) (int, error) {
	var d sdata
	d.Login = &login
	d.Password = &password

	body, _ := json.Marshal(d)

	reader := bytes.NewReader(body)

	rsp, err := getResponce("POST", "register", reader)
	if err != nil {
		return 0, err
	}
	defer rsp.Body.Close()

	err = nil
	if rsp.StatusCode != http.StatusOK {
		_, err = makeError(rsp)
	}
	return rsp.StatusCode, err
}

func Authorization(login string, password string, sessDuration int64) (int, error) {
	var d sdata
	d.Login = &login
	d.Password = &password
	d.SessionDuration = &sessDuration

	body, _ := json.Marshal(d)

	reader := bytes.NewReader(body)

	rsp, err := getResponce("POST", "login", reader)
	if err != nil {
		return 0, err
	}
	defer rsp.Body.Close()

	err = nil
	if rsp.StatusCode != http.StatusOK {
		_, err = makeError(rsp)
	}
	return rsp.StatusCode, err
}

/*
func Syncronize() (int, error) {
	body := []byte("")
	reader := bytes.NewReader(body)

	rsp, err := getResponce("POST", "sync", reader)
	if err != nil {
		return 0, err
	}
	defer rsp.Body.Close()

	err = nil
	if rsp.StatusCode != http.StatusOK {
		_, err = makeError(rsp)
	}
	return rsp.StatusCode, err
}
*/

func GetEntityList(lst co.RequestList) ([]co.Udata, int, error) {
	var da []co.Udata
	body, err := json.Marshal(lst)
	if err != nil {
		return da, 0, err
	}
	reader := bytes.NewReader(body)
	rsp, err := getResponce("POST", "list", reader)
	if err != nil {
		return da, 0, err
	}
	defer rsp.Body.Close()

	err = nil
	if rsp.StatusCode != http.StatusOK {
		_, err = makeError(rsp)
	} else {
		err = json.NewDecoder(rsp.Body).Decode(&da)
	}

	return da, rsp.StatusCode, err
}

func WriteEntityList(lst *[]co.Udata) ([]co.Udata, int, error) {
	var da []co.Udata
	body, err := json.Marshal(*lst)
	if err != nil {
		return da, 0, err
	}
	reader := bytes.NewReader(body)
	rsp, err := getResponce("POST", "wlist", reader)
	if err != nil {
		return da, 0, err
	}
	defer rsp.Body.Close()

	err = nil
	if rsp.StatusCode != http.StatusOK {
		_, err = makeError(rsp)
	} else {
		err = json.NewDecoder(rsp.Body).Decode(&da)
	}

	return da, rsp.StatusCode, err
}

func DeleteEntityList(lst *[]co.Udata) (int, error) {
	body, err := json.Marshal(*lst)
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(body)
	rsp, err := getResponce("POST", "del", reader)
	if err != nil {
		return 0, err
	}
	defer rsp.Body.Close()

	err = nil
	if rsp.StatusCode != http.StatusOK {
		_, err = makeError(rsp)
	}

	return rsp.StatusCode, err
}
