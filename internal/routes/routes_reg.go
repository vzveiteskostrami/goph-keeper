package routes

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/adb"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func Registerf(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
			//Authf(w, r)
		}
	case <-r.Context().Done():
		logging.S().Infow("Регистрация прервана на клиентской стороне.")
		w.WriteHeader(http.StatusGone)
	}
}

func Authf(w http.ResponseWriter, r *http.Request) {
	regIn, err := misc.ExtractRegInfo(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if regIn.SessionDuration == nil {
		http.Error(w, "Недостаточно аргументов. Отсутствует продолжительность сессии.", http.StatusPreconditionFailed)
		return
	}

	/*
		t, err := time.Parse(time.RFC3339)
		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			return
		}
	*/

	completed := make(chan struct{})

	code := http.StatusOK
	token := ""

	go func() {
		token, code, err = adb.Authent(regIn.Login,
			regIn.Password,
			time.Now().Add(time.Duration(*regIn.SessionDuration)*time.Minute))
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
