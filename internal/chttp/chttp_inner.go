package chttp

import (
	"context"
	"crypto/tls"
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

func makeError(rsp *http.Response) (int, error) {
	txt, _ := io.ReadAll(rsp.Body)
	err := fmt.Errorf(string(txt))
	return rsp.StatusCode, err
}
