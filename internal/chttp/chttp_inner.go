package chttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/cconfig"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

type sdata struct {
	Login           *string `json:"login,omitempty"`
	Password        *string `json:"password,omitempty"`
	SessionDuration *int64  `json:"session_duration,omitempty"`
}

func getResponce(method string, route string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, *cconfig.Get().ServerAddress+route, body)
	if err != nil {
		return nil, err
	}

	//ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Minute)
	defer cancel()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	if ok, _ := misc.FileExists("ADM\\token"); ok {
		key, _ := misc.UnicKeyForExeDir()
		b, _, err := misc.ReadFromFileProtectedZIP("ADM\\token", key)
		if err == nil {
			cookie := &http.Cookie{
				Name:  "token",
				Value: string(b),
			}
			req.AddCookie(cookie)
		}
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "token" {
			key, _ := misc.UnicKeyForExeDir()
			misc.SaveToFileProtectedZIP("ADM\\token", "token", key, []byte(cookie.Value))
		}
	}

	return resp, nil
}

func makeError(rsp *http.Response) (int, error) {
	txt, _ := io.ReadAll(rsp.Body)
	err := fmt.Errorf(string(txt))
	return rsp.StatusCode, err
}
