package chttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/cconfig"
)

type sdata struct {
	Login    *string `json:"login,omitempty"`
	Password *string `json:"password,omitempty"`
}

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

func Registration(login string, password string) error {

	var d sdata
	d.Login = &login
	d.Password = &password

	body, _ := json.Marshal(d)

	reader := bytes.NewReader(body)

	rsp, err := getResponce("POST", "register", reader)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func getResponce(method string, route string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, *cconfig.Get().ServerAddress+route, body)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
	defer cancel()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return resp, nil
}
