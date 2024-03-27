package routes

import (
	"io"
	"net/http"

	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
)

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

func Sessionf(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}
