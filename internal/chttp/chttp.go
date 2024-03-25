package chttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
