package routes

import (
	"bytes"
	"io"
	"net/http"

	"github.com/vzveiteskostrami/goph-keeper/internal/adb"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func Registerf(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	regIn, err := misc.ExtractRegInfo(io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed := make(chan struct{})

	code := http.StatusOK

	go func() {
		code, err = adb.Register(regIn.Login, regIn.Password)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			Authf(w, r)
		}
	case <-r.Context().Done():
		logging.S().Infow("Регистрация прервана на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}

func Authf(w http.ResponseWriter, r *http.Request) {
	regIn, err := misc.ExtractRegInfo(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed := make(chan struct{})

	code := http.StatusOK
	token := ""

	go func() {
		token, code, err = adb.Authent(regIn.Login, regIn.Password)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.SetCookie(w, &http.Cookie{Name: "token", Value: token, HttpOnly: true})
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(code)
			w.Write([]byte(token))
		}
	case <-r.Context().Done():
		logging.S().Infow("Аутентификация прервана на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}

func Echof(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed := make(chan struct{})

	go func() {
		w.Write(body)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
	case <-r.Context().Done():
		logging.S().Infow("Эхо прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}

func Readyf(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}
